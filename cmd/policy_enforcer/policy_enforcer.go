package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
	c          pb.SideCarClient
	conn       *grpc.ClientConn
)

func main() {

	c, conn = lib.InitializeRabbit(grpcAddr, &pb.ServiceRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})
	defer conn.Close()

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		startConsumingWithRetry(c, fmt.Sprintf("%s-in", serviceName), 5, 5*time.Second)

		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	wg.Wait()

}
func checkRequestApproval(requestApproval *pb.RequestApproval) error {
	var agreements []lib.Agreement
	validDataProviders := make(map[string]struct{})
	invalidDataProviders := make(map[string]struct{})
	allArchetypes := make(map[string]int)
	allComputeProviders := make(map[string]int)
	protoRequest := &pb.ValidationResponse{
		Type: requestApproval.Type,
		User: &pb.User{
			ID:       requestApproval.User.ID,
			UserName: requestApproval.User.UserName,
		},
		Auth: &pb.Auth{
			AccessToken:  "1234",
			RefreshToken: "1234",
		},
	}
	getValidAgreements(requestApproval, &agreements)
	if len(agreements) == 0 {
		logger.Sugar().Info("No agreements exist for this user ")

		c.SendValidationResponse(context.Background(), protoRequest)
	}
	// Populate validDataProviders, allArchetypes and allComputeProviders
	for _, agreement := range agreements {
		relation, ok := agreement.Relations[requestApproval.User.UserName]
		if ok {
			validDataProviders[agreement.Name] = struct{}{}
			for _, archetype := range relation.AllowedArchetypes {
				allArchetypes[archetype]++
			}
			for _, provider := range relation.AllowedComputeProviders {
				allComputeProviders[provider]++
			}
		}
	}

	// Find invalidDataProviders
	for _, provider := range requestApproval.DataProviders {
		if _, ok := validDataProviders[provider]; !ok {
			invalidDataProviders[provider] = struct{}{}
		}
	}

	allowedArchetypes := filterCommon(allArchetypes, len(agreements))
	allowedComputeProviders := filterCommon(allComputeProviders, len(agreements))

	protoRequest = &pb.ValidationResponse{
		Type:                 "validationResponse",
		ValidDataProviders:   mapKeysToSlice(validDataProviders),
		InvalidDataProviders: mapKeysToSlice(invalidDataProviders),
		Auth: &pb.Auth{
			AccessToken:  "1234",
			RefreshToken: "1234",
		},
		User: &pb.User{
			ID:       requestApproval.User.ID,
			UserName: requestApproval.User.UserName,
		},
		AllowedArcheTypes:       allowedArchetypes,
		AllowedComputeProviders: allowedComputeProviders,
	}
	c.SendValidationResponse(context.Background(), protoRequest)

	return nil
}

// Returns a slice containing keys of the input map that have a value equal to the count
func filterCommon(inputMap map[string]int, count int) []string {
	var outputSlice []string
	for key, value := range inputMap {
		if value == count {
			outputSlice = append(outputSlice, key)
		}
	}
	return outputSlice
}

// Returns a slice containing keys of the input map
func mapKeysToSlice(inputMap map[string]struct{}) []string {
	var outputSlice []string
	for key := range inputMap {
		outputSlice = append(outputSlice, key)
	}
	return outputSlice
}

// In this function I want to simulate checking the policy Enforcer to see whether:
//   - I can have an agreement with each data steward
//   - Get a result channel or endpoint
//   - Return an access token
// func checkRequestApproval(requestApproval *pb.RequestApproval) error {
// 	var agreements []lib.Agreement
// 	protoRequest := &pb.ValidationResponse{
// 		Type: requestApproval.Type,
// 		User: &pb.User{
// 			ID:       requestApproval.User.ID,
// 			UserName: requestApproval.User.UserName,
// 		},
// 		Auth: &pb.Auth{
// 			AccessToken:  "1234",
// 			RefreshToken: "1234",
// 		},
// 	}
// 	getValidAgreements(requestApproval, &agreements)
// 	if len(agreements) == 0 {
// 		logger.Sugar().Info("No agreements exist for this user ")

// 		c.SendValidationResponse(context.Background(), protoRequest)
// 	}

// 	// Initialize allowedArchetypes and allowedComputeProviders with values from the first agreement
// 	allowedArchetypes := make(map[string]bool)
// 	for _, archetype := range agreements[0].Relations[requestApproval.User.UserName].AllowedArchetypes {
// 		allowedArchetypes[archetype] = true
// 	}

// 	allowedComputeProviders := make(map[string]bool)
// 	for _, computeProvider := range agreements[0].Relations[requestApproval.User.UserName].AllowedComputeProviders {
// 		allowedComputeProviders[computeProvider] = true
// 	}

// 	// Iterate over the remaining agreements and intersect allowedArchetypes and allowedComputeProviders
// 	for _, agreement := range agreements[1:] {
// 		newAllowedArchetypes := make(map[string]bool)
// 		for _, archetype := range agreement.Relations[requestApproval.User.UserName].AllowedArchetypes {
// 			if allowedArchetypes[archetype] {
// 				newAllowedArchetypes[archetype] = true
// 			}
// 		}
// 		allowedArchetypes = newAllowedArchetypes

// 		newAllowedComputeProviders := make(map[string]bool)
// 		for _, computeProvider := range agreement.Relations[requestApproval.User.UserName].AllowedComputeProviders {
// 			if allowedComputeProviders[computeProvider] {
// 				newAllowedComputeProviders[computeProvider] = true
// 			}
// 		}
// 		allowedComputeProviders = newAllowedComputeProviders

// 		protoRequest.ValidDataProviders = append(protoRequest.ValidDataProviders, agreement.Name)
// 	}

// 	// Add requested data providers that are not in validDataProviders to invalidDataProviders
// 	for _, dataProvider := range requestApproval.DataProviders {
// 		if !contains(protoRequest.ValidDataProviders, dataProvider) {
// 			protoRequest.InvalidDataProviders = append(protoRequest.InvalidDataProviders, dataProvider)
// 		}
// 	}

// 	// Convert allowedArchetypes and allowedComputeProviders from map to slice
// 	for archetype := range allowedArchetypes {
// 		protoRequest.AllowedArcheTypes = append(protoRequest.AllowedArcheTypes, archetype)
// 	}
// 	for computeProvider := range allowedComputeProviders {
// 		protoRequest.AllowedComputeProviders = append(protoRequest.AllowedComputeProviders, computeProvider)
// 	}

// 	c.SendValidationResponse(context.Background(), protoRequest)

// 	return nil
// }

// func contains(slice []string, item string) bool {
// 	for _, a := range slice {
// 		if a == item {
// 			return true
// 		}
// 	}
// 	return false
// }

// protoRequest := &pb.ValidationResponse{
// 	Type:                 requestApproval.Type,
// 	ValidDataProviders:   []string{"Provider1", "Provider2"},
// 	InvalidDataProviders: []string{"Provider3"},
// 	Auth: &pb.Auth{
// 		AccessToken:  "YourAccessToken",
// 		RefreshToken: "YourRefreshToken",
// 	},
// 	AllowedArcheTypes:       []string{"Type1", "Type2"},
// 	AllowedComputeProviders: []string{"ComputeProvider1", "ComputeProvider2"},

// }

func getValidAgreements(requestApproval *pb.RequestApproval, agreements *[]lib.Agreement) {
	for _, steward := range requestApproval.DataProviders {
		output, err := lib.GetValueFromEtcd(etcdClient, "/policyEnforcer/agreements/"+steward)
		if err != nil {
			logger.Sugar().Errorf("Error retrieving from etcd: %v", err)
		}

		if output == "" {
			logger.Sugar().Errorf("Steward not found: %s", steward)
			continue
		}

		var agreement lib.Agreement
		err = json.Unmarshal([]byte(output), &agreement)
		if err != nil {
			logger.Sugar().Errorw("%s: error unmarshalling agreement. %v", serviceName, err)
		}

		user, ok := agreement.Relations[requestApproval.User.UserName]
		if !ok {
			continue
		}

		agreement.Relations = map[string]lib.Relation{
			requestApproval.User.UserName: user,
		}
		*agreements = append(*agreements, agreement)
	}
}
