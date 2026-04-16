package main

import (
"fmt"
"os"
"os/signal"
"strings"
"syscall"

"github.com/ibm-messaging/mq-golang/v5/ibmmq"
"github.com/ibm-messaging/mq-golang/v5/mqmetric"
log "github.com/sirupsen/logrus"

"github.com/ibmmq-top/mq-top/ui"
)

var (
BuildStamp    string
GitCommit     string
BuildPlatform string
)

func main() {
// Force TrueColor output regardless of what the container's TERM is set to.
// termenv (used by lipgloss) reads these at renderer-init time.
os.Setenv("COLORTERM", "truecolor")
os.Setenv("TERM", "xterm-256color")

log.SetOutput(os.Stderr)
log.SetLevel(log.WarnLevel)

if err := initConfig(); err != nil {
fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
os.Exit(1)
}

app := ui.NewApp()

// Signal handler — always active
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
go func() {
<-sigCh
app.Stop()
}()

if config.Demo {
// Demo mode: synthetic data, no MQ connection needed
go runDemo(app, config.Refresh)
} else {
// Live mode: connect to MQ and start collector
if err := initMQ(); err != nil {
fmt.Fprintf(os.Stderr, "Failed to connect to MQ: %v\n", err)
os.Exit(1)
}
defer mqmetric.EndConnection()

snapCh := make(chan Snapshot, 1)
go runCollector(config, snapCh)

go func() {
for snap := range snapCh {
app.Send(ui.Snapshot{
Timestamp: snap.Timestamp,
QMgr:      ui.QMgrInfo(snap.QMgr),
Queues:    toUIQueues(snap.Queues),
Channels:  toUIChannels(snap.Channels),
Topics:    toUITopics(snap.Topics),
Subs:      toUISubs(snap.Subs),
Error:     snap.Error,
})
}
}()
}

if err := app.Run(); err != nil {
fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
os.Exit(1)
}
}

func initMQ() error {
err := mqmetric.InitConnection(config.QMgrName, config.ReplyQ, "", &config.CC)
if err != nil {
if mqe, ok := err.(mqmetric.MQMetricError); ok {
if mqe.MQReturn.MQCC == ibmmq.MQCC_WARNING {
log.Warnf("MQ connection warning: %v", err)
err = nil
}
}
}
if err != nil {
return err
}

if config.QMgrName == "" || strings.HasPrefix(config.QMgrName, "*") {
config.QMgrName = mqmetric.GetResolvedQMgrName()
log.Infof("Resolved QMgr name: %s", config.QMgrName)
}

mqmetric.QueueInitAttributes()
mqmetric.ChannelInitAttributes()
mqmetric.TopicInitAttributes()
mqmetric.SubInitAttributes()
mqmetric.QueueManagerInitAttributes()

return nil
}

// ── Type adapters: internal → ui package ─────────────────────────────────────

func toUIQueues(qs []QueueInfo) []ui.QueueInfo {
out := make([]ui.QueueInfo, len(qs))
for i, q := range qs {
out[i] = ui.QueueInfo(q)
}
return out
}

func toUIChannels(cs []ChannelInfo) []ui.ChannelInfo {
out := make([]ui.ChannelInfo, len(cs))
for i, c := range cs {
out[i] = ui.ChannelInfo(c)
}
return out
}

func toUITopics(ts []TopicInfo) []ui.TopicInfo {
out := make([]ui.TopicInfo, len(ts))
for i, t := range ts {
out[i] = ui.TopicInfo(t)
}
return out
}

func toUISubs(ss []SubInfo) []ui.SubInfo {
out := make([]ui.SubInfo, len(ss))
for i, s := range ss {
out[i] = ui.SubInfo(s)
}
return out
}
