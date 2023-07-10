package main

// // INPUT
// policyResult := PolicyResult{
// 	ArchetypeId: "computeToData", //  or "dataThroughTtp"
// 	User: User{
// 		ID:       "GUID",
// 		UserName: "jstutterheim@uva.nl",
// 	},
// 	DataProviders:         []string{"VU", "UVA"},
// 	ComputeProvider:       "surf",
// 	RequestType:           "sqlDataRequest",
// }

// // Returns requiredServices & optionalServices for sqlDataRequest
// requestType := fetchRequestType("/requestType/" + policyResult.RequestType)

// // Returns all labeled services for sqlDataRequest
// allServices := fetchMicroservices(requestType)

// // Filter out the services relevant for `sqlDataRequest'
// services := filterServices(&requestType, &allServices)

// // Generate a microservice chain using a topological sort
// chain := generateChain(servicesToInclude, services)

// archetype := fetchArchetype("/archetypes/" + PolicyResult.ArchetypeId)
// splitServices := splitServicesByArchetype(chain, archetype.ComputeProvider)

// // OUTPUT computeToData:
// // Generated dataproviders chain:
// // 0: {query_service}
// // 1: {anonymize_service}
// // 2: {algorithm_service}
// // Generated computeproviders chain:
// //

// // OUTPUT dataThroughTtp
// // Generated dataproviders chain:
// // 0: {query_service}
// // 1: {anonymize_service}

// // Generated computeproviders chain:
// // 0: {algorithm_service}
