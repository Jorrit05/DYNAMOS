package main

import (
	"context"
	"os"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger      = lib.InitLogger(logLevel)
	COORDINATOR = make(chan struct{})
)

func main() {
	logger.Sugar().Debugf("Starting %s service", serviceName)

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	// Initialize the configuration for the microservice
	config, err := msinit.NewConfiguration(context.Background(), serviceName, grpcAddr, COORDINATOR, messageHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	// Wait here until the StopMicroservice channel is closed by the messageHandler
	<-config.StopMicroservice

	config.SafeExit(oce, serviceName)
	os.Exit(0)
}


func messageHandler(config *msinit.Configuration) func(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	return func(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: messageHandler", msComm.Traces)
		if err != nil {
			logger.Sugar().Warnf("Error starting span: %v", err)
		}
		defer span.End()

		// Wait till all services and connections have started
		<-COORDINATOR

		// TODO: here wait until all messages are received before further processing. Edit the handleSqlDataRequest function to
		// allow multiple messages to be processed at once. Then can do SendData to the next service with the merged and processed data.

		switch msComm.RequestType {
		case "sqlDataRequest":
			err := handleSqlDataRequest(ctx, msComm)
			if err != nil {
				logger.Sugar().Errorf("Failed to process %s message: %v", msComm.RequestType, err)
			}

		default:
			logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
		}

		// Send the data to the next microservice
		config.NextClient.SendData(ctx, msComm)

		// Close the channel (i.e., tell the waiting routine that processing is done)
		close(config.StopMicroservice)
		return nil
	}
}