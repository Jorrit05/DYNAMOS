//go:build local
// +build local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel

var serviceName = "UVA"
var local = true
var etcdEndpoints = "http://localhost:30005"
var port = ":8082"
var grpcAddr = "localhost:50052"
var firstPortMicroservice = 50052
var backoffLimit = int32(6)
var ttl = int32(30)
var activeDeadlineSeconds = int64(600)
var kubeconfig = "/Users/jorrit/.kube/config"
var rabbitMqUser = "normal_user"
var etcdJobRootKey = "/agents/jobs"
var tracingHost = "localhost:32002"
var queueDeleteAfter = int64(600)
