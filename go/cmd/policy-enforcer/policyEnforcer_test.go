package main

// import (
// 	"testing"

// 	"github.com/Jorrit05/DYNAMOS/go/pkg/lib"
// 	pb "github.com/Jorrit05/DYNAMOS/go/pkg/proto"
// 	"github.com/stretchr/testify/assert"
// 	// Import your pb and lib packages here
// )

// var requestApproval1 = &pb.RequestApproval{
// 	User: &pb.User{
// 		ID:       "1234",
// 		UserName: "jorrit.stutterheim@cloudnation.nl",
// 	},
// 	DataProviders: []string{"VU", "UVA"},
// }

// // Have to assume only agreements with the matching user are returned!
// var agreementSet1 = &[]lib.Agreement{
// 	{
// 		Name: "VU",
// 		Relations: map[string]lib.Relation{
// 			"jorrit.stutterheim@cloudnation.nl": {
// 				ID:                      "GUID",
// 				RequestTypes:            []string{"sqlDataRequest"},
// 				DataSets:                []string{"wageGap"},
// 				AllowedArchetypes:       []string{"computeToData", "dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 			},
// 		},
// 		ComputeProviders: []string{"surf", "otherCompany"},
// 		Archetypes:       []string{"computeToData", "dataThroughTtp", "reproducableScience"},
// 	},
// 	{
// 		Name: "UVA",
// 		Relations: map[string]lib.Relation{
// 			"jorrit.stutterheim@cloudnation.nl": {
// 				ID:                      "GUID",
// 				RequestTypes:            []string{"sqlDataRequest"},
// 				DataSets:                []string{"wageGap"},
// 				AllowedArchetypes:       []string{"computeToData", "dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 			},
// 		},
// 		ComputeProviders: []string{"surf", "otherCompany"},
// 		Archetypes:       []string{"computeToData", "dataThroughTtp", "reproducableScience"},
// 	},
// }

// var requestApproval2 = &pb.RequestApproval{
// 	User: &pb.User{
// 		ID:       "1234",
// 		UserName: "jorrit.stutterheim@cloudnation.nl",
// 	},
// 	DataProviders: []string{"VU", "UVA", "RUG"},
// }

// var agreementSet2 = &[]lib.Agreement{
// 	{
// 		Name: "VU",
// 		Relations: map[string]lib.Relation{
// 			"jorrit.stutterheim@cloudnation.nl": {
// 				ID:                      "GUID",
// 				RequestTypes:            []string{"sqlDataRequest"},
// 				DataSets:                []string{"wageGap"},
// 				AllowedArchetypes:       []string{"computeToData", "dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 			},
// 		},
// 		ComputeProviders: []string{"surf", "otherCompany"},
// 		Archetypes:       []string{"computeToData", "dataThroughTtp", "reproducableScience"},
// 	},
// 	{
// 		Name: "UVA",
// 		Relations: map[string]lib.Relation{
// 			"jorrit.stutterheim@cloudnation.nl": {
// 				ID:                      "GUID",
// 				RequestTypes:            []string{"sqlDataRequest"},
// 				DataSets:                []string{"wageGap"},
// 				AllowedArchetypes:       []string{"computeToData", "dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 			},
// 		},
// 		ComputeProviders: []string{"surf", "otherCompany"},
// 		Archetypes:       []string{"computeToData", "dataThroughTtp", "reproducableScience"},
// 	},
// 	{
// 		Name: "RUG",
// 		Relations: map[string]lib.Relation{
// 			"jorrit.stutterheim@cloudnation.nl": {
// 				ID:                      "GUID",
// 				RequestTypes:            []string{"sqlDataRequest"},
// 				DataSets:                []string{"wageGap"},
// 				AllowedArchetypes:       []string{"dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 			},
// 		},
// 		ComputeProviders: []string{"surf", "otherCompany"},
// 		Archetypes:       []string{"dataThroughTtp"},
// 	},
// }

// func TestValidateAgreements(t *testing.T) {
// 	// Setting up test cases
// 	testCases := []struct {
// 		name           string
// 		request        *pb.RequestApproval
// 		agreements     *[]lib.Agreement
// 		expectResponse *pb.ValidationResponse
// 	}{
// 		{
// 			name:       "Test case 1: With valid data providers",
// 			request:    requestApproval1,
// 			agreements: agreementSet1,
// 			expectResponse: &pb.ValidationResponse{
// 				Type:                 "validationResponse",
// 				ValidDataProviders:   []string{"VU", "UVA"},
// 				InvalidDataProviders: nil,
// 				Auth: &pb.Auth{
// 					AccessToken:  "1234",
// 					RefreshToken: "1234",
// 				},
// 				User: &pb.User{
// 					ID:       "1234",
// 					UserName: "jorrit.stutterheim@cloudnation.nl",
// 				},
// 				AllowedArcheTypes:       []string{"computeToData", "dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 				RequestApproved:         true,
// 			},
// 		},
// 		{
// 			name:       "Test case 2: With valid data providers",
// 			request:    requestApproval2,
// 			agreements: agreementSet2,
// 			expectResponse: &pb.ValidationResponse{
// 				Type:                 "validationResponse",
// 				ValidDataProviders:   []string{"VU", "RUG", "UVA"},
// 				InvalidDataProviders: nil,
// 				Auth: &pb.Auth{
// 					AccessToken:  "1234",
// 					RefreshToken: "1234",
// 				},
// 				User: &pb.User{
// 					ID:       "1234",
// 					UserName: "jorrit.stutterheim@cloudnation.nl",
// 				},
// 				AllowedArcheTypes:       []string{"dataThroughTtp"},
// 				AllowedComputeProviders: []string{"surf"},
// 				RequestApproved:         true,
// 			},
// 		},
// 		// Add more test cases as needed
// 	}

// 	// Run the test cases
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			response := validateAgreements(tc.request, tc.agreements)
// 			assert.Equal(t, tc.expectResponse, response, "They should be equal")
// 		})
// 	}
// }
