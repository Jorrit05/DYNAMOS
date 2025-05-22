//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var root = "/app/etcd/"
var serviceName = "orchestrator"
var port = ":8080"
var grpcAddr = "localhost:50051"
var apiVersion = "/api/v1"

// var logLevel = zap.DebugLevel

var logLevel = zap.DebugLevel
var requestTypeConfigLocation = root + "requestType.json"
var archetypeConfigLocation = root + "archetype.json"
var microserviceMetadataConfigLocation = root + "microservices.json"
var agreementsConfigLocation = root + "agreements.json"
var dataSetConfigLocation = root + "datasets.json"
var optionalMSConfigLocation = root + "optional_microservices.json"

var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
