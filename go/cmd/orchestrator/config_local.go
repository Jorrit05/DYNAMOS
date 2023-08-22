//go:build local
// +build local

package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
)

var serviceName = "orchestrator"
var port = ":8081"
var grpcAddr = "localhost:50051"
var apiVersion = "/api/v1"
var logLevel = zap.DebugLevel
var requestTypeConfigLocation = addEtcdDir("requestType.json")
var archetypeConfigLocation = addEtcdDir("archetype.json")
var microserviceMetadataConfigLocation = addEtcdDir("microservices.json")
var agreementsConfigLocation = addEtcdDir("agreements.json")
var dataSetConfigLocation = addEtcdDir("datasets.json")
var optionalMSConfigLocation = addEtcdDir("optional_microservices.json")

var etcdEndpoints = "http://localhost:30005"

func addEtcdDir(val string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("error")
	}
	dir := filepath.Dir(filename)

	path := fmt.Sprintf("%s/%s", filepath.Clean(filepath.Join(dir, "../../../configuration/etcd_launch_files/")), val)
	return path
}
