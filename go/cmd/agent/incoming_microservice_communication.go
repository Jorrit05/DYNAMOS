package main

import (
	"context"
	"fmt"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

func isJobWaiting(ctx context.Context, msComm *pb.MicroserviceCommunication, correlationId string) bool {
	logger.Debug("Enter isJobWaiting")

	ctx, span := trace.StartSpan(ctx, "isJobWaiting")
	defer span.End()

	// Check if there is a job waiting for this result
	waitingJobMutex.Lock()
	waitingJob, ok := waitingJobMap[correlationId]
	waitingJobMutex.Unlock()

	// Add additional trace attributes for debugging
	span.AddAttributes(trace.StringAttribute("correlationId", correlationId), trace.BoolAttribute("jobFound", ok))
	// Add waiting job attributes only if it is not nil
	if ok && waitingJob != nil {
		span.AddAttributes(trace.Int64Attribute("nrOfDataStewards", int64(waitingJob.nrOfDataStewards)))
	}
	logger.Sugar().Debugf("isJobWaiting: correlationId=%s, jobFound=%t, nrOfDataStewards=%v, mapContent=%v", correlationId, ok, waitingJob, waitingJobMap)

	if ok && waitingJob.nrOfDataStewards > 0 {
		logger.Sugar().Infof("Nr. of stewards: %d", waitingJob.nrOfDataStewards)
		handleFurtherProcessing(ctx, waitingJob.job.Name, msComm)
		waitingJob.nrOfDataStewards = waitingJob.nrOfDataStewards - 1
		if waitingJob.nrOfDataStewards == 0 {
			waitingJobMutex.Lock()
			delete(waitingJobMap, correlationId)
			waitingJobMutex.Unlock()
		}

		return true
	}
	logger.Sugar().Debugf("No job waiting: %t", ok)
	return false
}

func isHttpWaiting(ctx context.Context, msComm *pb.MicroserviceCommunication, correlationId string) bool {
	logger.Debug("Enter isHttpWaiting")

	ctx, span := trace.StartSpan(ctx, "isHttpWaiting")
	defer span.End()
	// Check if there is a http result waiting for this
	mutex.Lock()
	// Look up the corresponding channel in the request map
	dataResponseChan, ok := responseMap[correlationId]
	mutex.Unlock()

	// Add additional trace attributes for debugging
	span.AddAttributes(trace.StringAttribute("correlationId", correlationId), trace.BoolAttribute("httpChannelFound", ok))
	logger.Sugar().Debugf("isHttpWaiting: correlationId=%s, httpChannelFound=%t, mapContent=%v", correlationId, ok, responseMap)

	if ok {
		logger.Sugar().Info("Sending requestData to channel")

		// Send a signal on the channel to indicate that the response is ready
		dataResponseChan <- dataResponse{response: msComm, localContext: ctx}

		mutex.Lock()
		delete(responseMap, correlationId)
		mutex.Unlock()

		// logger.Debug("returning from responding......")
		return true
	}

	return false
}

func isThirdPartyWaiting(ctx context.Context, msComm *pb.MicroserviceCommunication, correlationId string) bool {
	logger.Debug("Enter isThirdPartyWaiting")

	ctx, span := trace.StartSpan(ctx, "isThirdPartyWaiting")
	defer span.End()

	// Check if there is a third party where this goes back to
	ttpMutex.Lock()
	returnAddress, ok := thirdPartyMap[correlationId]
	ttpMutex.Unlock()

	// Add additional trace attributes for debugging
	span.AddAttributes(trace.StringAttribute("correlationId", correlationId), trace.BoolAttribute("thirdPartyFound", ok))
	logger.Sugar().Debugf("isThirdPartyWaiting: correlationId=%s, thirdPartyFound=%t, mapContent=%v", correlationId, ok, thirdPartyMap)

	if ok {
		logger.Sugar().Infof("Sending sql response to returnAddress: %s", returnAddress)

		msComm.RequestMetadata.DestinationQueue = returnAddress

		c.SendMicroserviceComm(ctx, msComm)

		// logger.Debug("returning from forwarding to 3rd party......")
		return true
	}

	return false
}

func handleMicroserviceCommunication(ctx context.Context, grpcMsg *pb.SideCarMessage) error {

	logger.Debug("Start handleMicroserviceCommunication")

	msComm := &pb.MicroserviceCommunication{}
	msComm.RequestMetadata = &pb.RequestMetadata{}

	if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
	}

	correlationId := msComm.RequestMetadata.CorrelationId

	if isJobWaiting(ctx, msComm, correlationId) {
		return nil
	}

	if isHttpWaiting(ctx, msComm, correlationId) {
		return nil
	}

	if isThirdPartyWaiting(ctx, msComm, correlationId) {
		return nil
	}

	logger.Sugar().Errorw("unknown requestData response", "CorrelationId", correlationId)
	return fmt.Errorf("unknown requestData response")
}
