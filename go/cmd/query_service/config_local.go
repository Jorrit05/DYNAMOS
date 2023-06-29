//go:build local
// +build local

package main

import (
	"fmt"
	"go.uber.org/zap"
	"path/filepath"
	"runtime"
)

var serviceName = "queryService"
var grpcAddr = "localhost:3005"
var logLevel = zap.DebugLevel
var etcdEndpoints = "http://localhost:30005"
