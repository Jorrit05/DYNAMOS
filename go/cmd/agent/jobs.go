package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleFurtherProcessing(waitingJobName string, msComm *pb.MicroserviceCommunication) {

	msComm.DestinationQueue = waitingJobName
	msComm.ReturnAddress = agentConfig.RoutingKey

	c.SendMicroserviceComm(context.Background(), msComm)
}
