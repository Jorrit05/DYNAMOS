//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var serviceName = "orchestrator"
var port = ":8081"

var requestTypeConfigLocation = addEtcdDir("requestType.json")
var archetypeConfigLocation = addEtcdDir("archetype.json")
var microserviceMetadataConfigLocation = addEtcdDir("microservices.json")
var agreementsConfigLocation = addEtcdDir("agreements.json")
var agentConfigLocation = addEtcdDir("agents_temp.json")

var logFileLocation = fmt.Sprintf("/Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/cmd/orchestrator/%s.log", serviceName)
var etcdEndpoints = "http://localhost:30005"
var policyEnforcerEndpoint = "http://localhost:8082"

func addEtcdDir(val string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("error")
	}
	dir := filepath.Dir(filename)

	path := fmt.Sprintf("%s/%s", filepath.Clean(filepath.Join(dir, "../../configuration/etcd_launch_files/")), val)
	return path
}
