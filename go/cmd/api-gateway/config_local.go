//go:build local
// +build local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel
var etcdEndpoints = "http://localhost:30005"
var local = true
