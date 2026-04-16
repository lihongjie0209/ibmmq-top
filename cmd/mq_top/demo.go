package main

import (
	"math/rand"
	"time"

	"github.com/ibmmq-top/mq-top/ui"
)

var demoStart = time.Now()

// runDemo generates synthetic IBM MQ data and sends it to the UI at regular intervals.
// Run with: mq-top -demo
func runDemo(app *ui.App, interval time.Duration) {
	// Base depths — give variety: near-full, medium, low, empty
	baseDepths := []int64{4250, 120, 43, 2800, 5, 350, 0, 1, 980, 3900, 225, 10, 4999, 600, 2200}
	basePut := []int64{10, 2, 0, 25, 20, 3, 0, 0, 15, 45, 8, 1, 0, 5, 12}
	baseGet := []int64{8, 2, 0, 24, 22, 1, 0, 0, 14, 40, 7, 1, 1, 5, 10}

	queueNames := []string{
		"DEV.QUEUE.1",
		"DEV.QUEUE.2",
		"DEV.DEAD.LETTER.QUEUE",
		"APP.IN.QUEUE",
		"APP.OUT.QUEUE",
		"APP.RETRY.QUEUE",
		"SYSTEM.ADMIN.COMMAND.QUEUE",
		"SYSTEM.DEFAULT.LOCAL.QUEUE",
		"ORDER.PROCESSING.QUEUE",
		"PAYMENT.EVENTS.QUEUE",
		"NOTIFICATION.QUEUE",
		"AUDIT.LOG.QUEUE",
		"CRITICAL.ALERT.QUEUE",
		"BATCH.JOBS.QUEUE",
		"TO.QM2",           // transmission queue
	}
	// true = transmission (XmitQ) queue
	queueXmitQ := []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}

	// Remote queue definitions (static — no depth/rate metrics)
	remoteQueues := []ui.QueueInfo{
		{Name: "ORDER.TO.QM2", QType: "REMOTE", RemoteName: "ORDER.PROCESSING.QUEUE", RemoteQMgr: "QM2", XmitQueue: "TO.QM2"},
		{Name: "PAYMENT.TO.QM3", QType: "REMOTE", RemoteName: "PAYMENT.EVENTS.QUEUE", RemoteQMgr: "QM3", XmitQueue: "TO.QM3"},
		{Name: "PASSTHROUGH.Q", QType: "REMOTE", RemoteName: "APP.IN.QUEUE", RemoteQMgr: "", XmitQueue: ""},
	}

	channelNames := []string{
		"TO.QM2",
		"FROM.QM2",
		"APP.SVRCONN",
		"ADMIN.SVRCONN",
		"TO.QM3.BACKUP",
		"CLUSTER.RCVR",
	}
	chTypes    := []string{"SDR", "RCVR", "SVRCONN", "SVRCONN", "SDR", "CLUSRCVR"}
	chStatuses := []string{"RUNNING", "RUNNING", "RUNNING", "RUNNING", "STOPPED", "RUNNING"}
	chConns    := []string{"10.0.1.50(1414)", "10.0.1.50(1414)", "10.0.0.1(54321)", "10.0.0.1(54456)", "10.0.1.60(1414)", "10.0.1.70(1414)"}
	chRemoteQM := []string{"QM2", "QM2", "", "", "QM3", "QM1.CLUSTER"}
	chMsgs     := []int64{18432, 9871, 4320, 112, 0, 7654}
	chSent     := []int64{92_340_000, 44_100_000, 1_200_000, 88_000, 0, 33_400_000}
	chRcvd     := []int64{31_000_000, 55_200_000, 800_000, 22_000, 0, 41_200_000}

	for {
		// ── Queues ────────────────────────────────────────────────────────────
		queues := make([]ui.QueueInfo, len(queueNames))
		for i, name := range queueNames {
			depth := baseDepths[i] + int64(rand.Intn(41)-20)
			if depth < 0 {
				depth = 0
			}
			put := basePut[i] + int64(rand.Intn(7)-3)
			get := baseGet[i] + int64(rand.Intn(7)-3)
			if put < 0 {
				put = 0
			}
			if get < 0 {
				get = 0
			}
			var msgAge int64
			if depth > 0 {
				msgAge = int64(rand.Intn(300))
			}
			queues[i] = ui.QueueInfo{
				Name:          name,
				Depth:         depth,
				MaxDepth:      5000,
				InputHandles:  int64(rand.Intn(4)),
				OutputHandles: int64(rand.Intn(4)),
				MsgAge:        msgAge,
				PutRate:       put,
				GetRate:       get,
				IsXmitQ:       queueXmitQ[i],
				QType: func() string {
					if queueXmitQ[i] {
						return "XMIT"
					}
					return "LOCAL"
				}(),
			}
		}
		queues = append(queues, remoteQueues...)

		// ── Channels ──────────────────────────────────────────────────────────
		channels := make([]ui.ChannelInfo, len(channelNames))
		for i, name := range channelNames {
			msgs := chMsgs[i] + int64(rand.Intn(21)-10)
			if msgs < 0 {
				msgs = 0
			}
			channels[i] = ui.ChannelInfo{
				Name:       name,
				Type:       chTypes[i],
				Status:     chStatuses[i],
				ConnName:   chConns[i],
				RemoteQMgr: chRemoteQM[i],
				Messages:   msgs,
				BytesSent:  chSent[i] + int64(rand.Intn(10001)-5000),
				BytesRcvd:  chRcvd[i] + int64(rand.Intn(10001)-5000),
				SinceMsg:   int64(rand.Intn(60)),
			}
		}

		// ── Topics ────────────────────────────────────────────────────────────
		topics := []ui.TopicInfo{
			{TopicString: "/orders/new", Type: "LOCAL", Publishers: 2, Subscribers: 3,
				MsgPub:  850 + int64(rand.Intn(21)-10),
				MsgRcvd: 848 + int64(rand.Intn(21)-10)},
			{TopicString: "/payments/confirmed", Type: "LOCAL", Publishers: 1, Subscribers: 2,
				MsgPub:  430 + int64(rand.Intn(11)-5),
				MsgRcvd: 430 + int64(rand.Intn(11)-5)},
			{TopicString: "/notifications/#", Type: "LOCAL", Publishers: 4, Subscribers: 8,
				MsgPub:  2100 + int64(rand.Intn(51)-25),
				MsgRcvd: 2098 + int64(rand.Intn(51)-25)},
			{TopicString: "/inventory/updates", Type: "LOCAL", Publishers: 1, Subscribers: 5,
				MsgPub:  320 + int64(rand.Intn(11)-5),
				MsgRcvd: 319 + int64(rand.Intn(11)-5)},
		}

		// ── Subscriptions ─────────────────────────────────────────────────────
		subs := []ui.SubInfo{
			{Name: "ORDER.PROCESSOR", SubId: "SUB.00000001", Topic: "/orders/new", Type: "USER",
				MsgRcvd: 850 + int64(rand.Intn(21)-10), SinceMsg: int64(rand.Intn(10))},
			{Name: "PAYMENT.HANDLER", SubId: "SUB.00000002", Topic: "/payments/confirmed", Type: "USER",
				MsgRcvd: 430 + int64(rand.Intn(11)-5), SinceMsg: int64(rand.Intn(10))},
			{Name: "NOTIFY.SVC.1", SubId: "SUB.00000003", Topic: "/notifications/#", Type: "USER",
				MsgRcvd: 700 + int64(rand.Intn(21)-10), SinceMsg: int64(rand.Intn(5))},
			{Name: "NOTIFY.SVC.2", SubId: "SUB.00000004", Topic: "/notifications/#", Type: "USER",
				MsgRcvd: 698 + int64(rand.Intn(21)-10), SinceMsg: int64(rand.Intn(5))},
			{Name: "INVENTORY.SYNC", SubId: "SUB.00000005", Topic: "/inventory/updates", Type: "USER",
				MsgRcvd: 319 + int64(rand.Intn(11)-5), SinceMsg: int64(rand.Intn(15))},
		}

		uptime := int64(time.Since(demoStart).Seconds())
		conns := int64(23 + rand.Intn(5))

		app.Send(ui.Snapshot{
			Timestamp: time.Now(),
			QMgr: ui.QMgrInfo{
				Name:            "QM1 [DEMO]",
				Status:          "RUNNING",
				Uptime:          uptime,
				ConnectionCount: conns,
				CHINITStatus:    "RUNNING",
				CMDSrvStatus:    "RUNNING",
			},
			Queues:   queues,
			Channels: channels,
			Topics:   topics,
			Subs:     subs,
		})

		time.Sleep(interval)
	}
}
