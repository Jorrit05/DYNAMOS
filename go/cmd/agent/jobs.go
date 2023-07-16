package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

func PrettyPrintMicroserviceCommunication(msComm *pb.MicroserviceCommunication) {
	if msComm == nil {
		logger.Info("MicroserviceCommunication is nil")
		return
	}

	metadata := make([]string, 0, len(msComm.Metadata))
	for k, v := range msComm.Metadata {
		metadata = append(metadata, fmt.Sprintf("%s: %s", k, v))
	}

	requestMetadata := ""
	if msComm.RequestMetada != nil {
		requestMetadata = fmt.Sprintf("CorrelationId: %s, DestinationQueue: %s, JobName: %s, ReturnAddress: %s",
			msComm.RequestMetada.CorrelationId,
			msComm.RequestMetada.DestinationQueue,
			msComm.RequestMetada.JobName,
			msComm.RequestMetada.ReturnAddress,
		)
	}

	logger.Sugar().Infof("MicroserviceCommunication:\n  Type: %s\n  RequestType: %s\n  Metadata: {%s}\n  RequestMetada: {%s}\n",
		msComm.Type,
		msComm.RequestType,
		strings.Join(metadata, ", "),
		requestMetadata,
	)
}

func handleFurtherProcessing(ctx context.Context, waitingJobName string, msComm *pb.MicroserviceCommunication) {
	ctx, span := trace.StartSpan(ctx, serviceName+"/func: handleFurtherProcessing")
	defer span.End()

	msComm.RequestMetada.DestinationQueue = waitingJobName
	msComm.RequestMetada.ReturnAddress = agentConfig.RoutingKey
	PrettyPrintMicroserviceCommunication(msComm)
	logger.Sugar().Debugf("handleFurtherProcessing: %v", time.Now())

	c.SendMicroserviceComm(ctx, msComm)
}
