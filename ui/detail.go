package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	popupBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("220")).
				Padding(1, 2)
	popupTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)
	popupLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Width(20)
	popupValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)
	popupHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// renderDetailPopup draws a centered modal overlay inside an area of w×h.
// The background is dimmed with · chars.
func renderDetailPopup(title string, rows [][2]string, w, h int) string {
	var sb strings.Builder
	sb.WriteString(popupTitleStyle.Render(title))
	sb.WriteString("\n\n")
	for _, r := range rows {
		sb.WriteString(popupLabelStyle.Render(r[0]+":") + popupValueStyle.Render(r[1]) + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(popupHintStyle.Render("Space · Enter · Esc  to close"))

	popup := popupBorderStyle.Render(sb.String())

	return lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceChars("·"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("237")))
}
