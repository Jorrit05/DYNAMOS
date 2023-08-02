package main

import (
	"context"
	"encoding/json"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

// In this function I want to simulate checking the policy Enforcer to see whether:
//   - I can have an agreement with each data steward
//   - Get a result channel or endpoint
//   - Return an access token
func checkRequestApproval(ctx context.Context, requestApproval *pb.RequestApproval) error {
	logger.Debug("Starting checkRequestApproval")

	// ctx, span := trace.StartSpan(ctx, serviceName+"/func: checkRequestApproval")

	var agreements []api.Agreement

	protoRequest := &pb.ValidationResponse{
		Type:        "validationResponse",
		RequestType: requestApproval.Type,
		User: &pb.User{
			Id:       requestApproval.User.Id,
			UserName: requestApproval.User.UserName,
		},
		RequestApproved: false,
	}

	getValidAgreements(requestApproval, &agreements, protoRequest)
	if len(agreements) == 0 || len(protoRequest.ValidDataproviders) == 0 {
		logger.Sugar().Info("No agreements exist for this user ")
		c.SendValidationResponse(ctx, protoRequest)
		return nil
	}

	protoRequest.RequestApproved = len(protoRequest.ValidDataproviders) > 0

	protoRequest.Auth = generateAuthToken()
	c.SendValidationResponse(ctx, protoRequest)

	return nil
}

func generateAuthToken() *pb.Auth {
	return &pb.Auth{
		AccessToken:  "1234",
		RefreshToken: "1234",
	}
}

func getValidAgreements(requestApproval *pb.RequestApproval, agreements *[]api.Agreement, protoRequest *pb.ValidationResponse) {
	var invalidDataproviders []string
	protoRequest.ValidDataproviders = make(map[string]*pb.DataProvider)

	for _, steward := range requestApproval.DataProviders {
		output, err := etcd.GetValueFromEtcd(etcdClient, "/policyEnforcer/agreements/"+steward)
		if err != nil {
			logger.Sugar().Errorf("Error retrieving from etcd: %v", err)
		}

		if output == "" {
			logger.Sugar().Infof("Steward not found: %s", steward)
			invalidDataproviders = append(invalidDataproviders, steward)
			continue
		}

		var agreement api.Agreement
		err = json.Unmarshal([]byte(output), &agreement)
		if err != nil {
			logger.Sugar().Errorw("%s: error unmarshalling agreement. %v", serviceName, err)
		}

		user, ok := agreement.Relations[requestApproval.User.UserName]
		if !ok {
			invalidDataproviders = append(invalidDataproviders, steward)
			continue
		}

		matchedArchetypes, _ := lib.GetMatchedElements(user.AllowedArchetypes, agreement.Archetypes)
		if len(matchedArchetypes) == 0 {
			logger.Sugar().Warn("No matching valid archetypes for this user in this agreement (config error): %s", steward)
			invalidDataproviders = append(invalidDataproviders, steward)
			continue
		}
		// Initalize after checking valid archetypes.
		protoRequest.ValidDataproviders[steward] = &pb.DataProvider{}
		// Check if user allowed archetypes are actually supported archetypes in this agreement
		protoRequest.ValidDataproviders[steward].Archetypes = matchedArchetypes
		// Add matching compute providers
		protoRequest.ValidDataproviders[steward].ComputeProviders, _ = lib.GetMatchedElements(user.AllowedComputeProviders, agreement.ComputeProviders)

		agreement.Relations = map[string]api.Relation{
			requestApproval.User.UserName: user,
		}

		*agreements = append(*agreements, agreement)

	}

	protoRequest.InvalidDataproviders = invalidDataproviders
}
