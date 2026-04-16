package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// snapMsg is the internal Bubble Tea message carrying a fresh data snapshot.
type snapMsg Snapshot

// scrollTickMsg fires on each auto-scroll tick.
type scrollTickMsg struct{}

const (
	// scrollHoldWindow is how recently a same-direction key must have arrived
	// for the ticker to continue fast-scrolling.  OS key-repeat rate is ~30ms,
	// so 120ms gives plenty of slack without triggering on a single tap.
	scrollHoldWindow = 120 * time.Millisecond
	// scrollTickRate is how often we fire an extra scroll event while held.
	scrollTickRate = 40 * time.Millisecond
)

// model is the root Bubble Tea model.
type model struct {
	activeTab  int
	snap       Snapshot
	queues     queuesModel
	channels   channelsModel
	topics     topicsModel
	subs       subsModel
	width      int
	height     int
	ready      bool
	showDetail bool      // true while the detail overlay is visible
	scrollDir  int       // -1=up, 0=none, +1=down
	scrollLast time.Time // time the scroll key was last received (same direction)
	scrollTick bool      // true while ticker goroutine is running
}

func newModel() model {
	return model{
		queues:   newQueuesModel(),
		channels: newChannelsModel(),
		topics:   newTopicsModel(),
		subs:     newSubsModel(),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Detail overlay absorbs ALL key presses — any key closes it.
		if m.showDetail {
			m.showDetail = false
			return m, nil
		}

		switch msg.String() {
		case "q", "Q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab + 3) % 4
			return m, nil
		case "1", "2", "3", "4":
			m.activeTab = int(msg.String()[0] - '1')
			return m, nil
		case " ":
			m.showDetail = true
			return m, nil
		case "up", "k":
			now := time.Now()
			wasHeld := m.scrollDir == -1 && now.Sub(m.scrollLast) < scrollHoldWindow
			m.scrollDir = -1
			m.scrollLast = now
			m.doScroll(-1)
			if wasHeld {
				return m, m.startScrollTick()
			}
			return m, nil
		case "down", "j":
			now := time.Now()
			wasHeld := m.scrollDir == 1 && now.Sub(m.scrollLast) < scrollHoldWindow
			m.scrollDir = 1
			m.scrollLast = now
			m.doScroll(1)
			if wasHeld {
				return m, m.startScrollTick()
			}
			return m, nil
		default:
			// Any other key stops the auto-scroll and is forwarded to the active panel.
			m.scrollDir = 0
			var cmd tea.Cmd
			switch m.activeTab {
			case 0:
				m.queues, cmd = m.queues.Update(msg)
			case 1:
				m.channels, cmd = m.channels.Update(msg)
			case 2:
				m.topics, cmd = m.topics.Update(msg)
			case 3:
				m.subs, cmd = m.subs.Update(msg)
			}
			return m, cmd
		}

	case scrollTickMsg:
		m.scrollTick = false
		if m.scrollDir != 0 && time.Since(m.scrollLast) < scrollHoldWindow {
			m.doScroll(m.scrollDir)
			return m, m.startScrollTick()
		}
		// Key released — stop ticking.
		m.scrollDir = 0
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeAll()
		return m, nil

	case snapMsg:
		m.snap = Snapshot(msg)
		m.queues.setData(m.snap.Queues)
		m.channels.setData(m.snap.Channels)
		m.topics.setData(m.snap.Topics)
		m.subs.setData(m.snap.Subs)
		return m, nil
	}

	return m, nil
}

// startScrollTick schedules the next scroll tick if one isn't already pending.
func (m *model) startScrollTick() tea.Cmd {
	if m.scrollTick {
		return nil // already scheduled
	}
	m.scrollTick = true
	return tea.Tick(scrollTickRate, func(t time.Time) tea.Msg { return scrollTickMsg{} })
}

// doScroll moves the active panel by one row in the given direction.
func (m *model) doScroll(dir int) {
	switch m.activeTab {
	case 0:
		m.queues.scrollBy(dir)
	case 1:
		m.channels.scrollBy(dir)
	case 2:
		m.topics.scrollBy(dir)
	case 3:
		m.subs.scrollBy(dir)
	}
}

func (m model) View() string {
	if !m.ready {
		return "\n  Connecting to IBM MQ…\n"
	}
	header := buildHeader(m.snap, m.width)
	tabBar := buildTabBar(m.activeTab, m.width)
	footer := buildFooter(m.width)
	if m.showDetail {
		contentH := m.contentHeight()
		overlay := m.activeDetailView(contentH)
		overlay = padToHeight(overlay, contentH)
		return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, overlay, footer)
	}
	content := m.activeContent()
	content = padToHeight(content, m.contentHeight())
	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, content, footer)
}

// contentHeight is how many rows the table body gets.
// Computed dynamically from actual rendered line counts so that any
// change in header / tabBar / footer height is handled automatically.
func (m *model) contentHeight() int {
	if m.width == 0 || m.height == 0 {
		return 2
	}
	header := buildHeader(m.snap, m.width)
	tabBar := buildTabBar(m.activeTab, m.width)
	footer := buildFooter(m.width)
	fixed := strings.Count(header, "\n") + 1 +
		strings.Count(tabBar, "\n") + 1 +
		strings.Count(footer, "\n") + 1
	h := m.height - fixed
	if h < 2 {
		return 2
	}
	return h
}

func (m *model) resizeAll() {
	h := m.contentHeight()
	m.queues.resize(m.width, h)
	m.channels.resize(m.width, h)
	m.topics.resize(m.width, h)
	m.subs.resize(m.width, h)
}

func (m model) activeContent() string {
	switch m.activeTab {
	case 0:
		return m.queues.View()
	case 1:
		return m.channels.View()
	case 2:
		return m.topics.View()
	default:
		return m.subs.View()
	}
}

// activeDetailView returns the detail overlay for the currently highlighted row.
func (m model) activeDetailView(contentH int) string {
	var title string
	var rows [][2]string
	switch m.activeTab {
	case 0:
		title, rows = m.queues.detailData()
	case 1:
		title, rows = m.channels.detailData()
	case 2:
		title, rows = m.topics.detailData()
	default:
		title, rows = m.subs.detailData()
	}
	return renderDetailPopup(title, rows, m.width, contentH)
}

// padToHeight pads s with blank lines so it occupies exactly h lines,
// ensuring the footer is always rendered at the bottom of the terminal.
func padToHeight(s string, h int) string {
	lines := strings.Count(s, "\n") + 1
	if s == "" {
		lines = 0
	}
	if pad := h - lines; pad > 0 {
		s += strings.Repeat("\n", pad)
	}
	return s
}

