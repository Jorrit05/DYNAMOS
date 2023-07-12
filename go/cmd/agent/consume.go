package main

import (
	"context"
	"io"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func startConsumingWithRetry(c pb.SideCarClient, name string, maxRetries int, waitTime time.Duration) {
	for i := 0; i < maxRetries; i++ {
		err := startConsuming(c, name)
		if err == nil {
			return
		}

		logger.Sugar().Errorf("Failed to start consuming (attempt %d/%d): %v", i+1, maxRetries, err)

		// Wait for some time before retrying
		time.Sleep(waitTime)
	}
}

func startConsuming(c pb.SideCarClient, from string) error {
	stream, err := c.Consume(context.Background(), &pb.ConsumeRequest{QueueName: from, AutoAck: true})
	if err != nil {
		logger.Sugar().Errorf("Error on consume: %v", err)
	}

	for {
		grpcMsg, err := stream.Recv()
		if err == io.EOF {
			// The stream has ended.
			logger.Sugar().Warnw("Stream has ended", "error:", err)
			break
		}

		if err != nil {
			logger.Sugar().Fatalf("Failed to receive: %v", err)
		}

		logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

		switch grpcMsg.Type {
		case "compositionRequest":
			logger.Debug("Received compositionRequest")

			compositionRequest := &pb.CompositionRequest{}

			if err := grpcMsg.Body.UnmarshalTo(compositionRequest); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal compositionRequest message: %v", err)
			}
			go compositionRequestHandler(compositionRequest)
		case "microserviceCommunication":
			logger.Debug("Received microserviceCommunication")
			msComm := &pb.MicroserviceCommunication{}

			if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
			}

			// Check if there is a job waiting for this result
			waitingJobMutex.Lock()
			waitingJobName, ok := waitingJobMap[msComm.CorrelationId]
			waitingJobMutex.Unlock()

			if ok {
				// There was still a job waiting for this response
				handleFurtherProcessing(waitingJobName, msComm)
				waitingJobMutex.Lock()
				delete(waitingJobMap, msComm.CorrelationId)
				waitingJobMutex.Unlock()
				break
			}

			// Check if there is a http result waiting for this
			mutex.Lock()
			// Look up the corresponding channel in the request map
			requestData, ok := responseMap[msComm.CorrelationId]
			mutex.Unlock()

			if ok {
				logger.Sugar().Info("Sending requestData to channel")

				// Send a signal on the channel to indicate that the response is ready
				requestData.response <- msComm

				mutex.Lock()
				delete(responseMap, msComm.CorrelationId)
				mutex.Unlock()
				break
			}

			// Check if there is a third party where this goes back to
			ttpMutex.Lock()
			returnAddress, ok := thirdPartyMap[msComm.CorrelationId]
			ttpMutex.Unlock()

			if ok {
				logger.Sugar().Infof("Sending sql response to returnAddress: %s", returnAddress)
				// Send a signal on the channel to indicate that the response is ready
				msComm.DestinationQueue = returnAddress

				c.SendMicroserviceComm(context.Background(), msComm)
				break
			}
			logger.Sugar().Errorw("unknown requestData response", "CorrelationId", msComm.CorrelationId)

		case "sqlDataRequest":
			// Implicitly this means I am only a dataProvider
			logger.Debug("Received sqlDataRequest from Rabbit (third party)")
			sqlDataRequest := &pb.SqlDataRequest{}

			if err := grpcMsg.Body.UnmarshalTo(sqlDataRequest); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
			}

			waitingJobMutex.Lock()
			actualJobName, ok := waitingJobMap[sqlDataRequest.RequestMetada.JobName]
			waitingJobMutex.Unlock()

			ttpMutex.Lock()
			thirdPartyMap[sqlDataRequest.RequestMetada.CorrelationId] = sqlDataRequest.RequestMetada.ReturnAddress
			ttpMutex.Unlock()

			logger.Sugar().Warnf("jobName: %v", sqlDataRequest.RequestMetada.JobName)
			logger.Sugar().Warnf("actualJobName: %v", actualJobName)
			if ok {
				waitingJobMutex.Lock()
				delete(waitingJobMap, sqlDataRequest.RequestMetada.JobName)
				waitingJobMutex.Unlock()

				msComm := &pb.MicroserviceCommunication{}
				msComm.Type = sqlDataRequest.Type
				msComm.DestinationQueue = actualJobName
				msComm.ReturnAddress = agentConfig.RoutingKey
				msComm.CorrelationId = sqlDataRequest.RequestMetada.CorrelationId
				// Initialize the rest?
				any, err := anypb.New(sqlDataRequest)
				if err != nil {
					logger.Sugar().Error(err)
					return err
				}

				msComm.UserRequest = any

				logger.Sugar().Debugf("Sending SendMicroserviceInput to: %s", actualJobName)

				go c.SendMicroserviceComm(context.Background(), msComm)

			} else {
				logger.Sugar().Warnf("No job found for: %v", sqlDataRequest.RequestMetada.JobName)
			}

			// ttpMutex.Lock()
			// thirdPartyMap[sqlDataRequest.CorrelationId] = sqlDataRequest.ReturnAddress
			// ttpMutex.Unlock()

			// // /agents/jobs/UVA/activeJob/jorrit-3141334
			// activeJobKey := fmt.Sprintf("%s/%s/activeJob/%s", etcdJobRootKey, agentConfig.Name, jobName)
			// activeJob := ""
			// activeJob, err = etcd.GetValueFromEtcd(etcdClient, activeJobKey, etcd.WithMaxElapsedTime(15*time.Second))
			// if err != nil {
			// 	switch e := err.(type) {
			// 	case *etcd.ErrKeyNotFound:
			// 		logger.Sugar().Infof("Key not found error, deploying job anyway: %v", e.Error())
			// 		activeJob, err = generateChainAndDeploy(jobName)
			// 		if err != nil {
			// 			logger.Sugar().Errorf("Deploying failed: %v", e.Error())
			// 			continue
			// 		}
			// 	case *etcd.ErrEtcdOperation:
			// 		logger.Sugar().Errorf("Etcd operation error: %v", e.Error())
			// 		continue
			// 	default:
			// 		logger.Sugar().Errorf("Unknown error: %v", e.Error())
			// 		continue
			// 	}
			// }

		default:
			logger.Sugar().Fatalf("Unknown message type: %s", grpcMsg.Type)
		}
	}
	return err
}
