package main

import (
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
)

func registerPolicyEnforcerConfiguration() {
	// Load request types
	var requestsTypes []lib.RequestType
	lib.UnmarshalJsonFile(requestTypeConfigLocation, &requestsTypes)

	for _, requestType := range requestsTypes {
		lib.SaveStructToEtcd[lib.RequestType](etcdClient, fmt.Sprintf("/requestTypes/%s", requestType.Name), requestType)
	}

	// Load archetypes
	var archeTypes []lib.Archetype
	lib.UnmarshalJsonFile(archetypeConfigLocation, &archeTypes)

	for _, archeType := range archeTypes {
		lib.SaveStructToEtcd[lib.Archetype](etcdClient, fmt.Sprintf("/archetypes/%s", archeType.Name), archeType)
	}

	// Load labels and allowedOutputs (microservice.json)
	var microservices []lib.MicroserviceMetadata

	lib.UnmarshalJsonFile(microserviceMetadataConfigLocation, &microservices)

	for _, microservice := range microservices {
		lib.SaveStructToEtcd[lib.MicroserviceMetadata](etcdClient, fmt.Sprintf("/microservices/%s/chainMetadata", microservice.Name), microservice)
	}

	// Load agreemnents  (agreemnents.json)
	var agreements []lib.Agreement

	lib.UnmarshalJsonFile(agreementsConfigLocation, &agreements)

	for _, agreement := range agreements {
		lib.SaveStructToEtcd[lib.Agreement](etcdClient, fmt.Sprintf("/policyEnforcer/agreements/%s", agreement.Name), agreement)
	}

	// Load agents that are online. Temporary until agents are created  (agents_temp.json)
	var agents []Agent

	lib.UnmarshalJsonFile(agentConfigLocation, &agents)

	for _, agent := range agents {
		lib.SaveStructToEtcd[Agent](etcdClient, fmt.Sprintf("/dataStewards/%s", agent.Name), agent)
	}

}

type Agent struct {
	Name          string      `json:"name"`
	Services      interface{} `json:"services"`
	ActiveSince   interface{} `json:"ActiveSince"`
	ConfigUpdated interface{} `json:"ConfigUpdated"`
	RoutingKey    string      `json:"RoutingKey"`
}
