package main

import (
	"context"
	"io"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
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
		case "sqlDataRequestResponse":
			logger.Debug("Received sqlDataRequestResponse")
			sqlResult := &pb.SqlDataRequestResponse{}

			if err := grpcMsg.Body.UnmarshalTo(sqlResult); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
			}

			waitingJobMutex.Lock()
			waitingJobName, ok := waitingJobMap[sqlResult.CorrelationId]
			waitingJobMutex.Unlock()

			if ok {
				// There was still a job waiting for this response
				handleFurtherProcessing(waitingJobName, sqlResult)
				waitingJobMutex.Lock()
				delete(waitingJobMap, sqlResult.CorrelationId)
				waitingJobMutex.Unlock()
				break
			}

			mutex.Lock()
			// Look up the corresponding channel in the request map
			requestData, ok := responseMap[sqlResult.CorrelationId]
			mutex.Unlock()

			if ok {
				logger.Sugar().Info("Sending requestData to channel")
				// Send a signal on the channel to indicate that the response is ready
				requestData.response <- sqlResult
				break
			}

			ttpMutex.Lock()
			// Look up the corresponding channel in the request map
			returnAddress, ok := thirdPartyMap[sqlResult.CorrelationId]
			ttpMutex.Unlock()

			if ok {
				logger.Sugar().Infof("Sending sql response to returnAddress: %s", returnAddress)
				// Send a signal on the channel to indicate that the response is ready
				sqlResult.DestinationQueue = returnAddress

				c.SendSqlDataRequestResponse(context.Background(), sqlResult)
				break
			}
			logger.Sugar().Errorw("unknown requestData response", "CorrelationId", sqlResult.CorrelationId)

		case "sqlDataRequest":
			// Implicitly this means I am only a dataProvider
			logger.Debug("Received sqlDataRequest from Rabbit (third party)")
			sqlDataRequest := &pb.SqlDataRequest{}

			if err := grpcMsg.Body.UnmarshalTo(sqlDataRequest); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
			}

			ttpMutex.Lock()
			thirdPartyMap[sqlDataRequest.CorrelationId] = sqlDataRequest.ReturnAddress
			ttpMutex.Unlock()

			// Get the jobname of this user
			jobName, err := getJobName(sqlDataRequest.User.UserName)
			if err != nil {
				break
			}

			msChainMutex.Lock()
			msChain, ok := msChainMap[jobName]
			msChainMutex.Unlock()

			if ok {
				actualJobName, err := deployJob(msChain, jobName)
				if err != nil {
					break
				}

				sqlDataRequest.DestinationQueue = actualJobName
				sqlDataRequest.ReturnAddress = agentConfig.RoutingKey

				logger.Sugar().Debugf("Sending sqlDataRequest to sidecar hopefully: %s", sqlDataRequest.DestinationQueue)
				go c.SendSqlDataRequest(context.Background(), sqlDataRequest)
			} else {
				logger.Sugar().Warnf("unknown sqlRequest on job: %s", jobName)
			}

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
