//go:build !local
// +build !local

package main

var root = "/app/etcd/"
var serviceName = "orchestrator"
var port = ":8080"
var grpcAddr = "localhost:3005"
var apiVersion = "/api/v1"

var requestTypeConfigLocation = root + "requestType.json"
var archetypeConfigLocation = root + "archetype.json"
var microserviceMetadataConfigLocation = root + "microservices.json"
var agreementsConfigLocation = root + "agreements.json"
var agentConfigLocation = root + "agents_temp.json"

var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"
var policyEnforcerEndpoint = "http://policyenforcer.svc.cluster.local:8082"
