package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/mqmetric"
)

type Config struct {
	QMgrName          string
	ReplyQ            string
	MonitoredQueues   string
	MonitoredChannels string
	MonitoredTopics   string
	MonitoredSubs     string
	Refresh           time.Duration
	refreshStr        string
	CC                mqmetric.ConnectionConfig
	connName          string
	channel           string
	userId            string
	password          string
	server            string
	clientMode        bool
	Demo              bool
}

var config Config

func initConfig() error {
	flag.BoolVar(&config.Demo, "demo", false, "Run with synthetic demo data (no MQ connection required)")
	flag.StringVar(&config.QMgrName, "ibmmq.queueManager", "", "Queue manager name (empty = default)")
	flag.StringVar(&config.ReplyQ, "ibmmq.replyQueue", "SYSTEM.DEFAULT.MODEL.QUEUE", "Model queue for replies")
	flag.StringVar(&config.MonitoredQueues, "ibmmq.monitoredQueues", "*", "Comma-separated list of monitored queue patterns")
	flag.StringVar(&config.MonitoredChannels, "ibmmq.monitoredChannels", "*", "Comma-separated list of monitored channel patterns")
	flag.StringVar(&config.MonitoredTopics, "ibmmq.monitoredTopics", "*", "Comma-separated list of monitored topic patterns")
	flag.StringVar(&config.MonitoredSubs, "ibmmq.monitoredSubscriptions", "*", "Comma-separated list of monitored subscription patterns")
	flag.StringVar(&config.refreshStr, "refresh", "5s", "Refresh interval (e.g. 5s, 10s, 1m)")
	flag.StringVar(&config.connName, "ibmmq.connName", "", "MQ client connection name (host(port))")
	flag.StringVar(&config.channel, "ibmmq.channel", "", "MQ client channel name")
	flag.StringVar(&config.userId, "ibmmq.userId", "", "MQ user ID")
	flag.StringVar(&config.password, "ibmmq.password", "", "MQ password")
	flag.StringVar(&config.server, "ibmmq.server", "", "MQSERVER string: CHANNEL/TCP/host(port) (alternative to connName+channel)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mq-top: htop-like IBM MQ monitor\n\nUsage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	var err error
	config.Refresh, err = time.ParseDuration(config.refreshStr)
	if err != nil || config.Refresh < time.Second {
		return fmt.Errorf("invalid refresh interval %q (minimum 1s)", config.refreshStr)
	}

	// Build MQ ConnectionConfig
	config.CC.UserId = config.userId
	config.CC.Password = config.password
	config.CC.UseStatus = true
	config.CC.WaitInterval = 3 // seconds; used to set PCF message expiry

	if config.server != "" {
		os.Setenv("MQSERVER", config.server)
	}

	if config.connName != "" || config.channel != "" {
		config.CC.ClientMode = true
		config.CC.ConnName = config.connName
		config.CC.Channel = config.channel
	} else if config.server != "" || os.Getenv("MQSERVER") != "" {
		config.CC.ClientMode = true
	}

	return nil
}
