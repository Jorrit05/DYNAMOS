package main

// INPUT
policyResult := PolicyResult{
	ArchetypeId: "computeToData", //  or "dataThroughTtp"
	User: User{
		ID:       "<GUID>",
		UserName: "jstutterheim@uva.nl",
	},
	DataProviders:         ["VU", "UVA"],
	ComputeProvider:       "surf",
	RequestType:           "sqlDataRequest",
}

// Returns requiredServices for sqlDataRequest
requestType := fetchRequestType("/requestType/" + policyResult.RequestType)

// Returns all labeled services for sqlDataRequest
allServices := fetchMicroservices(requestType)

// Generate a microservice chain using a topological sort
chain := generateChain(servicesToInclude, services)

// OUTPUT
// Generated chain with one optional service
// 0: {query_service}
// 1: {anonymize_service}
// 2: {algorithm_service}
//

// OUTPUT
// Generated for only a computeProvider
// 0: {aggregate_service}
// 1: {algorithm_service}
// 1: {graph_service}
