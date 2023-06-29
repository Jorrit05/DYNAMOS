//go:build !local
// +build !local

package main

import (
	"fmt"
	"os"
)

var hostname = os.Getenv("CONTAINER_NAME")
var serviceName = fmt.Sprintf("%s_service", hostname)

var logFileLocation = fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
