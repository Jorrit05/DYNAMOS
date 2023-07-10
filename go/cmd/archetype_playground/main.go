package main

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	Service  *MicroserviceMetada
	OutEdges []*Node
}

type RequestType struct {
	Type             string   `json:"type"`
	RequiredServices []string `json:"requiredServices"`
	OptionalServices []string `json:"optionalServices"`
}

type MicroserviceMetada struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"AllowedOutputs"`
}

type User struct {
	ID       string
	UserName string
}

type PolicyResult struct {
	ArchetypeId           string
	User                  User
	DataProviders         []string
	ComputeProvider       string
	RequestType           string
	RequiredMicroservices []string
}

type Archetype struct {
	ComputeProvider string
	ResultRecipient string
}

type CompositionRequest struct {
	Microservices []MicroserviceMetada
	ArchetypeId   string
	User          User
	RequestType   string
}

func (cr CompositionRequest) PrettyPrint() {
	crJSON, err := json.MarshalIndent(cr, "", "  ")
	if err != nil {
		fmt.Println("Error printing composition request:", err)
		return
	}

	fmt.Println(string(crJSON))
}

func main() {

	policyResult := PolicyResult{
		ArchetypeId: "computeToData",
		User: User{
			ID:       "GUID",
			UserName: "jstutterheim@uva.nl",
		},
		DataProviders:   []string{"VU", "UVA"},
		ComputeProvider: "",
		RequestType:     "sqlDataRequest",
		// RequiredMicroservices: []string{"anonymize_service"},
	}

	// Returns requiredServices & optionalServices for sqlDataRequest
	requestType := fetchRequestType("/requestType/" + policyResult.RequestType)
	fmt.Println(requestType)

	// Returns a slice of all labeled services contained in
	// requiredServices & optionalServices for sqlDataRequest
	allServices := fetchServices(&requestType)
	services := filterServices(&requestType, &allServices)
	fmt.Println("services")
	fmt.Println(services)

	err := generateGraphViz(requestType, services, "services.dot")
	if err != nil {
		panic(err)
	}
	// Merge requestType requiredServices with PolicyResult.RequiredMicroservices
	servicesToInclude := []string{}
	// servicesToInclude := append(requestType.RequiredServices, policyResult.RequiredMicroservices...)
	// fmt.Println("servicesToInclude")
	// fmt.Println(servicesToInclude)
	// fmt.Println("------------------------")

	// Generate a service microservice chain for the request
	chain, err := generateChain(servicesToInclude, services)
	if err != nil {
		fmt.Printf("Error generating chain: %v\n", err)
		return
	}

	archetype := fetchFromETCD("/archetypes/computeToData")
	// archetype := fetchFromETCD("/archetypes/dataThroughTtp")

	splits := splitServicesByArchetype(chain, archetype.ComputeProvider)

	// Print the service chain
	// fmt.Println("Generated service chain:")
	// for i, service := range chain {
	// 	fmt.Printf("%d: %s\n", i, service)
	// }
	// fmt.Println(chain)

	fmt.Println("Generated dataproviders chain:")
	for i, service := range splits.DataProviderServices {
		fmt.Printf("%d: %s\n", i, service)
	}
	fmt.Println("Generated computeproviders chain:")
	for i, service := range splits.ComputeProviderServices {
		fmt.Printf("%d: %s\n", i, service)
	}

}
