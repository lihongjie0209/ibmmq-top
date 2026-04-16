package ui

import (
"fmt"
"sort"

tea "github.com/charmbracelet/bubbletea"
)

const (
subSortName = iota
subSortTopic
subSortMsgs
subSortCols
)

type subsModel struct {
	tbl     simpleTable
	data    []SubInfo
	sorted  []SubInfo // mirrors last rendered row order for detail lookup
	sortCol int
	sortAsc bool
	width   int
	height  int
	filter  filterState
}

func newSubsModel() subsModel { return subsModel{sortAsc: true} }

func (m subsModel) Update(msg tea.Msg) (subsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filter.handle(msg) {
			m.render()
			return m, nil
		}
		switch msg.String() {
		case "s":
			m.sortCol = (m.sortCol + 1) % subSortCols
			m.render()
			return m, nil
		case "r":
			m.sortAsc = !m.sortAsc
			m.render()
			return m, nil
		}
	}
	return m, nil
}

func (m subsModel) View() string {
	if m.filter.searching {
		return m.tbl.View() + "\n" + renderSearchBar(m.filter.searchQuery, m.width)
	}
	return m.tbl.View()
}

func (m *subsModel) scrollBy(n int) {
	if n < 0 {
		m.tbl.MoveUp(-n)
	} else {
		m.tbl.MoveDown(n)
	}
}

func (m *subsModel) setData(data []SubInfo) { m.data = data; m.render() }
func (m *subsModel) resize(w, h int)        { m.width = w; m.height = h; m.render() }

func (m *subsModel) render() {
if m.width == 0 || m.height == 0 {
return
}
cursor := m.tbl.Cursor()

// Fixed cols: SubId(12)+Type(8)+MsgRcvd(9)+SinceMsg(9)=38; 6 cols × 2 = 12 padding.
avail := m.width - 38 - 12
subNameW := clamp(avail*40/100, 16, 45)
topicW := clamp(avail-subNameW, 20, 55)

cols := []stColumn{
{Title: colHdr("Subscription Name", subSortName, m.sortCol, m.sortAsc), Width: subNameW},
{Title: "SubId", Width: 12},
{Title: colHdr("Topic", subSortTopic, m.sortCol, m.sortAsc), Width: topicW},
{Title: "Type", Width: 8},
{Title: colHdr("MsgRcvd", subSortMsgs, m.sortCol, m.sortAsc), Width: 9},
{Title: "SinceMsg", Width: 9},
}

	// Apply hide-system and search filters.
	filtered := make([]SubInfo, 0, len(m.data))
	for _, s := range m.data {
		if m.filter.hideSystem && isSystem(s.Name) {
			continue
		}
		if !matchesFilter(s.Name, m.filter.searchQuery) {
			continue
		}
		filtered = append(filtered, s)
	}

	sort.Slice(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		aSys, bSys := isSystem(a.Name), isSystem(b.Name)
		if aSys != bSys {
			return !aSys
		}
		var less bool
		switch m.sortCol {
		case subSortTopic:
			less = a.Topic < b.Topic
		case subSortMsgs:
			less = a.MsgRcvd < b.MsgRcvd
		default:
			less = a.Name < b.Name
		}
		if m.sortAsc {
			return less
		}
		return !less
	})

	m.sorted = filtered // save for detail lookup
	rows := make([][]string, 0, len(filtered))
	for _, s := range filtered {
rows = append(rows, []string{
nameStyle.Render(s.Name),
dimStyle.Render(s.SubId),
dimStyle.Render(s.Topic),
dimStyle.Render(s.Type),
rateStyle.Render(fmt.Sprintf("%d", s.MsgRcvd)),
dimStyle.Render(fmt.Sprintf("%ds", s.SinceMsg)),
})
}

	tableH := m.height - 2
	if m.filter.searching {
		tableH--
	}
	t := newSimpleTable(cols, rows, tableH)
if cursor >= len(rows) {
cursor = len(rows) - 1
}
if cursor >= 0 {
t.SetCursor(cursor)
}
m.tbl = t
}

// detailData returns the title and field rows for the currently selected subscription.
func (m *subsModel) detailData() (string, [][2]string) {
	c := m.tbl.Cursor()
	if c < 0 || c >= len(m.sorted) {
		return "Subscription Detail", [][2]string{{"Info", "No row selected"}}
	}
	s := m.sorted[c]
	rows := [][2]string{
		{"Name", s.Name},
		{"Sub ID", s.SubId},
		{"Topic", s.Topic},
		{"Type", s.Type},
		{"Msgs Received", fmt.Sprintf("%d", s.MsgRcvd)},
		{"Since Last Msg", fmt.Sprintf("%ds", s.SinceMsg)},
	}
	return "Subscription Detail", rows
}
