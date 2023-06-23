//go:build local
// +build local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel

var serviceName = "UVA"
var local = true
var etcdEndpoints = "http://localhost:30005"
var port = ":8082"
var grpcAddr = "localhost:3005"
