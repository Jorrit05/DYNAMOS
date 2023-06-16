//go:build !local
// +build !local

package main

import (
	"fmt"
)

var root = "/app/"
var serviceName = "orchestrator"
var port = ":8081"

var requestTypeConfigLocation = root + "requestType.json"
var archetypeConfigLocation = root + "archetype.json"
var microserviceMetadataConfigLocation = root + "microservices.json"
var agreementsConfigLocation = root + "agreements.json"
var agentConfigLocation = root + "agents_temp.json"

var logFileLocation = fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
var policyEnforcerEndpoint = "http://policyenforcer.svc.cluster.local:8082"
