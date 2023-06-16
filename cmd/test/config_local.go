//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var serviceName = "test"

var requestTypeConfigLocation = addEtcdDir("requestType.json")
var archetypeConfigLocation = addEtcdDir("archetype.json")
var microserviceMetadataConfigLocation = addEtcdDir("microservices.json")
var agreementsConfigLocation = addEtcdDir("agreements.json")
var agentConfigLocation = addEtcdDir("agents_temp.json")

var logFileLocation = fmt.Sprintf("/Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/cmd/test/%s.log", serviceName)
var etcdEndpoints = "http://localhost:30005"

func addEtcdDir(val string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("error")
	}
	dir := filepath.Dir(filename)

	path := fmt.Sprintf("%s/%s", filepath.Clean(filepath.Join(dir, "../../configuration/etcd_launch_files/")), val)
	return path
}
