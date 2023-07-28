//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel
var serviceName = ""
var local = false
