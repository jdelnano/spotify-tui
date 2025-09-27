package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

type playlistsLoadedMsg struct {
	playlists []spotifyPkg.SimplePlaylist
}

type tracksLoadedMsg struct {
	tracks []spotifyPkg.PlaylistTrack
}

type searchResultsMsg struct {
	results []spotifyPkg.FullTrack
}

type currentlyPlayingMsg struct {
	playing *spotifyPkg.CurrentlyPlaying
}

type errMsg struct {
	err error
}

type statusMsg struct {
	message string
}

type libraryLoadedMsg struct {
	categories []LibraryCategory
}

type savedTracksLoadedMsg struct {
	tracks []spotifyPkg.SavedTrack
}

type recentlyPlayedLoadedMsg struct {
	items []spotifyPkg.RecentlyPlayedItem
}

type moreTracksLoadedMsg struct {
	tracks []spotifyPkg.SavedTrack
}

type savedAlbumsLoadedMsg struct {
	albums []spotifyPkg.SavedAlbum
}

type moreAlbumsLoadedMsg struct {
	albums []spotifyPkg.SavedAlbum
}

type albumTracksLoadedMsg struct {
	tracks []spotifyPkg.SimpleTrack
	album  *spotifyPkg.SavedAlbum
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPlaylists(),
		m.loadLibrary(),
		m.fetchCurrentlyPlaying(),
		tea.Every(time.Second*1, func(t time.Time) tea.Msg {
			return t
		}),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isSearching {
			return m.handleSearchInput(msg)
		}
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case playlistsLoadedMsg:
		m.playlists = msg.playlists
		m.statusMessage = fmt.Sprintf("Loaded %d playlists", len(msg.playlists))
		return m, nil

	case tracksLoadedMsg:
		m.playlistTracks = msg.tracks
		m.trackCursor = 0
		m.statusMessage = fmt.Sprintf("Loaded %d tracks", len(msg.tracks))
		return m, nil

	case searchResultsMsg:
		m.searchResults = msg.results
		m.searchCursor = 0
		m.statusMessage = fmt.Sprintf("Found %d results", len(msg.results))
		return m, nil

	case currentlyPlayingMsg:
		m.currentlyPlaying = msg.playing
		return m, nil

	case time.Time:
		// Increment progress locally for smooth updates
		if m.currentlyPlaying.Playing {
			// Add 1 second to progress
			m.currentlyPlaying.Progress += 1000

			// Don't let it exceed the track duration
			if m.currentlyPlaying.Progress > m.currentlyPlaying.Item.Duration {
				m.currentlyPlaying.Progress = m.currentlyPlaying.Item.Duration
			}
		}
		// Always fetch currently playing info for live updates
		return m, m.fetchCurrentlyPlaying()

	case errMsg:
		m.err = msg.err
		m.statusMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case statusMsg:
		m.statusMessage = msg.message
		return m, nil

	case libraryLoadedMsg:
		m.libraryCategories = msg.categories
		return m, nil

	case savedTracksLoadedMsg:
		m.savedTracks = msg.tracks
		m.trackCursor = 0
		m.statusMessage = fmt.Sprintf("Loaded %d liked songs", len(msg.tracks))
		return m, nil

	case recentlyPlayedLoadedMsg:
		// Convert recently played items to saved tracks for display
		m.savedTracks = nil
		for _, item := range msg.items {
			m.savedTracks = append(m.savedTracks, spotifyPkg.SavedTrack{
				FullTrack: spotifyPkg.FullTrack{
					SimpleTrack: item.Track,
				},
			})
		}
		m.trackCursor = 0
		m.statusMessage = fmt.Sprintf("Loaded %d recently played tracks", len(msg.items))
		return m, nil

	case moreTracksLoadedMsg:
		m.savedTracks = append(m.savedTracks, msg.tracks...)
		m.savedTracksOffset += len(msg.tracks)
		m.isLoadingMore = false
		m.statusMessage = fmt.Sprintf("Loaded %d/%d tracks", len(m.savedTracks), m.savedTracksTotal)
		return m, nil

	case savedAlbumsLoadedMsg:
		m.savedAlbums = msg.albums
		m.albumCursor = 0
		m.statusMessage = fmt.Sprintf("Loaded %d albums", len(msg.albums))
		return m, nil

	case moreAlbumsLoadedMsg:
		m.savedAlbums = append(m.savedAlbums, msg.albums...)
		m.savedAlbumsOffset += len(msg.albums)
		m.isLoadingMoreAlbums = false
		m.statusMessage = fmt.Sprintf("Loaded %d/%d albums", len(m.savedAlbums), m.savedAlbumsTotal)
		return m, nil

	case albumTracksLoadedMsg:
		m.albumTracks = msg.tracks
		m.selectedAlbum = msg.album
		m.isViewingAlbum = true
		m.albumTrackCursor = 0
		m.statusMessage = fmt.Sprintf("Loaded %d tracks from %s", len(msg.tracks), msg.album.Name)
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "/":
		m.viewMode = SearchView
		m.isSearching = true
		m.searchQuery = ""
		return m, nil

	case "p":
		m.viewMode = PlaylistView
	case "tab":
		if m.viewMode == PlaylistView {
			if m.selectedPlaylist == nil {
				m.selectedPlaylist = nil
				m.selectedLibraryItem = nil
			}
		}
		return m, nil

	case "up", "k":
		switch m.viewMode {
		case PlaylistView:
			if m.selectedPlaylist == nil && m.selectedLibraryItem == nil {
				if m.playlistCursor == 0 && len(m.libraryCategories) > 0 {
					m.libraryCursor = len(m.libraryCategories) - 1
					m.playlistCursor = -1
				} else if m.playlistCursor > 0 {
					m.playlistCursor--
				} else if m.libraryCursor > 0 {
					m.libraryCursor--
				}
			} else if m.selectedPlaylist != nil || m.selectedLibraryItem != nil {
				if m.selectedLibraryItem != nil && m.selectedLibraryItem.Type == "saved_albums" {
					if m.isViewingAlbum && m.albumTrackCursor > 0 {
						m.albumTrackCursor--
					} else if !m.isViewingAlbum && m.albumCursor > 0 {
						m.albumCursor--
					}
				} else if m.trackCursor > 0 {
					m.trackCursor--
				}
			}
		case SearchView:
			if m.searchCursor > 0 {
				m.searchCursor--
			}
		}
		return m, nil

	case "down", "j":
		switch m.viewMode {
		case PlaylistView:
			if m.selectedPlaylist == nil && m.selectedLibraryItem == nil {
				if m.playlistCursor == -1 && len(m.playlists) > 0 {
					m.playlistCursor = 0
					m.libraryCursor = -1
				} else if m.playlistCursor >= 0 && m.playlistCursor < len(m.playlists)-1 {
					m.playlistCursor++
				} else if m.playlistCursor == -1 && m.libraryCursor < len(m.libraryCategories)-1 {
					m.libraryCursor++
				} else if m.libraryCursor >= 0 && m.libraryCursor == len(m.libraryCategories)-1 && len(m.playlists) > 0 {
					m.playlistCursor = 0
					m.libraryCursor = -1
				}
			} else if m.selectedLibraryItem != nil {
				if (m.selectedLibraryItem.Type == "saved_tracks" || m.selectedLibraryItem.Type == "recently_played") && m.trackCursor < len(m.savedTracks)-1 {
					m.trackCursor++
					// Load more tracks if we're near the end and it's saved_tracks
					if m.selectedLibraryItem.Type == "saved_tracks" &&
						m.trackCursor >= len(m.savedTracks)-10 &&
						!m.isLoadingMore &&
						m.savedTracksOffset < m.savedTracksTotal {
						return m, m.loadMoreSavedTracks()
					}
				} else if m.selectedLibraryItem.Type == "saved_albums" {
					if m.isViewingAlbum && m.albumTrackCursor < len(m.albumTracks)-1 {
						m.albumTrackCursor++
					} else if !m.isViewingAlbum && m.albumCursor < len(m.savedAlbums)-1 {
						m.albumCursor++
						// Load more albums if we're near the end
						if m.albumCursor >= len(m.savedAlbums)-5 &&
							!m.isLoadingMoreAlbums &&
							m.savedAlbumsOffset < m.savedAlbumsTotal {
							return m, m.loadMoreSavedAlbums()
						}
					}
				} else if m.selectedLibraryItem.Type == "playlist" && m.trackCursor < len(m.playlistTracks)-1 {
					m.trackCursor++
				}
			} else if m.selectedPlaylist != nil && m.trackCursor < len(m.playlistTracks)-1 {
				m.trackCursor++
			}
		case SearchView:
			if m.searchCursor < len(m.searchResults)-1 {
				m.searchCursor++
			}
		}
		return m, nil

	case "left", "h":
		if m.viewMode == PlaylistView {
			if m.selectedPlaylist != nil {
				m.selectedPlaylist = nil
				m.playlistTracks = nil
				m.trackCursor = 0
			} else if m.selectedLibraryItem != nil {
				// If viewing album tracks, go back to album list
				if m.isViewingAlbum && m.selectedLibraryItem.Type == "saved_albums" {
					m.isViewingAlbum = false
					m.selectedAlbum = nil
					m.albumTracks = nil
					m.albumTrackCursor = 0
				} else {
					// Reset pagination state when leaving library items
					m.selectedLibraryItem = nil
					m.savedTracks = nil
					m.savedAlbums = nil
					m.playlistTracks = nil
					m.trackCursor = 0
					m.albumCursor = 0
					m.savedTracksOffset = 0
					m.savedTracksTotal = 0
					m.savedAlbumsOffset = 0
					m.savedAlbumsTotal = 0
					m.hasLoadedInitial = false
					m.isLoadingMore = false
					m.isLoadingMoreAlbums = false
					m.isViewingAlbum = false
					m.selectedAlbum = nil
					m.albumTracks = nil
					m.albumTrackCursor = 0
				}
			}
		}
		return m, nil

	case "right", "l", "enter":
		switch m.viewMode {
		case PlaylistView:
			if m.selectedPlaylist == nil && m.selectedLibraryItem == nil {
				if m.libraryCursor >= 0 && m.libraryCursor < len(m.libraryCategories) {
					m.selectedLibraryItem = &m.libraryCategories[m.libraryCursor]
					return m, m.loadLibraryItems(m.libraryCategories[m.libraryCursor])
				} else if m.playlistCursor >= 0 && m.playlistCursor < len(m.playlists) {
					m.selectedPlaylist = &m.playlists[m.playlistCursor]
					return m, m.loadPlaylistTracks(m.playlists[m.playlistCursor].ID)
				}
			} else if m.selectedLibraryItem != nil {
				if (m.selectedLibraryItem.Type == "saved_tracks" || m.selectedLibraryItem.Type == "recently_played") && len(m.savedTracks) > 0 {
					// Add current track to recently played before playing new one
					if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
						m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
							Track: m.currentlyPlaying.Item.SimpleTrack,
						})
					}
					return m, m.playSavedTrack(m.savedTracks[m.trackCursor])
				} else if m.selectedLibraryItem.Type == "saved_albums" {
					if m.isViewingAlbum && len(m.albumTracks) > 0 {
						// Play a track from the album
						return m, m.playAlbumTrack(m.albumTracks[m.albumTrackCursor])
					} else if len(m.savedAlbums) > 0 {
						// Select an album to view its tracks
						return m, m.loadAlbumTracks(m.savedAlbums[m.albumCursor])
					}
				} else if m.selectedLibraryItem.Type == "playlist" && len(m.playlistTracks) > 0 {
					// Add current track to recently played before playing new one
					if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
						m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
							Track: m.currentlyPlaying.Item.SimpleTrack,
						})
					}
					return m, m.playTrack(m.playlistTracks[m.trackCursor])
				}
			} else if m.selectedPlaylist != nil && len(m.playlistTracks) > 0 {
				// Add current track to recently played before playing new one
				if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
					m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
						Track: m.currentlyPlaying.Item.SimpleTrack,
					})
				}
				return m, m.playTrack(m.playlistTracks[m.trackCursor])
			}
		case SearchView:
			if len(m.searchResults) > 0 {
				// Add current track to recently played before playing new one
				if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
					m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
						Track: m.currentlyPlaying.Item.SimpleTrack,
					})
				}
				return m, m.playSearchResult(m.searchResults[m.searchCursor])
			}
		}
		return m, nil

	case " ":
		return m, m.togglePlayback()

	case ">":
		return m, m.nextTrack()

	case "<":
		// Check to see if we have a previous track to fall back on, and if not just
		// return nil and don't disrupt the listening experience
		if len(m.recentlyPlayed) == 0 {
			return m, nil
		}

		track := m.recentlyPlayed[len(m.recentlyPlayed)-1]
		// "pop" last element from slice
		m.recentlyPlayed = m.recentlyPlayed[:len(m.recentlyPlayed)-1]
		newCurrentTrack := spotifyPkg.PlaylistTrack{
			Track: spotifyPkg.FullTrack{
				SimpleTrack: track.Track,
			},
		}
		return m, m.playTrack(newCurrentTrack)

	}

	return m, nil
}

func (m *Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.isSearching = false
		if m.searchQuery != "" {
			return m, m.searchTracks()
		}
		return m, nil

	case "esc":
		m.isSearching = false
		m.viewMode = PlaylistView
		return m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil

	case " ":
		m.searchQuery += " "
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
		}
	}

	return m, nil
}

func (m *Model) loadPlaylists() tea.Cmd {
	return func() tea.Msg {
		playlists, err := m.spotifyClient.GetPlaylists()
		if err != nil {
			return errMsg{err}
		}
		return playlistsLoadedMsg{playlists}
	}
}

func (m *Model) loadPlaylistTracks(playlistID spotifyPkg.ID) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.spotifyClient.GetPlaylistTracks(playlistID)
		if err != nil {
			return errMsg{err}
		}
		return tracksLoadedMsg{tracks}
	}
}

func (m *Model) searchTracks() tea.Cmd {
	return func() tea.Msg {
		results, err := m.spotifyClient.SearchTracks(m.searchQuery)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{results}
	}
}

func (m *Model) playTrack(track spotifyPkg.PlaylistTrack) tea.Cmd {
	return func() tea.Msg {
		activeDevice, err := m.spotifyClient.GetActiveDevice()
		if err != nil {
			return errMsg{err}
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.Track.URI)
		if err != nil {
			return errMsg{err}
		}

		// Update the currently playing state immediately
		m.currentlyPlaying = &spotifyPkg.CurrentlyPlaying{
			Playing:  true,
			Item:     &track.Track,
			Progress: 0,
		}

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Track.Name, track.Track.Artists[0].Name)}
	}
}

func (m *Model) playSearchResult(track spotifyPkg.FullTrack) tea.Cmd {
	return func() tea.Msg {
		activeDevice, err := m.spotifyClient.GetActiveDevice()
		if err != nil {
			return errMsg{err}
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.URI)
		if err != nil {
			return errMsg{err}
		}
		// Update the currently playing state immediately
		m.currentlyPlaying = &spotifyPkg.CurrentlyPlaying{
			Playing:  true,
			Item:     &track,
			Progress: 0,
		}

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Name, track.Artists[0].Name)}
	}
}

func (m *Model) togglePlayback() tea.Cmd {
	return func() tea.Msg {
		// First check if we have any active devices
		devices, err := m.spotifyClient.GetDevices()
		if err != nil {
			return errMsg{err}
		}

		if len(devices) == 0 {
			return statusMsg{"No active devices. Open Spotify on a device first."}
		}

		activeDevice := spotifyPkg.PlayerDevice{
			ID: "",
		}

		// Find an active device
		for _, device := range devices {
			if device.Active {
				activeDevice = device
				break
			}
		}

		if activeDevice.ID == "" {
			return statusMsg{"No active playback. Start playing a track first."}
		}

		opts := &spotifyPkg.PlayOptions{
			DeviceID: &activeDevice.ID,
		}

		if m.currentlyPlaying != nil && m.currentlyPlaying.Playing {
			err := m.spotifyClient.Pause(opts)
			if err != nil {
				return errMsg{err}
			}
			// Change the model state to reflect we just paused the song
			m.currentlyPlaying.Playing = false

			if m.currentlyPlaying.Item != nil {
				artists := ""
				if len(m.currentlyPlaying.Item.Artists) > 0 {
					artists = m.currentlyPlaying.Item.Artists[0].Name
				}
				return statusMsg{fmt.Sprintf("Paused: %s - %s", m.currentlyPlaying.Item.Name, artists)}
			}
			return statusMsg{"Paused"}
		} else {
			err := m.spotifyClient.Resume(opts)
			if err != nil {
				return errMsg{err}
			}
			// Change the model state to reflect we just resumed the song
			if m.currentlyPlaying != nil {
				m.currentlyPlaying.Playing = true
				if m.currentlyPlaying.Item != nil {
					artists := ""
					if len(m.currentlyPlaying.Item.Artists) > 0 {
						artists = m.currentlyPlaying.Item.Artists[0].Name
					}
					return statusMsg{fmt.Sprintf("Playing: %s - %s", m.currentlyPlaying.Item.Name, artists)}
				}
			}
			return statusMsg{"Resumed"}
		}
	}
}

func (m *Model) nextTrack() tea.Cmd {
	return func() tea.Msg {
		return statusMsg{"Next track is not yet implemented"}
	}
}

func (m *Model) fetchCurrentlyPlaying() tea.Cmd {
	return func() tea.Msg {
		playing, err := m.spotifyClient.GetCurrentlyPlaying()
		if err != nil {
			// Don't show error, just return nil (no track playing)
			return currentlyPlayingMsg{nil}
		}
		return currentlyPlayingMsg{playing}
	}
}

func (m *Model) loadLibrary() tea.Cmd {
	return func() tea.Msg {
		var categories []LibraryCategory

		// Fetch Liked Songs count (just the count, not all tracks)
		savedTracksCount, err := m.spotifyClient.GetSavedTracksCount()
		if err == nil && savedTracksCount > 0 {
			categories = append(categories, LibraryCategory{
				Name:      "Liked Songs",
				Icon:      "❤️",
				Type:      "saved_tracks",
				ItemCount: savedTracksCount,
			})
		}

		// Fetch all user playlists to find special ones
		playlists, _ := m.spotifyClient.GetPlaylists()

		// Look for Release Radar
		for _, playlist := range playlists {
			if playlist.Name == "Release Radar" {
				trackCount := int(playlist.Tracks.Total)
				categories = append(categories, LibraryCategory{
					Name:       "Release Radar",
					Icon:       "📡",
					Type:       "playlist",
					ItemCount:  trackCount,
					PlaylistID: playlist.ID,
				})
				break
			}
		}

		// Look for Discover Weekly
		for _, playlist := range playlists {
			if playlist.Name == "Discover Weekly" {
				trackCount := int(playlist.Tracks.Total)
				categories = append(categories, LibraryCategory{
					Name:       "Discover Weekly",
					Icon:       "🎲",
					Type:       "playlist",
					ItemCount:  trackCount,
					PlaylistID: playlist.ID,
				})
				break
			}
		}

		// Look for Daily Mix playlists
		for _, playlist := range playlists {
			if len(playlist.Name) > 9 && playlist.Name[:9] == "Daily Mix" {
				trackCount := int(playlist.Tracks.Total)
				categories = append(categories, LibraryCategory{
					Name:       playlist.Name,
					Icon:       "🎶",
					Type:       "playlist",
					ItemCount:  trackCount,
					PlaylistID: playlist.ID,
				})
			}
		}

		// Add Recently Played category
		categories = append(categories, LibraryCategory{
			Name: "Recently Played",
			Icon: "🕐",
			Type: "recently_played",
		})

		// Add Saved Albums
		savedAlbumsCount, err := m.spotifyClient.GetSavedAlbumsCount()
		if err == nil && savedAlbumsCount > 0 {
			categories = append(categories, LibraryCategory{
				Name:      "Saved Albums",
				Icon:      "💿",
				Type:      "saved_albums",
				ItemCount: savedAlbumsCount,
			})
		}

		// Add Followed Artists
		followedArtists, err := m.spotifyClient.GetFollowedArtists()
		if err == nil && len(followedArtists) > 0 {
			categories = append(categories, LibraryCategory{
				Name:      "Followed Artists",
				Icon:      "🎤",
				Type:      "followed_artists",
				ItemCount: len(followedArtists),
			})
		}

		return libraryLoadedMsg{categories: categories}
	}
}

func (m *Model) loadLibraryItems(category LibraryCategory) tea.Cmd {
	return func() tea.Msg {
		switch category.Type {
		case "saved_tracks":
			// Load first page of saved tracks (50 tracks)
			tracks, total, err := m.spotifyClient.GetSavedTracksPage(50, 0)
			if err != nil {
				return errMsg{err}
			}
			// Store the total count for pagination
			m.savedTracksTotal = total
			m.savedTracksOffset = len(tracks)
			m.hasLoadedInitial = true
			return savedTracksLoadedMsg{tracks}
		case "playlist":
			tracks, err := m.spotifyClient.GetPlaylistTracks(category.PlaylistID)
			if err != nil {
				return errMsg{err}
			}
			return tracksLoadedMsg{tracks}
		case "recently_played":
			items, err := m.spotifyClient.GetRecentlyPlayed()
			if err != nil {
				return errMsg{err}
			}
			return recentlyPlayedLoadedMsg{items}
		case "saved_albums":
			// Load first page of saved albums (20 albums)
			albums, total, err := m.spotifyClient.GetSavedAlbumsPage(20, 0)
			if err != nil {
				return errMsg{err}
			}
			// Store the total count for pagination
			m.savedAlbumsTotal = total
			m.savedAlbumsOffset = len(albums)
			return savedAlbumsLoadedMsg{albums}
		default:
			return statusMsg{"This library category is not yet implemented"}
		}
	}
}

func (m *Model) loadMoreSavedTracks() tea.Cmd {
	return func() tea.Msg {
		if m.isLoadingMore || m.savedTracksOffset >= m.savedTracksTotal {
			return nil
		}

		m.isLoadingMore = true
		tracks, _, err := m.spotifyClient.GetSavedTracksPage(50, m.savedTracksOffset)
		if err != nil {
			return errMsg{err}
		}
		return moreTracksLoadedMsg{tracks}
	}
}

func (m *Model) loadMoreSavedAlbums() tea.Cmd {
	return func() tea.Msg {
		if m.isLoadingMoreAlbums || m.savedAlbumsOffset >= m.savedAlbumsTotal {
			return nil
		}

		m.isLoadingMoreAlbums = true
		albums, _, err := m.spotifyClient.GetSavedAlbumsPage(20, m.savedAlbumsOffset)
		if err != nil {
			return errMsg{err}
		}
		return moreAlbumsLoadedMsg{albums}
	}
}

func (m *Model) loadAlbumTracks(album spotifyPkg.SavedAlbum) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.spotifyClient.GetAlbumTracks(album.ID)
		if err != nil {
			return errMsg{err}
		}
		return albumTracksLoadedMsg{tracks: tracks, album: &album}
	}
}

func (m *Model) playAlbumTrack(track spotifyPkg.SimpleTrack) tea.Cmd {
	return func() tea.Msg {
		activeDevice, err := m.spotifyClient.GetActiveDevice()
		if err != nil {
			return errMsg{err}
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.URI)
		if err != nil {
			return errMsg{err}
		}

		// Get artist names from the album
		artists := ""
		if m.selectedAlbum != nil {
			for i, artist := range m.selectedAlbum.Artists {
				if i > 0 {
					artists += ", "
				}
				artists += artist.Name
			}
		}

		// Add to recently played before playing new track
		if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
			m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
				Track: m.currentlyPlaying.Item.SimpleTrack,
			})
		}

		// Update currently playing with the album track
		// Note: We're creating a minimal FullTrack from SimpleTrack for tracking purposes
		m.currentlyPlaying = &spotifyPkg.CurrentlyPlaying{
			Playing: true,
			Item: &spotifyPkg.FullTrack{
				SimpleTrack: track,
			},
		}

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Name, artists)}
	}
}

func (m *Model) playSavedTrack(track spotifyPkg.SavedTrack) tea.Cmd {
	return func() tea.Msg {
		activeDevice, err := m.spotifyClient.GetActiveDevice()
		if err != nil {
			return errMsg{err}
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.URI)
		if err != nil {
			return errMsg{err}
		}

		// Update the currently playing state immediately
		m.currentlyPlaying = &spotifyPkg.CurrentlyPlaying{
			Playing:  true,
			Item:     &track.FullTrack,
			Progress: 0,
		}

		// Get artist names
		artists := ""
		for i, artist := range track.Artists {
			if i > 0 {
				artists += ", "
			}
			artists += artist.Name
		}

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Name, artists)}
	}
}
