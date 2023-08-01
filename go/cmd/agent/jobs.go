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

	RequestMetadatata := ""
	if msComm.RequestMetadata != nil {
		RequestMetadatata = fmt.Sprintf("CorrelationId: %s, DestinationQueue: %s, JobName: %s, ReturnAddress: %s",
			msComm.RequestMetadata.CorrelationId,
			msComm.RequestMetadata.DestinationQueue,
			msComm.RequestMetadata.JobName,
			msComm.RequestMetadata.ReturnAddress,
		)
	}

	logger.Sugar().Infof("MicroserviceCommunication:\n  Type: %s\n  RequestType: %s\n  Metadata: {%s}\n  RequestMetadata: {%s}\n",
		msComm.Type,
		msComm.RequestType,
		strings.Join(metadata, ", "),
		RequestMetadatata,
	)
}

func handleFurtherProcessing(ctx context.Context, waitingJobName string, msComm *pb.MicroserviceCommunication) {
	ctx, span := trace.StartSpan(ctx, serviceName+"/func: handleFurtherProcessing")
	defer span.End()

	msComm.RequestMetadata.DestinationQueue = waitingJobName
	msComm.RequestMetadata.ReturnAddress = agentConfig.RoutingKey
	// PrettyPrintMicroserviceCommunication(msComm)
	logger.Sugar().Debugf("handleFurtherProcessing: %v", time.Now())

	c.SendMicroserviceComm(ctx, msComm)
}
