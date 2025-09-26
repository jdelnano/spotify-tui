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

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPlaylists(),
		m.loadLibrary(),
		m.fetchCurrentlyPlaying(),
		tea.Every(time.Second*5, func(t time.Time) tea.Msg {
			return t
		}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.viewMode == NowPlayingView {
			return m, m.fetchCurrentlyPlaying()
		}
		return m, nil

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
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "/":
		m.viewMode = SearchView
		m.isSearching = true
		m.searchQuery = ""
		return m, nil

	case "n":
		m.viewMode = NowPlayingView
		return m, m.fetchCurrentlyPlaying()
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
				if m.trackCursor > 0 {
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
				if m.selectedLibraryItem.Type == "saved_tracks" && m.trackCursor < len(m.savedTracks)-1 {
					m.trackCursor++
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
				m.selectedLibraryItem = nil
				m.savedTracks = nil
				m.playlistTracks = nil
				m.trackCursor = 0
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
				if m.selectedLibraryItem.Type == "saved_tracks" && len(m.savedTracks) > 0 {
					// Add current track to recently played before playing new one
					if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
						m.recentlyPlayed = append(m.recentlyPlayed, &spotifyPkg.RecentlyPlayedItem{
							Track: m.currentlyPlaying.Item.SimpleTrack,
						})
					}
					return m, m.playSavedTrack(m.savedTracks[m.trackCursor])
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

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m Model) loadPlaylists() tea.Cmd {
	return func() tea.Msg {
		playlists, err := m.spotifyClient.GetPlaylists()
		if err != nil {
			return errMsg{err}
		}
		return playlistsLoadedMsg{playlists}
	}
}

func (m Model) loadPlaylistTracks(playlistID spotifyPkg.ID) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.spotifyClient.GetPlaylistTracks(playlistID)
		if err != nil {
			return errMsg{err}
		}
		return tracksLoadedMsg{tracks}
	}
}

func (m Model) searchTracks() tea.Cmd {
	return func() tea.Msg {
		results, err := m.spotifyClient.SearchTracks(m.searchQuery)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{results}
	}
}

func (m Model) playTrack(track spotifyPkg.PlaylistTrack) tea.Cmd {
	return func() tea.Msg {
		devices, err := m.spotifyClient.GetDevices()
		if err != nil {
			return errMsg{err}
		}

		if len(devices) == 0 {
			return statusMsg{"No active devices found. Please open Spotify on a device."}
		}

		var activeDevice spotifyPkg.ID
		for _, device := range devices {
			if device.Active {
				activeDevice = device.ID
				break
			}
		}

		if activeDevice == "" {
			activeDevice = devices[0].ID
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.Track.URI)
		if err != nil {
			return errMsg{err}
		}

		// Change the model state to reflect we started a song
		m.currentlyPlaying.Playing = true
		m.currentlyPlaying.Item = &track.Track

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Track.Name, track.Track.Artists[0].Name)}
	}
}

func (m Model) playSearchResult(track spotifyPkg.FullTrack) tea.Cmd {
	return func() tea.Msg {
		devices, err := m.spotifyClient.GetDevices()
		if err != nil {
			return errMsg{err}
		}

		if len(devices) == 0 {
			return statusMsg{"No active devices found. Please open Spotify on a device."}
		}

		var activeDevice spotifyPkg.ID
		for _, device := range devices {
			if device.Active {
				activeDevice = device.ID
				break
			}
		}

		if activeDevice == "" {
			activeDevice = devices[0].ID
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.URI)
		if err != nil {
			return errMsg{err}
		}
		// Change the model state to reflect we started a song
		m.currentlyPlaying.Playing = true
		m.currentlyPlaying.Item = &track

		return statusMsg{fmt.Sprintf("Playing: %s - %s", track.Name, track.Artists[0].Name)}
	}
}

func (m Model) togglePlayback() tea.Cmd {
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

		if m.currentlyPlaying.Playing {
			err := m.spotifyClient.Pause(opts)
			if err != nil {
				return errMsg{err}
			}
			// Change the model state to reflect we just pasued the song
			m.currentlyPlaying.Playing = false

			return statusMsg{fmt.Sprintf("Paused: %s - %s", m.currentlyPlaying.Item.Name, m.currentlyPlaying.Item.SimpleTrack.Artists[0].Name)}
		} else {
			err := m.spotifyClient.Resume(opts)
			if err != nil {
				return errMsg{err}
			}
			// Change the model state to reflect we just resumed the song
			m.currentlyPlaying.Playing = true

			return statusMsg{fmt.Sprintf("Playing: %s - %s", m.currentlyPlaying.Item.Name, m.currentlyPlaying.Item.SimpleTrack.Artists[0].Name)}
		}
	}
}

func (m Model) nextTrack() tea.Cmd {
	return func() tea.Msg {
		err := m.spotifyClient.Next()
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{"Skipped to next track"}
	}
}

func (m Model) previousTrack(track *spotifyPkg.RecentlyPlayedItem) tea.Cmd {
	return func() tea.Msg {
		if len(m.recentlyPlayed) == 0 {
			return statusMsg{"The previous track queue is empty"}
		}

		newCurrentTrack := spotifyPkg.PlaylistTrack{
			Track: spotifyPkg.FullTrack{
				SimpleTrack: track.Track,
			},
		}

		// play the "new" track
		m.playTrack(newCurrentTrack)

		return statusMsg{fmt.Sprintf("Playing: %s - %s", m.currentlyPlaying.Item.Name, m.currentlyPlaying.Item.Artists[0].Name)}
	}
}

func (m Model) fetchCurrentlyPlaying() tea.Cmd {
	return func() tea.Msg {
		playing, err := m.spotifyClient.GetCurrentlyPlaying()
		if err != nil {
			return currentlyPlayingMsg{nil}
		}
		return currentlyPlayingMsg{playing}
	}
}

func (m Model) loadLibrary() tea.Cmd {
	return func() tea.Msg {
		var categories []LibraryCategory

		// Fetch Liked Songs count
		savedTracks, err := m.spotifyClient.GetSavedTracks()

		if err == nil {
			categories = append(categories, LibraryCategory{
				Name:      "Liked Songs",
				Icon:      "❤️",
				Type:      "saved_tracks",
				ItemCount: min(len(savedTracks), 100),
			})
		}

		// Fetch Release Radar
		releaseRadar, err := m.spotifyClient.GetPlaylistByName("Release Radar")
		fmt.Println(releaseRadar)
		if err == nil && releaseRadar != nil {
			// Get track count for Release Radar
			tracks, _ := m.spotifyClient.GetPlaylistTracks(releaseRadar.ID)
			categories = append(categories, LibraryCategory{
				Name:       "Release Radar",
				Icon:       "📡",
				Type:       "playlist",
				ItemCount:  min(len(tracks), 100),
				PlaylistID: releaseRadar.ID,
			})
		}

		return libraryLoadedMsg{categories: categories}
	}
}

func (m Model) loadLibraryItems(category LibraryCategory) tea.Cmd {
	return func() tea.Msg {
		switch category.Type {
		case "saved_tracks":
			tracks, err := m.spotifyClient.GetSavedTracks()
			if err != nil {
				return errMsg{err}
			}
			return savedTracksLoadedMsg{tracks}
		case "playlist":
			tracks, err := m.spotifyClient.GetPlaylistTracks(category.PlaylistID)
			if err != nil {
				return errMsg{err}
			}
			return tracksLoadedMsg{tracks}
		default:
			return statusMsg{"This library category is not yet implemented"}
		}
	}
}

func (m Model) playSavedTrack(track spotifyPkg.SavedTrack) tea.Cmd {
	return func() tea.Msg {
		devices, err := m.spotifyClient.GetDevices()
		if err != nil {
			return errMsg{err}
		}

		if len(devices) == 0 {
			return statusMsg{"No active devices found. Please open Spotify on a device."}
		}

		var activeDevice spotifyPkg.ID
		for _, device := range devices {
			if device.Active {
				activeDevice = device.ID
				break
			}
		}

		if activeDevice == "" {
			activeDevice = devices[0].ID
		}

		err = m.spotifyClient.PlayTrack(activeDevice, track.URI)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{fmt.Sprintf("Playing: %s", track.Name)}
	}
}
