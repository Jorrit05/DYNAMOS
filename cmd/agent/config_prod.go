//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var logLevel = zap.InfoLevel
var serviceName = ""
var local = false
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
var port = ":8080"
var grpcAddr = "localhost:3005"
