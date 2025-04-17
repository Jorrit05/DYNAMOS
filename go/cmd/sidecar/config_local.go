//go:build local
// +build local

package main

import "go.uber.org/zap"

var (
	logLevel            = zap.DebugLevel
	exchangeName        = "topic_exchange"
	rabbitPort          = "30020"
	mbtTestingQueueName = "mbt_testing_queue"
	useMbtQueue         = true
	rabbitDNS           = "localhost"
	etcdEndpoints       = "http://localhost:30005"
	grpcPort            = 50051
)
