package main

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkWaitingJob(ctx context.Context, waitingJob *batchv1.Job) bool {
	logger.Sugar().Debugf("Checking waiting job: %s", waitingJob.Name)
	job, err := clientSet.BatchV1().Jobs(strings.ToLower(serviceName)).Get(ctx, waitingJob.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Job does not exist
			logger.Sugar().Debugf("Job does not exist")
		} else {
			logger.Sugar().Errorf("Error getting Job:", err)
		}
		return false
	}

	// 2. Check the Job's status
	if job.Status.Active > 0 {
		// Job has active (running) containers
		logger.Sugar().Debugf("Job is running with active containers")
		return true
	} else if job.Status.Succeeded > 0 {
		// Job completed successfully
		logger.Sugar().Debugf("Job completed successfully")
		return false
	} else if job.Status.Failed > 0 {
		// Job failed
		logger.Sugar().Debugf("Job failed")
		return false
	} else {
		// Job exists but may be in a pending or unknown state
		logger.Sugar().Debugf("Job exists but its status is unclear")
		return false
	}

}

func isJobWaiting(ctx context.Context, msComm *pb.MicroserviceCommunication, correlationId string) bool {
	logger.Debug("Enter isJobWaiting")

	ctx, span := trace.StartSpan(ctx, "isJobWaiting")
	defer span.End()

	// Check if there is a job waiting for this result
	waitingJobMutex.Lock()
	waitingJob, ok := waitingJobMap[correlationId]
	waitingJobMutex.Unlock()

	if ok {
		ok = checkWaitingJob(ctx, waitingJob)
	}
	logger.Sugar().Debugf("Job waiting: %t", ok)
	if ok {
		// There was still a job waiting for this response
		handleFurtherProcessing(ctx, waitingJob.Name, msComm)
		// waitingJobMutex.Lock()
		// delete(waitingJobMap, correlationId)
		// waitingJobMutex.Unlock()
		return true
	}

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

	logger.Debug("Received microserviceCommunication")

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
