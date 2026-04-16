package ui

import (
"fmt"
"sort"

tea "github.com/charmbracelet/bubbletea"
)

const (
tpSortStr = iota
tpSortPub
tpSortSub
tpSortMsgPub
tpSortCols
)

type topicsModel struct {
tbl     simpleTable
data    []TopicInfo
sorted  []TopicInfo // mirrors last rendered row order for detail lookup
sortCol int
sortAsc bool
width   int
height  int
}

func newTopicsModel() topicsModel { return topicsModel{sortAsc: true} }

func (m topicsModel) Update(msg tea.Msg) (topicsModel, tea.Cmd) {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "s":
m.sortCol = (m.sortCol + 1) % tpSortCols
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

func (m topicsModel) View() string { return m.tbl.View() }

func (m *topicsModel) scrollBy(n int) {
	if n < 0 {
		m.tbl.MoveUp(-n)
	} else {
		m.tbl.MoveDown(n)
	}
}

func (m *topicsModel) setData(data []TopicInfo) { m.data = data; m.render() }
func (m *topicsModel) resize(w, h int)          { m.width = w; m.height = h; m.render() }

func (m *topicsModel) render() {
if m.width == 0 || m.height == 0 {
return
}
cursor := m.tbl.Cursor()

// Fixed cols: Type(10)+Pubs(6)+Subs(6)+MsgPub(10)+MsgRcvd(10)=42; 6 cols × 2 = 12 padding.
topicW := clamp(m.width-42-12, 24, 60)

cols := []stColumn{
{Title: colHdr("Topic String", tpSortStr, m.sortCol, m.sortAsc), Width: topicW},
{Title: "Type", Width: 10},
{Title: colHdr("Pubs", tpSortPub, m.sortCol, m.sortAsc), Width: 6},
{Title: colHdr("Subs", tpSortSub, m.sortCol, m.sortAsc), Width: 6},
{Title: colHdr("MsgPub", tpSortMsgPub, m.sortCol, m.sortAsc), Width: 10},
{Title: "MsgRcvd", Width: 10},
}

sorted := make([]TopicInfo, len(m.data))
copy(sorted, m.data)
sort.Slice(sorted, func(i, j int) bool {
a, b := sorted[i], sorted[j]
var less bool
switch m.sortCol {
case tpSortPub:
less = a.Publishers < b.Publishers
case tpSortSub:
less = a.Subscribers < b.Subscribers
case tpSortMsgPub:
less = a.MsgPub < b.MsgPub
default:
less = a.TopicString < b.TopicString
}
if m.sortAsc {
return less
}
return !less
})

	m.sorted = sorted // save for detail lookup
	rows := make([][]string, 0, len(sorted))
	for _, tp := range sorted {
rows = append(rows, []string{
nameStyle.Render(tp.TopicString),
dimStyle.Render(tp.Type),
rateStyle.Render(fmt.Sprintf("%d", tp.Publishers)),
rateStyle.Render(fmt.Sprintf("%d", tp.Subscribers)),
rateStyle.Render(fmt.Sprintf("%d", tp.MsgPub)),
rateStyle.Render(fmt.Sprintf("%d", tp.MsgRcvd)),
})
}

t := newSimpleTable(cols, rows, m.height-2)
if cursor >= len(rows) {
cursor = len(rows) - 1
}
if cursor >= 0 {
t.SetCursor(cursor)
}
m.tbl = t
}

// detailData returns the title and field rows for the currently selected topic.
func (m *topicsModel) detailData() (string, [][2]string) {
	c := m.tbl.Cursor()
	if c < 0 || c >= len(m.sorted) {
		return "Topic Detail", [][2]string{{"Info", "No row selected"}}
	}
	tp := m.sorted[c]
	rows := [][2]string{
		{"Topic String", tp.TopicString},
		{"Type", tp.Type},
		{"Publishers", fmt.Sprintf("%d", tp.Publishers)},
		{"Subscribers", fmt.Sprintf("%d", tp.Subscribers)},
		{"Msgs Published", fmt.Sprintf("%d", tp.MsgPub)},
		{"Msgs Received", fmt.Sprintf("%d", tp.MsgRcvd)},
	}
	return "Topic Detail", rows
}
