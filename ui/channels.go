package ui

import (
"fmt"
"sort"

tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
)

const (
chSortName = iota
chSortStatus
chSortMsgs
chSortBytes
chSortCols
)

type channelsModel struct {
tbl     simpleTable
data    []ChannelInfo
sorted  []ChannelInfo // mirrors last rendered row order for detail lookup
sortCol int
sortAsc bool
width   int
height  int
}

func newChannelsModel() channelsModel { return channelsModel{sortAsc: true} }

func (m channelsModel) Update(msg tea.Msg) (channelsModel, tea.Cmd) {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "s":
m.sortCol = (m.sortCol + 1) % chSortCols
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

func (m channelsModel) View() string { return m.tbl.View() }

func (m *channelsModel) scrollBy(n int) {
	if n < 0 {
		m.tbl.MoveUp(-n)
	} else {
		m.tbl.MoveDown(n)
	}
}

func (m *channelsModel) setData(data []ChannelInfo) { m.data = data; m.render() }
func (m *channelsModel) resize(w, h int)            { m.width = w; m.height = h; m.render() }

func (m *channelsModel) render() {
if m.width == 0 || m.height == 0 {
return
}
cursor := m.tbl.Cursor()

// Fixed cols: Type(8)+Status(11)+Msgs(9)+Sent(10)+Rcvd(10)+SinceMsg(9)=57; 8 cols × 2 = 16 padding.
avail := m.width - 57 - 16
nameW := clamp(avail*55/100, 18, 48)
remoteW := clamp(avail-nameW, 16, 36)

cols := []stColumn{
{Title: colHdr("Channel Name", chSortName, m.sortCol, m.sortAsc), Width: nameW},
{Title: "Type", Width: 8},
{Title: colHdr("Status", chSortStatus, m.sortCol, m.sortAsc), Width: 11},
{Title: "Remote QM / ConnName", Width: remoteW},
{Title: colHdr("Msgs", chSortMsgs, m.sortCol, m.sortAsc), Width: 9},
{Title: colHdr("Sent", chSortBytes, m.sortCol, m.sortAsc), Width: 10},
{Title: "Rcvd", Width: 10},
{Title: "SinceMsg", Width: 9},
}

sorted := make([]ChannelInfo, len(m.data))
copy(sorted, m.data)
sort.Slice(sorted, func(i, j int) bool {
a, b := sorted[i], sorted[j]
var less bool
switch m.sortCol {
case chSortStatus:
less = a.Status < b.Status
case chSortMsgs:
less = a.Messages < b.Messages
case chSortBytes:
less = a.BytesSent < b.BytesSent
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
	for _, c := range sorted {
rows = append(rows, []string{
nameStyle.Render(c.Name),
chlTypeCell(c.Type),
statusStyle(c.Status).Render(c.Status),
remoteQMOrConn(c.RemoteQMgr, c.ConnName),
rateStyle.Render(fmt.Sprintf("%d", c.Messages)),
rateStyle.Render(fmtBytes(c.BytesSent)),
rateStyle.Render(fmtBytes(c.BytesRcvd)),
dimStyle.Render(fmt.Sprintf("%ds", c.SinceMsg)),
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

func fmtBytes(b int64) string {
const k = 1024
switch {
case b >= k*k*k:
return fmt.Sprintf("%.1fG", float64(b)/float64(k*k*k))
case b >= k*k:
return fmt.Sprintf("%.1fM", float64(b)/float64(k*k))
case b >= k:
return fmt.Sprintf("%.1fK", float64(b)/float64(k))
default:
return fmt.Sprintf("%dB", b)
}
}

// chlTypeCell renders the channel type with a short color-coded abbreviation.
func chlTypeCell(t string) string {
switch t {
case "SDR", "CLUSSDR":
return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(t) // orange
case "RCVR", "CLUSRCVR":
return lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true).Render(t) // light blue
case "SVRCONN":
return lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(t) // green
case "REQUESTER", "SERVER":
return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(t) // yellow
default:
return dimStyle.Render(t)
}
}

// remoteQMOrConn shows the remote QM name when set (SDR/RCVR), else the connection IP.
func remoteQMOrConn(remoteQM, connName string) string {
if remoteQM != "" {
return remoteQStyle.Render(remoteQM)
}
return dimStyle.Render(connName)
}

// detailData returns the title and field rows for the currently selected channel.
func (m *channelsModel) detailData() (string, [][2]string) {
	c := m.tbl.Cursor()
	if c < 0 || c >= len(m.sorted) {
		return "Channel Detail", [][2]string{{"Info", "No row selected"}}
	}
	ch := m.sorted[c]
	remoteInfo := ch.ConnName
	if ch.RemoteQMgr != "" {
		remoteInfo = ch.RemoteQMgr + "  (" + ch.ConnName + ")"
	}
	rows := [][2]string{
		{"Name", ch.Name},
		{"Type", ch.Type},
		{"Status", ch.Status},
		{"Remote QM", ch.RemoteQMgr},
		{"Connection", ch.ConnName},
		{"Remote Info", remoteInfo},
		{"Messages", fmt.Sprintf("%d", ch.Messages)},
		{"Bytes Sent", fmtBytes(ch.BytesSent)},
		{"Bytes Rcvd", fmtBytes(ch.BytesRcvd)},
		{"Since Last Msg", fmt.Sprintf("%ds", ch.SinceMsg)},
	}
	return "Channel Detail", rows
}
