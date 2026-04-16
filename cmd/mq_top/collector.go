package main

import (
	"strings"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"github.com/ibm-messaging/mq-golang/v5/mqmetric"
	log "github.com/sirupsen/logrus"
)

// Snapshot holds a point-in-time view of all MQ metrics.
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
	Name        string
	Type        string
	Status      string
	ConnName    string
	RemoteQMgr  string // remote QM name for SDR/RCVR/CLUSSDR channels
	Messages    int64
	BytesSent   int64
	BytesRcvd   int64
	SinceMsg    int64
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
	Name    string
	SubId   string
	Topic   string
	Type    string
	MsgRcvd int64
	SinceMsg int64
}

func runCollector(cfg Config, out chan<- Snapshot) {
	for {
		snap := collect(cfg)
		select {
		case out <- snap:
		default:
		}
		if snap.Error != "" {
			log.Warnf("Collection error: %s – retrying in %s", snap.Error, cfg.Refresh)
		}
		time.Sleep(cfg.Refresh)
	}
}

func collect(cfg Config) Snapshot {
	snap := Snapshot{Timestamp: time.Now()}

	if err := collectQMgr(cfg, &snap); err != nil {
		snap.Error = err.Error()
		return snap
	}
	collectQueues(cfg, &snap)
	collectChannels(cfg, &snap)
	collectTopics(cfg, &snap)
	collectSubs(cfg, &snap)

	return snap
}

func collectQMgr(cfg Config, snap *Snapshot) error {
	if err := mqmetric.CollectQueueManagerStatus(); err != nil {
		return err
	}
	st := mqmetric.GetObjectStatus("", mqmetric.OT_Q_MGR)
	attrs := st.Attributes

	qi := QMgrInfo{Name: cfg.QMgrName}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_NAME]; ok {
		for _, v := range attr.Values {
			if !v.IsInt64 {
				qi.Name = strings.TrimSpace(v.ValueString)
			}
		}
	}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_STATUS]; ok {
		for _, v := range attr.Values {
			if v.IsInt64 {
				qi.Status = qmgrStatusString(int(v.ValueInt64))
			}
		}
	}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_UPTIME]; ok {
		for _, v := range attr.Values {
			if v.IsInt64 {
				qi.Uptime = v.ValueInt64
			}
		}
	}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_CONNECTION_COUNT]; ok {
		for _, v := range attr.Values {
			if v.IsInt64 {
				qi.ConnectionCount = v.ValueInt64
			}
		}
	}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_CHINIT_STATUS]; ok {
		for _, v := range attr.Values {
			if v.IsInt64 {
				qi.CHINITStatus = mqSvcStatusString(int(v.ValueInt64))
			}
		}
	}

	if attr, ok := attrs[mqmetric.ATTR_QMGR_CMD_SERVER_STATUS]; ok {
		for _, v := range attr.Values {
			if v.IsInt64 {
				qi.CMDSrvStatus = mqSvcStatusString(int(v.ValueInt64))
			}
		}
	}

	snap.QMgr = qi
	return nil
}

func collectQueues(cfg Config, snap *Snapshot) {
	if err := mqmetric.CollectQueueStatus(cfg.MonitoredQueues); err != nil {
		log.Debugf("CollectQueueStatus error: %v", err)
		return
	}
	st := mqmetric.GetObjectStatus("", mqmetric.OT_Q)
	attrs := st.Attributes

	keys := allKeys(attrs)
	for _, key := range keys {
		q := QueueInfo{Name: key}
		q.Depth = int64Val(attrs, mqmetric.ATTR_Q_DEPTH, key)
		q.MaxDepth = int64Val(attrs, mqmetric.ATTR_Q_MAX_DEPTH, key)
		q.InputHandles = int64Val(attrs, mqmetric.ATTR_Q_IPPROCS, key)
		q.OutputHandles = int64Val(attrs, mqmetric.ATTR_Q_OPPROCS, key)
		q.MsgAge = int64Val(attrs, mqmetric.ATTR_Q_MSGAGE, key)
		q.PutRate = int64Val(attrs, mqmetric.ATTR_Q_INTERVAL_PUT, key)
		q.GetRate = int64Val(attrs, mqmetric.ATTR_Q_INTERVAL_GET, key)
		q.IsXmitQ = int64Val(attrs, mqmetric.ATTR_Q_USAGE, key) == int64(ibmmq.MQUS_TRANSMISSION)
		snap.Queues = append(snap.Queues, q)
	}
}

func collectChannels(cfg Config, snap *Snapshot) {
	if err := mqmetric.CollectChannelStatus(cfg.MonitoredChannels); err != nil {
		log.Debugf("CollectChannelStatus error: %v", err)
		return
	}
	st := mqmetric.GetObjectStatus("", mqmetric.OT_CHANNEL)
	attrs := st.Attributes

	keys := allKeys(attrs)
	for _, key := range keys {
		ch := ChannelInfo{Name: key}

		if attr, ok := attrs[mqmetric.ATTR_CHL_TYPE]; ok {
			if v, ok := attr.Values[key]; ok {
				if v.IsInt64 {
					ch.Type = chlTypeString(int(v.ValueInt64))
				}
			}
		}
		if attr, ok := attrs[mqmetric.ATTR_CHL_STATUS]; ok {
			if v, ok := attr.Values[key]; ok {
				if v.IsInt64 {
					ch.Status = chlStatusString(int(v.ValueInt64))
				}
			}
		}
		if attr, ok := attrs[mqmetric.ATTR_CHL_CONNNAME]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					ch.ConnName = strings.TrimSpace(v.ValueString)
				}
			}
		}
		if attr, ok := attrs[mqmetric.ATTR_CHL_RQMNAME]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					ch.RemoteQMgr = strings.TrimSpace(v.ValueString)
				}
			}
		}
		ch.Messages = int64Val(attrs, mqmetric.ATTR_CHL_MESSAGES, key)
		ch.BytesSent = int64Val(attrs, mqmetric.ATTR_CHL_BYTES_SENT, key)
		ch.BytesRcvd = int64Val(attrs, mqmetric.ATTR_CHL_BYTES_RCVD, key)
		ch.SinceMsg = int64Val(attrs, mqmetric.ATTR_CHL_SINCE_MSG, key)
		snap.Channels = append(snap.Channels, ch)
	}
}

func collectTopics(cfg Config, snap *Snapshot) {
	if err := mqmetric.CollectTopicStatus(cfg.MonitoredTopics); err != nil {
		log.Debugf("CollectTopicStatus error: %v", err)
		return
	}
	st := mqmetric.GetObjectStatus("", mqmetric.OT_TOPIC)
	attrs := st.Attributes

	keys := allKeys(attrs)
	for _, key := range keys {
		t := TopicInfo{}
		if attr, ok := attrs[mqmetric.ATTR_TOPIC_STRING]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					t.TopicString = strings.TrimSpace(v.ValueString)
				}
			}
		}
		if t.TopicString == "" {
			t.TopicString = key
		}
		if attr, ok := attrs[mqmetric.ATTR_TOPIC_STATUS_TYPE]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					t.Type = strings.TrimSpace(v.ValueString)
				}
			}
		}
		t.Publishers = int64Val(attrs, mqmetric.ATTR_TOPIC_PUBLISHER_COUNT, key)
		t.Subscribers = int64Val(attrs, mqmetric.ATTR_TOPIC_SUBSCRIBER_COUNT, key)
		t.MsgPub = int64Val(attrs, mqmetric.ATTR_TOPIC_PUB_MESSAGES, key)
		t.MsgRcvd = int64Val(attrs, mqmetric.ATTR_TOPIC_SUB_MESSAGES, key)
		snap.Topics = append(snap.Topics, t)
	}
}

func collectSubs(cfg Config, snap *Snapshot) {
	if err := mqmetric.CollectSubStatus(cfg.MonitoredSubs); err != nil {
		log.Debugf("CollectSubStatus error: %v", err)
		return
	}
	st := mqmetric.GetObjectStatus("", mqmetric.OT_SUB)
	attrs := st.Attributes

	keys := allKeys(attrs)
	for _, key := range keys {
		s := SubInfo{}
		if attr, ok := attrs[mqmetric.ATTR_SUB_NAME]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					s.Name = strings.TrimSpace(v.ValueString)
				}
			}
		}
		if s.Name == "" {
			s.Name = key
		}
		if attr, ok := attrs[mqmetric.ATTR_SUB_ID]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					s.SubId = strings.TrimSpace(v.ValueString)
				}
			}
		}
		if attr, ok := attrs[mqmetric.ATTR_SUB_TOPIC_STRING]; ok {
			if v, ok := attr.Values[key]; ok {
				if !v.IsInt64 {
					s.Topic = strings.TrimSpace(v.ValueString)
				}
			}
		}
		if attr, ok := attrs[mqmetric.ATTR_SUB_TYPE]; ok {
			if v, ok := attr.Values[key]; ok {
				if v.IsInt64 {
					s.Type = subTypeString(int(v.ValueInt64))
				}
			}
		}
		s.MsgRcvd = int64Val(attrs, mqmetric.ATTR_SUB_MESSAGES, key)
		s.SinceMsg = int64Val(attrs, mqmetric.ATTR_SUB_SINCE_PUB_MSG, key)
		snap.Subs = append(snap.Subs, s)
	}
}

// allKeys returns unique object keys across all attributes in a StatusSet.
func allKeys(attrs map[string]*mqmetric.StatusAttribute) []string {
	seen := make(map[string]struct{})
	for _, attr := range attrs {
		for key := range attr.Values {
			seen[key] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for key := range seen {
		result = append(result, key)
	}
	return result
}

// int64Val returns the int64 value of an attribute for a specific key, or 0 if absent/string.
func int64Val(attrs map[string]*mqmetric.StatusAttribute, attrName string, key string) int64 {
	if attr, ok := attrs[attrName]; ok {
		if v, ok := attr.Values[key]; ok && v.IsInt64 {
			return v.ValueInt64
		}
	}
	return 0
}

func qmgrStatusString(status int) string {
	// MQItoString class "QMSTA" → MQQMSTA_RUNNING, MQQMSTA_QUIESCING, MQQMSTA_STANDBY
	s := ibmmq.MQItoString("QMSTA", status)
	s = strings.TrimPrefix(s, "MQQMSTA_")
	if s == "" {
		return "UNKNOWN"
	}
	return s
}

func mqSvcStatusString(status int) string {
	// MQItoString class "SVC_STATUS" → MQSVC_STATUS_RUNNING etc.
	s := ibmmq.MQItoString("SVC_STATUS", status)
	s = strings.TrimPrefix(s, "MQSVC_STATUS_")
	if s == "" {
		return "UNKNOWN"
	}
	return s
}

func chlStatusString(status int) string {
	// MQItoString class "CHS" → MQCHS_RUNNING, MQCHS_STOPPED etc.
	s := ibmmq.MQItoString("CHS", status)
	s = strings.TrimPrefix(s, "MQCHS_")
	if s == "" {
		return "UNKNOWN"
	}
	return s
}

func chlTypeString(ctype int) string {
	// MQItoString class "CHT" → MQCHT_SENDER, MQCHT_RECEIVER etc.
	s := ibmmq.MQItoString("CHT", ctype)
	s = strings.TrimPrefix(s, "MQCHT_")
	if s == "" {
		return "UNKNOWN"
	}
	return s
}

func subTypeString(subType int) string {
	// MQItoString class "SUBTYPE" → MQSUBTYPE_USER, MQSUBTYPE_ADMIN etc.
	s := ibmmq.MQItoString("SUBTYPE", subType)
	s = strings.TrimPrefix(s, "MQSUBTYPE_")
	if s == "" {
		return "UNKNOWN"
	}
	return s
}
