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
var logFileLocation = fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
