//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var serviceName = "policyEnforcer"

var etcdEndpoints = "http://localhost:30005"

var grpcAddr = "localhost:3005"
