//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var logLevel = zap.InfoLevel

var exchangeName = "topic_exchange"
var rabbitPort = "5672"
var rabbitDNS = "rabbitmq.core.svc.cluster.local"

// var policyEnforcerQueueName = "policyEnforcer"
