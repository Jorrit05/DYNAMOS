package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func fetchFromETCD(key string) Archetype {
	// fetch from etcd here...
	fmt.Println(key)
	// For now, just return some hardcoded values for demonstration
	switch key {
	case "/archetypes/computeToData":
		return Archetype{
			ComputeProvider: "DataProvider",
			ResultRecipient: "Requestor",
		}
	case "/archetypes/dataThroughTtp":
		return Archetype{
			ComputeProvider: "other",
			ResultRecipient: "Requestor",
		}
	default:
		return Archetype{}
	}
}

func fetchRequestType(key string) RequestType {
	switch key {
	case "/requestType/sqlDataRequest":
		return RequestType{
			Type:             "sqlDataRequest",
			RequiredServices: []string{"query_service", "algorithm_service"},
			OptionalServices: []string{"aggregate_service", "anonymize_service", "graph_service"},
		}
	default:
		return RequestType{}
	}
}

func fetchServices(req *RequestType) []MicroserviceMetada {
	if req == nil {
		log.Print("RequestType is a nil pointer")
		return []MicroserviceMetada{}
	}
	// Define services
	servicesJson := `[
		{
			"name": "query_service",
			"label": "DataProvider",
			"allowedOutputs" : ["algorithm_service", "anonymize_service", "aggregate_service"]
		},
		{
			"name": "anonymize_service",
			"label": "DataProvider",
			"allowedOutputs" : ["algorithm_service", "aggregate_service"]
		},
		{
			"name": "aggregate_service",
			"label": "ComputeProvider",
			"allowedOutputs" : ["algorithm_service"]
		},
		{
			"name": "algorithm_service",
			"label": "ComputeProvider",
			"allowedOutputs" : ["graph_service"]
		},
		{
			"name": "graph_service",
			"label": "ComputeProvider",
			"allowedOutputs" : []
		}
	]`

	var allServices []MicroserviceMetada
	json.Unmarshal([]byte(servicesJson), &allServices)

	return allServices
}

func filterServices(req *RequestType, allServices *[]MicroserviceMetada) []MicroserviceMetada {

	// Create a map of the required and optional services for quick lookup
	serviceMap := make(map[string]bool)
	for _, service := range append(req.RequiredServices, req.OptionalServices...) {
		serviceMap[service] = true
	}

	// Filter the services to include only those that are required or optional
	var filteredServices []MicroserviceMetada
	for _, service := range *allServices {
		if serviceMap[service.Name] {
			filteredServices = append(filteredServices, service)
		}
	}

	return filteredServices
}
