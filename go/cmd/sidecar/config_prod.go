//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var exchangeName = "topic_exchange"
var rabbitPort = "5672"
var rabbitDNS = "rabbitmq.core.svc.cluster.local"
var grpcPort = 50051
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"

var logLevel = zap.DebugLevel

// var logLevel = zap.InfoLevel
