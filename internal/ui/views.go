package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("237"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	playlistPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	trackPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	libraryPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)
)

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.viewMode {
	case SearchView:
		return m.searchView()
	case NowPlayingView:
		return m.nowPlayingView()
	default:
		return m.mainView()
	}
}

func (m Model) mainView() string {
	libraryPanel := m.renderLibraryPanel()
	playlistPanel := m.renderPlaylistPanel()
	trackPanel := m.renderTrackPanel()

	leftColumnWidth := m.width / 3
	trackWidth := m.width - leftColumnWidth - 4

	libraryHeight := 10
	playlistHeight := m.height - libraryHeight - 4

	libraryPanel = lipgloss.NewStyle().
		Width(leftColumnWidth).
		Height(libraryHeight).
		Render(libraryPanel)

	playlistPanel = lipgloss.NewStyle().
		Width(leftColumnWidth).
		Height(playlistHeight).
		Render(playlistPanel)

	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		libraryPanel,
		playlistPanel,
	)

	trackPanel = lipgloss.NewStyle().
		Width(trackWidth).
		Height(m.height - 3).
		Render(trackPanel)

	main := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		trackPanel,
	)

	status := m.renderStatusBar()
	return lipgloss.JoinVertical(lipgloss.Left, main, status)
}

func (m Model) renderPlaylistPanel() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("🎵 Playlists") + "\n\n")

	if len(m.playlists) == 0 {
		content.WriteString(statusStyle.Render("No playlists found"))
	} else {
		visibleStart := 0
		visibleEnd := len(m.playlists)
		maxVisible := m.height - 18

		if len(m.playlists) > maxVisible {
			if m.playlistCursor >= maxVisible {
				visibleStart = m.playlistCursor - maxVisible + 1
			}
			visibleEnd = visibleStart + maxVisible
			if visibleEnd > len(m.playlists) {
				visibleEnd = len(m.playlists)
			}
		}

		for i := visibleStart; i < visibleEnd; i++ {
			playlist := m.playlists[i]
			name := playlist.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			line := fmt.Sprintf(" %s", name)
			if i == m.playlistCursor && m.selectedPlaylist == nil && m.libraryCursor == -1 {
				line = selectedStyle.Render("▶" + line)
			} else {
				line = normalStyle.Render(" " + line)
			}
			content.WriteString(line + "\n")
		}
	}

	return playlistPanelStyle.Render(content.String())
}

func (m Model) renderTrackPanel() string {
	var content strings.Builder

	if m.selectedLibraryItem != nil {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%s %s", m.selectedLibraryItem.Icon, m.selectedLibraryItem.Name)) + "\n\n")

		if m.selectedLibraryItem.Type == "saved_tracks" || m.selectedLibraryItem.Type == "recently_played" {
			if len(m.savedTracks) == 0 {
				content.WriteString(statusStyle.Render("No tracks found"))
			} else {
				visibleStart := 0
				visibleEnd := len(m.savedTracks)
				maxVisible := m.height - 8

				if len(m.savedTracks) > maxVisible {
					if m.trackCursor >= maxVisible {
						visibleStart = m.trackCursor - maxVisible + 1
					}
					visibleEnd = visibleStart + maxVisible
					if visibleEnd > len(m.savedTracks) {
						visibleEnd = len(m.savedTracks)
					}
				}

				for i := visibleStart; i < visibleEnd; i++ {
					track := m.savedTracks[i]

					artists := ""
					for j, artist := range track.Artists {
						if j > 0 {
							artists += ", "
						}
						artists += artist.Name
					}

					line := fmt.Sprintf(" %s - %s", track.Name, artists)
					if len(line) > 60 {
						line = line[:57] + "..."
					}

					if i == m.trackCursor && m.viewMode == PlaylistView {
						line = selectedStyle.Render("▶" + line)
					} else {
						line = normalStyle.Render(" " + line)
					}
					content.WriteString(line + "\n")
				}

				// Show loading indicator if loading more tracks
				if m.selectedLibraryItem.Type == "saved_tracks" && m.isLoadingMore {
					content.WriteString("\n" + statusStyle.Render(fmt.Sprintf("Loading more tracks... (%d/%d)", len(m.savedTracks), m.savedTracksTotal)))
				} else if m.selectedLibraryItem.Type == "saved_tracks" && len(m.savedTracks) < m.savedTracksTotal {
					content.WriteString("\n" + statusStyle.Render(fmt.Sprintf("Showing %d of %d tracks (scroll for more)", len(m.savedTracks), m.savedTracksTotal)))
				}
			}
		} else if m.selectedLibraryItem.Type == "saved_albums" {
			if len(m.savedAlbums) == 0 {
				content.WriteString(statusStyle.Render("No albums found"))
			} else {
				visibleStart := 0
				visibleEnd := len(m.savedAlbums)
				maxVisible := m.height - 8

				if len(m.savedAlbums) > maxVisible {
					if m.albumCursor >= maxVisible {
						visibleStart = m.albumCursor - maxVisible + 1
					}
					visibleEnd = visibleStart + maxVisible
					if visibleEnd > len(m.savedAlbums) {
						visibleEnd = len(m.savedAlbums)
					}
				}

				for i := visibleStart; i < visibleEnd; i++ {
					album := m.savedAlbums[i]

					artists := ""
					for j, artist := range album.Artists {
						if j > 0 {
							artists += ", "
						}
						artists += artist.Name
					}

					line := fmt.Sprintf(" %s - %s", album.Name, artists)
					if len(line) > 60 {
						line = line[:57] + "..."
					}

					if i == m.albumCursor && m.viewMode == PlaylistView {
						line = selectedStyle.Render("▶" + line)
					} else {
						line = normalStyle.Render(" " + line)
					}
					content.WriteString(line + "\n")
				}

				// Show loading indicator if loading more albums
				if m.isLoadingMoreAlbums {
					content.WriteString("\n" + statusStyle.Render(fmt.Sprintf("Loading more albums... (%d/%d)", len(m.savedAlbums), m.savedAlbumsTotal)))
				} else if len(m.savedAlbums) < m.savedAlbumsTotal {
					content.WriteString("\n" + statusStyle.Render(fmt.Sprintf("Showing %d of %d albums (scroll for more)", len(m.savedAlbums), m.savedAlbumsTotal)))
				}
			}
		} else if m.selectedLibraryItem.Type == "playlist" {
			if len(m.playlistTracks) == 0 {
				content.WriteString(statusStyle.Render("No tracks found"))
			} else {
				visibleStart := 0
				visibleEnd := len(m.playlistTracks)
				maxVisible := m.height - 8

				if len(m.playlistTracks) > maxVisible {
					if m.trackCursor >= maxVisible {
						visibleStart = m.trackCursor - maxVisible + 1
					}
					visibleEnd = visibleStart + maxVisible
					if visibleEnd > len(m.playlistTracks) {
						visibleEnd = len(m.playlistTracks)
					}
				}

				for i := visibleStart; i < visibleEnd; i++ {
					track := m.playlistTracks[i].Track

					artists := ""
					for j, artist := range track.Artists {
						if j > 0 {
							artists += ", "
						}
						artists += artist.Name
					}

					line := fmt.Sprintf(" %s - %s", track.Name, artists)
					if len(line) > 60 {
						line = line[:57] + "..."
					}

					if i == m.trackCursor && m.viewMode == PlaylistView {
						line = selectedStyle.Render("▶" + line)
					} else {
						line = normalStyle.Render(" " + line)
					}
					content.WriteString(line + "\n")
				}
			}
		}
	} else if m.selectedPlaylist != nil {
		content.WriteString(headerStyle.Render(fmt.Sprintf("🎵 %s", m.selectedPlaylist.Name)) + "\n\n")

		if len(m.playlistTracks) == 0 {
			content.WriteString(statusStyle.Render("No tracks in playlist"))
		} else {
			visibleStart := 0
			visibleEnd := len(m.playlistTracks)
			maxVisible := m.height - 8

			if len(m.playlistTracks) > maxVisible {
				if m.trackCursor >= maxVisible {
					visibleStart = m.trackCursor - maxVisible + 1
				}
				visibleEnd = visibleStart + maxVisible
				if visibleEnd > len(m.playlistTracks) {
					visibleEnd = len(m.playlistTracks)
				}
			}

			for i := visibleStart; i < visibleEnd; i++ {
				track := m.playlistTracks[i].Track

				artists := ""
				for j, artist := range track.Artists {
					if j > 0 {
						artists += ", "
					}
					artists += artist.Name
				}

				line := fmt.Sprintf(" %s - %s", track.Name, artists)
				if len(line) > 60 {
					line = line[:57] + "..."
				}

				if i == m.trackCursor && m.viewMode == PlaylistView {
					line = selectedStyle.Render("▶" + line)
				} else {
					line = normalStyle.Render(" " + line)
				}
				content.WriteString(line + "\n")
			}
		}
	} else {
		content.WriteString(headerStyle.Render("🎵 Tracks") + "\n\n")
		content.WriteString(statusStyle.Render("Select a library item or playlist to view tracks"))
	}

	return trackPanelStyle.Render(content.String())
}

func (m Model) searchView() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("🔍 Search Tracks") + "\n\n")

	searchBar := fmt.Sprintf("Search: %s", m.searchQuery)
	if m.isSearching {
		searchBar += "█"
	}
	content.WriteString(searchBar + "\n\n")

	if len(m.searchResults) > 0 {
		content.WriteString("Results:\n")
		for i, track := range m.searchResults {
			if i >= 20 {
				break
			}

			artists := ""
			for j, artist := range track.Artists {
				if j > 0 {
					artists += ", "
				}
				artists += artist.Name
			}

			line := fmt.Sprintf(" %s - %s", track.Name, artists)
			if i == m.searchCursor {
				line = selectedStyle.Render("▶" + line)
			} else {
				line = normalStyle.Render(" " + line)
			}
			content.WriteString(line + "\n")
		}
	} else if m.searchQuery != "" && !m.isSearching {
		content.WriteString(statusStyle.Render("No results found"))
	}

	status := m.renderStatusBar()
	main := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-3).
		Padding(1, 2).
		Render(content.String())

	return lipgloss.JoinVertical(lipgloss.Left, main, status)
}

func (m Model) nowPlayingView() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("🎵 Now Playing") + "\n\n")

	if m.currentlyPlaying != nil && m.currentlyPlaying.Item != nil {
		track := m.currentlyPlaying.Item
		artists := ""
		for i, artist := range track.Artists {
			if i > 0 {
				artists += ", "
			}
			artists += artist.Name
		}

		content.WriteString(fmt.Sprintf("Track: %s\n", track.Name))
		content.WriteString(fmt.Sprintf("Artist: %s\n", artists))
		content.WriteString(fmt.Sprintf("Album: %s\n", track.Album.Name))

		progress := int(m.currentlyPlaying.Progress) / 1000
		duration := int(track.Duration) / 1000
		progressBar := m.renderProgressBar(progress, duration, 40)
		content.WriteString(fmt.Sprintf("\n%s\n", progressBar))

		content.WriteString(fmt.Sprintf("%d:%02d / %d:%02d\n",
			progress/60, progress%60,
			duration/60, duration%60))

		if m.currentlyPlaying.Playing {
			content.WriteString("\n▶ Playing")
		} else {
			content.WriteString("\n⏸ Paused")
		}
	} else {
		content.WriteString(statusStyle.Render("Nothing is currently playing"))
	}

	status := m.renderStatusBar()
	main := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-3).
		Padding(1, 2).
		Render(content.String())

	return lipgloss.JoinVertical(lipgloss.Left, main, status)
}

func (m Model) renderProgressBar(current, total, width int) string {
	if total == 0 {
		return strings.Repeat("─", width)
	}

	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("─", width-filled)
	return bar
}

func (m Model) renderLibraryPanel() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("📚 Library") + "\n\n")

	if len(m.libraryCategories) == 0 {
		content.WriteString(statusStyle.Render("Loading library..."))
	} else {
		for i, category := range m.libraryCategories {
			line := fmt.Sprintf("%s %s", category.Icon, category.Name)
			if category.ItemCount > 0 {
				line += fmt.Sprintf(" (%d)", category.ItemCount)
			}

			if i == m.libraryCursor && m.selectedPlaylist == nil {
				line = selectedStyle.Render("▶ " + line)
			} else {
				line = normalStyle.Render("  " + line)
			}
			content.WriteString(line + "\n")
		}
	}

	return libraryPanelStyle.Render(content.String())
}

func (m Model) renderStatusBar() string {
	help := "↑↓: Navigate | PgUp/PgDn: Fast scroll | Enter: Select | /: Search | n: Now Playing | Space: Play/Pause | q: Quit"
	if m.statusMessage != "" {
		help = m.statusMessage
	}
	return statusStyle.Render(help)
}
