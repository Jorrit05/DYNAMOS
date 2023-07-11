package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleFurtherProcessing(waitingJobName string, sqlResult *pb.SqlDataRequestResponse) {
	sqlResult.DestinationQueue = waitingJobName
	sqlResult.ReturnAddress = agentConfig.RoutingKey
	c.SendSqlDataRequestResponse(context.Background(), sqlResult)
}
