//go:build local
// +build local

package main

import (
	"fmt"
)

var serviceName = "policyEnforcer"
var requestTypeConfigLocation = "/Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/stack/config/requestType.json"
var archetypeConfigLocation = "/Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/stack/config/archetype.json"
var microserviceMetadataConfigLocation = "/Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/stack/config/microservices.json"
var logFileLocation = fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
