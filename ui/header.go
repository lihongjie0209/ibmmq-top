package ui

import (
"fmt"
"strings"
"time"

"github.com/charmbracelet/lipgloss"
)

var tabNames = []string{"Queues", "Channels", "Topics", "Subscriptions"}

// buildHeader renders a full-width dark-background header bar (1 line).
func buildHeader(snap Snapshot, width int) string {
qm := snap.QMgr
var parts []string

parts = append(parts, titleStyle.Render(" mq-top "))
parts = append(parts, hSepStyle.Render("│"))

if qm.Name == "" {
parts = append(parts, hValueStyle.Render(" Connecting… "))
} else {
parts = append(parts, hValueStyle.Render(" "+qm.Name+" "))
parts = append(parts, hSepStyle.Render("│"))

dot := statusStyle(qm.Status).Copy().Background(lipgloss.Color("17")).Render("●")
parts = append(parts, headerBg.Render(" "))
parts = append(parts, dot)
parts = append(parts, hValueStyle.Render(" "+qm.Status+" "))
parts = append(parts, hSepStyle.Render("│"))

parts = append(parts, hLabelStyle.Render(" Up:"))
parts = append(parts, hValueStyle.Render(fmtUptimeSecs(qm.Uptime)+" "))
parts = append(parts, hSepStyle.Render("│"))

parts = append(parts, hLabelStyle.Render(" Conn:"))
parts = append(parts, hValueStyle.Render(fmt.Sprintf("%d ", qm.ConnectionCount)))
parts = append(parts, hSepStyle.Render("│"))

parts = append(parts, hLabelStyle.Render(" CHINIT:"))
parts = append(parts, statusStyle(qm.CHINITStatus).Copy().Background(lipgloss.Color("17")).Render(" "+qm.CHINITStatus+" "))
parts = append(parts, hSepStyle.Render("│"))

parts = append(parts, hLabelStyle.Render(" CmdSrv:"))
parts = append(parts, statusStyle(qm.CMDSrvStatus).Copy().Background(lipgloss.Color("17")).Render(" "+qm.CMDSrvStatus+" "))
}

if !snap.Timestamp.IsZero() {
parts = append(parts, hSepStyle.Render("│"))
parts = append(parts, hDimStyle.Render(" "+snap.Timestamp.Format("15:04:05")+" "))
}

content := strings.Join(parts, "")
return headerBg.Width(width).Render(content)
}

// buildTabBar renders a 2-line tab row: active tab has a top border (cap),
// inactive tabs are bottom-aligned next to it. Gives an htop-style tab look.
func buildTabBar(active, width int) string {
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if i == active {
			tabs[i] = activeTabStyle.Render(label)
		} else {
			tabs[i] = inactiveTabStyle.Render(label)
		}
	}
	// Bottom-align so the shorter inactive tabs sit at the bottom of the
	// taller active tab, producing the visual "tab cap" effect.
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	return tabBarBg.Width(width).Render(row)
}

// fmtUptimeSecs formats uptime given in seconds.
func fmtUptimeSecs(secs int64) string {
if secs <= 0 {
return "—"
}
d := time.Duration(secs) * time.Second
days := int(d.Hours()) / 24
h := int(d.Hours()) % 24
m := int(d.Minutes()) % 60
if days > 0 {
return fmt.Sprintf("%dd%dh%dm", days, h, m)
}
return fmt.Sprintf("%dh%dm", h, m)
}
