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

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadPlaylists(),
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
		return m, nil

	case "up", "k":
		switch m.viewMode {
		case PlaylistView:
			if m.selectedPlaylist == nil {
				if m.playlistCursor > 0 {
					m.playlistCursor--
				}
			} else {
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
			if m.selectedPlaylist == nil {
				if m.playlistCursor < len(m.playlists)-1 {
					m.playlistCursor++
				}
			} else {
				if m.trackCursor < len(m.playlistTracks)-1 {
					m.trackCursor++
				}
			}
		case SearchView:
			if m.searchCursor < len(m.searchResults)-1 {
				m.searchCursor++
			}
		}
		return m, nil

	case "left", "h":
		if m.viewMode == PlaylistView && m.selectedPlaylist != nil {
			m.selectedPlaylist = nil
			m.playlistTracks = nil
			m.trackCursor = 0
		}
		return m, nil

	case "right", "l", "enter":
		switch m.viewMode {
		case PlaylistView:
			if m.selectedPlaylist == nil && len(m.playlists) > 0 {
				m.selectedPlaylist = &m.playlists[m.playlistCursor]
				return m, m.loadPlaylistTracks(m.playlists[m.playlistCursor].ID)
			} else if m.selectedPlaylist != nil && len(m.playlistTracks) > 0 {
				return m, m.playTrack(m.playlistTracks[m.trackCursor])
			}
		case SearchView:
			if len(m.searchResults) > 0 {
				return m, m.playSearchResult(m.searchResults[m.searchCursor])
			}
		}
		return m, nil

	case " ":
		return m, m.togglePlayback()

	case ">":
		return m, m.nextTrack()

	case "<":
		return m, m.previousTrack()
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

		return statusMsg{fmt.Sprintf("Playing: %s", track.Track.Name)}
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

		return statusMsg{fmt.Sprintf("Playing: %s", track.Name)}
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

			return statusMsg{fmt.Sprintf("Paused: %s", m.currentlyPlaying.Item.Name)}
		} else {
			err := m.spotifyClient.Resume(opts)
			if err != nil {
				return errMsg{err}
			}
			// Change the model state to reflect we just resumed the song
			m.currentlyPlaying.Playing = true

			return statusMsg{fmt.Sprintf("Playing: %s", m.currentlyPlaying.Item.Name)}
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

func (m Model) previousTrack() tea.Cmd {
	return func() tea.Msg {
		err := m.spotifyClient.Previous()
		if err != nil {
			fmt.Println(err)
			return errMsg{err}
		}
		return statusMsg{"Back to previous track"}
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
