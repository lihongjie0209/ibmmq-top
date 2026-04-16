package ui

import "strings"

// buildFooter renders the full-width key-binding help bar.
func buildFooter(width int, hideSystem bool) string {
	hideSysLabel := "HideSys"
	if hideSystem {
		hideSysLabel = "ShowSys"
	}
	keys := []struct{ k, desc string }{
		{"q", "Quit"},
		{"Tab", "NextTab"},
		{"1-4", "JumpTab"},
		{"↑↓/jk", "Scroll"},
		{"Space", "Detail"},
		{"s", "SortNext"},
		{"r", "SortRev"},
		{"h", hideSysLabel},
		{"/", "Search"},
	}
	var sb strings.Builder
	sb.WriteString(" ")
	for i, kv := range keys {
		sb.WriteString(footerKey.Render(kv.k))
		sb.WriteString(footerBg.Render(":" + kv.desc))
		if i < len(keys)-1 {
			sb.WriteString(footerSep.Render("  "))
		}
	}
	return footerBg.Width(width).Render(sb.String())
}

