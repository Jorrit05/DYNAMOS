//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var serviceName = "UVA"
var local = true
var etcdEndpoints = "http://localhost:30005"

var grpcAddr = "localhost:3005"
