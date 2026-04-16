package ui

import (
"fmt"
"sort"

tea "github.com/charmbracelet/bubbletea"
)

const (
qSortName = iota
qSortDepth
qSortPct
qSortPutRate
qSortGetRate
qSortCols
)

type queuesModel struct {
	tbl     simpleTable
	data    []QueueInfo
	sorted  []QueueInfo // mirrors last rendered row order for detail lookup
	sortCol int
	sortAsc bool
	width   int
	height  int
}

func newQueuesModel() queuesModel { return queuesModel{sortAsc: true} }

func (m queuesModel) Update(msg tea.Msg) (queuesModel, tea.Cmd) {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "s":
m.sortCol = (m.sortCol + 1) % qSortCols
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

func (m queuesModel) View() string { return m.tbl.View() }

func (m *queuesModel) scrollBy(n int) {
	if n < 0 {
		m.tbl.MoveUp(-n)
	} else {
		m.tbl.MoveDown(n)
	}
}

func (m *queuesModel) setData(data []QueueInfo) { m.data = data; m.render() }
func (m *queuesModel) resize(w, h int)          { m.width = w; m.height = h; m.render() }

func (m *queuesModel) render() {
if m.width == 0 || m.height == 0 {
return
}
cursor := m.tbl.Cursor() // preserve scroll position

// Fixed col widths: Type(5)+Depth(7)+Max(7)+Usage(17)+InHnd(6)+OutHnd(7)+MsgAge(8)+Put/s(7)+Get/s(7)=71
// Each of 10 columns adds 2 chars cell-padding → overhead = 20.
nameW := clamp(m.width-71-20, 20, 55)

cols := []stColumn{
{Title: colHdr("Queue Name", qSortName, m.sortCol, m.sortAsc), Width: nameW},
{Title: "Type", Width: 5},
{Title: colHdr("Depth", qSortDepth, m.sortCol, m.sortAsc), Width: 7},
{Title: "Max", Width: 7},
{Title: colHdr("Usage", qSortPct, m.sortCol, m.sortAsc), Width: 17},
{Title: "InHnd", Width: 6},
{Title: "OutHnd", Width: 7},
{Title: "MsgAge", Width: 8},
{Title: colHdr("Put/s", qSortPutRate, m.sortCol, m.sortAsc), Width: 7},
{Title: colHdr("Get/s", qSortGetRate, m.sortCol, m.sortAsc), Width: 7},
}

sorted := make([]QueueInfo, len(m.data))
copy(sorted, m.data)
sort.Slice(sorted, func(i, j int) bool {
a, b := sorted[i], sorted[j]
var less bool
switch m.sortCol {
case qSortDepth:
less = a.Depth < b.Depth
case qSortPct:
less = pct(a.Depth, a.MaxDepth) < pct(b.Depth, b.MaxDepth)
case qSortPutRate:
less = a.PutRate < b.PutRate
case qSortGetRate:
less = a.GetRate < b.GetRate
default:
less = a.Name < b.Name
}
if m.sortAsc {
return less
}
return !less
})

	m.sorted = sorted // save for detail lookup
	rows := make([][]string, 0, len(sorted))
	for _, q := range sorted {
		if q.QType == "REMOTE" {
			remoteLabel := "→" + q.RemoteQMgr
			if q.RemoteQMgr == "" {
				remoteLabel = "→(passthrough)"
			}
			rows = append(rows, []string{
				nameStyle.Render(q.Name),
				queueTypeCell(q.QType),
				dimStyle.Render("—"),
				dimStyle.Render("—"),
				remoteQStyle.Render(remoteLabel),
				dimStyle.Render("—"),
				dimStyle.Render("—"),
				dimStyle.Render("—"),
				dimStyle.Render("—"),
				dimStyle.Render("—"),
			})
			continue
		}
		p := pct(q.Depth, q.MaxDepth)
		rows = append(rows, []string{
			nameStyle.Render(q.Name),
			queueTypeCell(q.QType),
			depthColor(p).Render(fmt.Sprintf("%d", q.Depth)),
			dimStyle.Render(fmt.Sprintf("%d", q.MaxDepth)),
			renderUsageBar(p, 10), // 10 bar chars + " 99.9%" = 17 visual chars
			dimStyle.Render(fmt.Sprintf("%d", q.InputHandles)),
			dimStyle.Render(fmt.Sprintf("%d", q.OutputHandles)),
			dimStyle.Render(fmtMsgAge(q.MsgAge)),
			rateStyle.Render(fmt.Sprintf("%d", q.PutRate)),
			rateStyle.Render(fmt.Sprintf("%d", q.GetRate)),
		})
	}

t := newSimpleTable(cols, rows, m.height-2)
// Restore cursor — clamp to valid range
if cursor >= len(rows) {
cursor = len(rows) - 1
}
if cursor >= 0 {
t.SetCursor(cursor)
}
m.tbl = t
}

func pct(depth, maxDepth int64) float64 {
if maxDepth <= 0 {
return 0
}
return float64(depth) / float64(maxDepth) * 100.0
}

func fmtMsgAge(secs int64) string {
if secs <= 0 {
return "—"
}
if secs < 60 {
return fmt.Sprintf("%ds", secs)
}
if secs < 3600 {
return fmt.Sprintf("%dm%ds", secs/60, secs%60)
}
return fmt.Sprintf("%dh%dm", secs/3600, (secs%3600)/60)
}

// queueTypeCell renders a short colored type badge.
func queueTypeCell(qType string) string {
	switch qType {
	case "XMIT":
		return xmitqStyle.Render("XMIT")
	case "REMOTE":
		return remoteQStyle.Render("REM")
	default:
		return dimStyle.Render("LOCAL")
	}
}

// detailData returns the title and field rows for the currently selected queue.
func (m *queuesModel) detailData() (string, [][2]string) {
	c := m.tbl.Cursor()
	if c < 0 || c >= len(m.sorted) {
		return "Queue Detail", [][2]string{{"Info", "No row selected"}}
	}
	q := m.sorted[c]
	if q.QType == "REMOTE" {
		rows := [][2]string{
			{"Name", q.Name},
			{"Type", "REMOTE"},
			{"Remote Queue (RNAME)", q.RemoteName},
			{"Remote QMgr (RQMNAME)", q.RemoteQMgr},
			{"Xmit Queue (XMITQ)", q.XmitQueue},
		}
		return "Queue Detail", rows
	}
	p := pct(q.Depth, q.MaxDepth)
	qType := "LOCAL"
	if q.IsXmitQ {
		qType = "TRANSMISSION (XMIT)"
	}
	rows := [][2]string{
		{"Name", q.Name},
		{"Type", qType},
		{"Depth", fmt.Sprintf("%d", q.Depth)},
		{"Max Depth", fmt.Sprintf("%d", q.MaxDepth)},
		{"Usage", fmt.Sprintf("%.1f%%", p)},
		{"Input Handles", fmt.Sprintf("%d", q.InputHandles)},
		{"Output Handles", fmt.Sprintf("%d", q.OutputHandles)},
		{"Oldest Message", fmtMsgAge(q.MsgAge)},
		{"Put Rate", fmt.Sprintf("%d/s", q.PutRate)},
		{"Get Rate", fmt.Sprintf("%d/s", q.GetRate)},
	}
	return "Queue Detail", rows
}
