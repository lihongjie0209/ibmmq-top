package ui

import (
"fmt"
"strings"

"github.com/charmbracelet/lipgloss"
)

// ── Color palette ─────────────────────────────────────────────────────────────

var (
// Header bar — dark navy background
headerBg   = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("255"))
titleStyle = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("220")).Bold(true)
hLabelStyle = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("39")).Bold(true)
hValueStyle = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("255"))
hSepStyle   = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("240"))
hDimStyle   = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("242"))

// Footer bar
footerBg    = lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("250"))
footerKey   = lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("220")).Bold(true)
footerSep   = lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("240"))

// Status colors
statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
statusStopped = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
statusWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
statusUnknown = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

// Queue depth gradient
depthGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
depthYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
depthRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

// Tab bar — active tab has top+side borders (open bottom = sits on content)
activeTabStyle = lipgloss.NewStyle().
Bold(true).
Foreground(lipgloss.Color("17")).
Background(lipgloss.Color("220")).
BorderStyle(lipgloss.NormalBorder()).
BorderTop(true).
BorderLeft(true).
BorderRight(true).
BorderBottom(false).
BorderForeground(lipgloss.Color("220")).
Padding(0, 1)
inactiveTabStyle = lipgloss.NewStyle().
Foreground(lipgloss.Color("244")).
Background(lipgloss.Color("234")).
Padding(0, 1)
tabBarBg = lipgloss.NewStyle().Background(lipgloss.Color("234")).Foreground(lipgloss.Color("244"))

// Table cell helpers
dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
xmitqStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)  // cyan — transmission queue
remoteQStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))             // sky blue — remote QM name

// Numeric rate/count colors
rateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
)

// statusStyle picks the right colour style for a status string.
func statusStyle(s string) lipgloss.Style {
switch s {
case "RUNNING":
return statusRunning
case "STOPPED", "DISCONNECTED", "STOPPING":
return statusStopped
case "STARTING", "RETRYING", "PAUSED", "QUIESCING", "STANDBY":
return statusWarning
default:
return statusUnknown
}
}

// depthColor returns the right style for a queue depth percentage.
func depthColor(p float64) lipgloss.Style {
switch {
case p >= 80:
return depthRed
case p >= 50:
return depthYellow
default:
return depthGreen
}
}

// renderUsageBar builds a colored htop-style bar + percentage string.
// Total visual width = barWidth + 1 (space) + 6 (" 99.9%") = barWidth + 7
func renderUsageBar(pct float64, barWidth int) string {
filled := int(pct / 100.0 * float64(barWidth))
if filled > barWidth {
filled = barWidth
}
if filled < 0 {
filled = 0
}
bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
label := fmt.Sprintf("%s %5.1f%%", bar, pct)
return depthColor(pct).Render(label)
}

// colHdr returns a column title with sort indicator if active.
func colHdr(name string, col, sortCol int, asc bool) string {
if col != sortCol {
return name
}
if asc {
return name + " ▲"
}
return name + " ▼"
}

// clamp restricts v to [lo, hi].
func clamp(v, lo, hi int) int {
if v < lo {
return lo
}
if v > hi {
return hi
}
return v
}

// truncate shortens s to at most n runes, appending "…" if needed.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}
