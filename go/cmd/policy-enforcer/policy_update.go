package main

import (
	"context"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func checkPolicyUpdate(ctx context.Context, policyUpdate *pb.PolicyUpdate) {
	var agreements []api.Agreement

	protoRequest := &pb.ValidationResponse{
		Type:        "policyUpdate",
		RequestType: policyUpdate.Type,
		User: &pb.User{
			Id:       policyUpdate.User.Id,
			UserName: policyUpdate.User.UserName,
		},
		RequestApproved: false,
	}

	getValidAgreements(policyUpdate.DataProviders, policyUpdate.User, &agreements, protoRequest)
	policyUpdate.ValidationResponse = &pb.ValidationResponse{}
	policyUpdate.RequestMetadata.DestinationQueue = "orchestrator-in"
	policyUpdate.ValidationResponse = protoRequest

	if len(agreements) == 0 || len(protoRequest.ValidDataproviders) == 0 {
		logger.Sugar().Warn("No more agreements exist for this user")
		c.SendPolicyUpdate(ctx, policyUpdate)

	}

	protoRequest.RequestApproved = len(protoRequest.ValidDataproviders) > 0

	c.SendPolicyUpdate(ctx, policyUpdate)
}
