//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var serviceName = "policyEnforcer"

var logFileLocation = fmt.Sprintf("/Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/cmd/orchestrator/%s.log", serviceName)
var etcdEndpoints = "http://localhost:30005"
var port = ":8082"
