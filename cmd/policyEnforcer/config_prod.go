//go:build !local
// +build !local

package main

import (
	"fmt"
)

var serviceName = "policyEnforcer"
var requestTypeConfigLocation = "/app/requestType.json"
var archetypeConfigLocation = "/app/archetype.json"
var microserviceMetadataConfigLocation = "/app/microservices.json"
var agreementsConfigLocation = "/app/agreements.json"

var logFileLocation = fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
