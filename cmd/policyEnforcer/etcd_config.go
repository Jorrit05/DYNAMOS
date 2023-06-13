package main

import (
	"fmt"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
)

func registerPolicyEnforcerConfiguration() {
	// Load request types
	var requestsTypes []RequestType
	lib.UnmarshalJsonFile(requestTypeConfigLocation, &requestsTypes)

	for _, requestType := range requestsTypes {
		lib.SaveStructToEtcd[RequestType](etcdClient, fmt.Sprintf("/requestTypes/%s", requestType.Type), requestType)
	}

	// Load archetypes
	var archeTypes []Archetype
	lib.UnmarshalJsonFile(archetypeConfigLocation, &archeTypes)

	for _, archeType := range archeTypes {
		lib.SaveStructToEtcd[Archetype](etcdClient, fmt.Sprintf("/archetypes/%s", archeType.Name), archeType)
	}

	// Load labels and allowedOutputs (microservice.json)
	var microservices []MicroserviceMetadata

	lib.UnmarshalJsonFile(microserviceMetadataConfigLocation, &microservices)

	for _, microservice := range microservices {
		lib.SaveStructToEtcd[MicroserviceMetadata](etcdClient, fmt.Sprintf("/microservices/%s/chainMetadata", microservice.Name), microservice)
	}
}
