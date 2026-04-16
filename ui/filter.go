package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// filterState holds per-panel filter/search/hide-system state.
type filterState struct {
	hideSystem  bool
	searchQuery string
	searching   bool
}

// isSystem reports whether name is a system-defined IBM MQ object.
// This includes SYSTEM.*, $SYS/, and AMQ.* (internal dynamic queues).
func isSystem(name string) bool {
	upper := strings.ToUpper(name)
	return strings.HasPrefix(upper, "SYSTEM.") ||
		strings.HasPrefix(upper, "$SYS/") ||
		strings.HasPrefix(upper, "AMQ.")
}

// matchesFilter reports whether name contains query (case-insensitive).
// An empty query matches everything.
func matchesFilter(name, query string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(query))
}

// handle processes a key event for the filter state.
// Returns true if the event was consumed (and the caller should re-render).
func (f *filterState) handle(msg tea.KeyMsg) bool {
	if !f.searching {
		switch msg.String() {
		case "h":
			f.hideSystem = !f.hideSystem
			return true
		case "/":
			f.searching = true
			f.searchQuery = ""
			return true
		}
		return false
	}
	// In search mode: capture all input.
	switch msg.String() {
	case "esc":
		f.searching = false
		f.searchQuery = ""
		return true
	case "enter":
		f.searching = false
		return true
	case "backspace", "ctrl+h":
		if len(f.searchQuery) > 0 {
			runes := []rune(f.searchQuery)
			f.searchQuery = string(runes[:len(runes)-1])
		}
		return true
	default:
		if msg.Type == tea.KeyRunes {
			f.searchQuery += string(msg.Runes)
			return true
		}
	}
	return false
}

var searchBarStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("57")).
	Foreground(lipgloss.Color("255"))

// renderSearchBar renders a full-width search input bar.
func renderSearchBar(query string, width int) string {
	content := "/ " + query + "█  ESC:clear  Enter:confirm"
	return searchBarStyle.Width(width).Render(content)
}
