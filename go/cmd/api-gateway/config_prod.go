//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var serviceName = "api-gateway"
var port = ":8080"
var grpcAddr = "localhost:50051"
var apiVersion = "/api/v1"
var local = false
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
var logLevel = zap.DebugLevel
