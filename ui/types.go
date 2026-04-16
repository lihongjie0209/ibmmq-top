// Package ui implements the htop-style terminal UI for mq-top using tview.
package ui

import "time"

// Snapshot is the data contract between the collector and the UI.
type Snapshot struct {
	Timestamp time.Time
	QMgr      QMgrInfo
	Queues    []QueueInfo
	Channels  []ChannelInfo
	Topics    []TopicInfo
	Subs      []SubInfo
	Error     string
}

type QMgrInfo struct {
	Name            string
	Status          string
	Uptime          int64
	ConnectionCount int64
	CHINITStatus    string
	CMDSrvStatus    string
}

type QueueInfo struct {
	Name          string
	Depth         int64
	MaxDepth      int64
	InputHandles  int64
	OutputHandles int64
	MsgAge        int64
	PutRate       int64
	GetRate       int64
	IsXmitQ       bool   // true when MQIA_USAGE = MQUS_TRANSMISSION
}

type ChannelInfo struct {
	Name       string
	Type       string
	Status     string
	ConnName   string
	RemoteQMgr string // remote QM name for SDR/RCVR/CLUSSDR channels
	Messages   int64
	BytesSent  int64
	BytesRcvd  int64
	SinceMsg   int64
}

type TopicInfo struct {
	TopicString string
	Type        string
	Publishers  int64
	Subscribers int64
	MsgPub      int64
	MsgRcvd     int64
}

type SubInfo struct {
	Name     string
	SubId    string
	Topic    string
	Type     string
	MsgRcvd  int64
	SinceMsg int64
}
