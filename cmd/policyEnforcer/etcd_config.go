package main

import (
	"fmt"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
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
		lib.SaveStructToEtcd[lib.Agreement](etcdClient, fmt.Sprintf("/agreements/%s", agreement.Name), agreement)
	}
}
