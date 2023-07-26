//go:build local
// +build local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel

var exchangeName = "topic_exchange"
var rabbitPort = "30020"
var rabbitDNS = "localhost"
var etcdEndpoints = "http://localhost:30005"
var grpcPort = 50051
