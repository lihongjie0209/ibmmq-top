package ui

// simpleTable is a lightweight ANSI-aware replacement for charmbracelet/bubbles/table.
//
// bubbles/table v0.20.0 uses runewidth.Truncate() on cell values, which counts
// ANSI escape code characters (e.g. '[', '3', '8') as visible width-1 characters.
// This causes pre-styled (lipgloss-rendered) cells to be truncated into broken
// escape sequences that the terminal silently discards — producing blank cells.
//
// simpleTable uses lipgloss.Width() (which correctly strips ANSI) for measuring,
// and a custom ansiTrunc() for truncation that preserves escape sequences.

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// stColumn defines a table column (same fields as bubbles/table.Column).
type stColumn struct {
	Title string
	Width int
}

// simpleTable stores columns, rows, cursor position and viewport offset.
type simpleTable struct {
	cols   []stColumn
	rows   [][]string
	cursor int
	height int // number of visible data rows (header not counted)
	offset int // index of first visible row
}

// newSimpleTable creates a simpleTable with the given columns, rows and height.
func newSimpleTable(cols []stColumn, rows [][]string, height int) simpleTable {
	t := simpleTable{cols: cols, rows: rows, height: height}
	return t
}

// Cursor returns the index of the selected row.
func (t simpleTable) Cursor() int { return t.cursor }

// SetCursor jumps to row n and adjusts the viewport.
func (t *simpleTable) SetCursor(n int) {
	if len(t.rows) == 0 {
		return
	}
	t.cursor = clamp(n, 0, len(t.rows)-1)
	t.adjustOffset()
}

// MoveUp scrolls the cursor up by n rows.
func (t *simpleTable) MoveUp(n int) {
	if len(t.rows) == 0 {
		return
	}
	t.cursor = clamp(t.cursor-n, 0, len(t.rows)-1)
	t.adjustOffset()
}

// MoveDown scrolls the cursor down by n rows.
func (t *simpleTable) MoveDown(n int) {
	if len(t.rows) == 0 {
		return
	}
	t.cursor = clamp(t.cursor+n, 0, len(t.rows)-1)
	t.adjustOffset()
}

func (t *simpleTable) adjustOffset() {
	if t.cursor < t.offset {
		t.offset = t.cursor
	} else if t.cursor >= t.offset+t.height {
		t.offset = t.cursor - t.height + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

// Styles used by simpleTable (package-level, built once).
var (
	stHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("57")).
			BorderBottom(true)
	stSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)
)

// View renders the header followed by visible rows.
func (t simpleTable) View() string {
	if t.height <= 0 {
		return ""
	}

	var sb strings.Builder

	// ── Header row ──────────────────────────────────────────────────────────
	hCells := make([]string, 0, len(t.cols))
	for _, col := range t.cols {
		if col.Width <= 0 {
			continue
		}
		title := runewidth.Truncate(col.Title, col.Width, "…")
		title += strings.Repeat(" ", col.Width-runewidth.StringWidth(title))
		hCells = append(hCells, stHeader.Padding(0, 1).Render(title))
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, hCells...))
	sb.WriteString("\n")

	// ── Data rows ────────────────────────────────────────────────────────────
	end := clamp(t.offset+t.height, 0, len(t.rows))
	for i := t.offset; i < end; i++ {
		sb.WriteString(t.renderRow(i))
		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (t simpleTable) renderRow(idx int) string {
	row := t.rows[idx]
	selected := idx == t.cursor

	parts := make([]string, 0, len(t.cols))
	for j, col := range t.cols {
		if col.Width <= 0 {
			continue
		}
		var raw string
		if j < len(row) {
			raw = row[j]
		}

		if selected {
			// Strip all ANSI codes; the selection background will be applied
			// to the entire row at once after joining.
			plain := ansiStrip(raw)
			plain = runewidth.Truncate(plain, col.Width, "…")
			plain += strings.Repeat(" ", col.Width-runewidth.StringWidth(plain))
			parts = append(parts, " "+plain+" ")
		} else {
			// ANSI-aware truncation keeps escape sequences intact.
			cell := ansiTrunc(raw, col.Width)
			cell = ansiPad(cell, col.Width)
			parts = append(parts, " "+cell+" ")
		}
	}

	joined := strings.Join(parts, "")
	if selected {
		return stSelected.Render(joined)
	}
	return joined
}

// ── ANSI helpers ─────────────────────────────────────────────────────────────

// ansiStrip removes all ANSI SGR escape sequences, returning plain text.
func ansiStrip(s string) string {
	var sb strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\033':
			inEsc = true
		case inEsc:
			if r == 'm' {
				inEsc = false
			}
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// ansiTrunc truncates s to at most maxW visible cells, preserving ANSI SGR
// escape sequences. A reset code is appended when truncation occurs so any
// open color sequence is properly closed.
func ansiTrunc(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s // already fits — return with ANSI codes intact
	}
	var sb strings.Builder
	visW := 0
	inEsc := false
	truncated := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			sb.WriteRune(r)
			continue
		}
		if inEsc {
			sb.WriteRune(r)
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		rw := runewidth.RuneWidth(r)
		if visW+rw > maxW-1 { // reserve 1 cell for the "…" tail
			sb.WriteString("…")
			truncated = true
			break
		}
		sb.WriteRune(r)
		visW += rw
	}
	if truncated || inEsc {
		sb.WriteString("\033[0m") // close any open ANSI sequence
	}
	return sb.String()
}

// ansiPad appends spaces to bring s to exactly targetW visible cells.
func ansiPad(s string, targetW int) string {
	w := lipgloss.Width(s)
	if w >= targetW {
		return s
	}
	return s + strings.Repeat(" ", targetW-w)
}
