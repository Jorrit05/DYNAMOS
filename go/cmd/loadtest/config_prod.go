//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel
var serviceName = "loadTest"
var local = false
