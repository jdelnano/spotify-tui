package tests

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

type SimpleModel struct {
	viewMode      int
	searchQuery   string
	isSearching   bool
	cursor        int
	width, height int
	quitting      bool
}

func (m SimpleModel) Init() tea.Cmd {
	return nil
}

func (m SimpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "/":
			m.isSearching = true
			m.searchQuery = ""
			return m, nil
		case "esc":
			if m.isSearching {
				m.isSearching = false
				return m, nil
			}
		case "backspace":
			if m.isSearching && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
		case "enter":
			if m.isSearching {
				m.isSearching = false
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.cursor++
		case "1":
			m.viewMode = 0
		case "2":
			m.viewMode = 1
		case "3":
			m.viewMode = 2
		default:
			if m.isSearching && len(msg.Runes) > 0 {
				m.searchQuery += string(msg.Runes)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m SimpleModel) View() string {
	if m.quitting {
		return ""
	}

	content := "Spotify TUI Test\n\n"

	if m.isSearching {
		content += "Search: " + m.searchQuery + "_\n"
	} else {
		switch m.viewMode {
		case 0:
			content += "Playlists View\n"
			content += "Press / to search, q to quit\n"
		case 1:
			content += "Now Playing View\n"
		case 2:
			content += "Library View\n"
		}
		content += "\nCursor: " + string(rune(m.cursor+'0')) + "\n"
	}

	return content
}

func TestSimpleNavigation(t *testing.T) {
	model := SimpleModel{viewMode: 0}

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return output != "" && output != "Spotify TUI Test\n\nPlaylists View\nPress / to search, q to quit\n\nCursor: 0\n"
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestSimpleSearch(t *testing.T) {
	model := SimpleModel{viewMode: 0}

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return output != "" && output != "Spotify TUI Test\n\nPlaylists View\nPress / to search, q to quit\n\nCursor: 0\n"
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	searchTerm := "test"
	for _, ch := range searchTerm {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return output != "" && output != "Spotify TUI Test\n\nSearch: _\n"
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return !isSearchMode(output)
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestSimpleCursorMovement(t *testing.T) {
	model := SimpleModel{viewMode: 0, cursor: 5}

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return output != "" && output != "Spotify TUI Test\n\nPlaylists View\nPress / to search, q to quit\n\nCursor: 5\n"
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		output := string(bts)
		return output != "" && output != "Spotify TUI Test\n\nPlaylists View\nPress / to search, q to quit\n\nCursor: 6\n"
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestSimpleWindowResize(t *testing.T) {
	model := SimpleModel{viewMode: 0}

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second))

	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	time.Sleep(100 * time.Millisecond)

	tm.Send(tea.WindowSizeMsg{Width: 60, Height: 20})

	time.Sleep(100 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func isSearchMode(output string) bool {
	return len(output) > 0 && output[0] == 'S' && len(output) > 7 && output[0:7] == "Search:"
}

