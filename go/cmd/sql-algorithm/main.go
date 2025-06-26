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

	// TODO: get number of data providers and add to NewConfiguration function, use from sql-aggregate/main.go the nr of data stewards.
	// TODO: should add here some logic if a trusted third party is used?

	config, err := msinit.NewConfiguration(context.Background(), serviceName, grpcAddr, COORDINATOR, messageHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	// Wait here until the message arrives in the messageHandler
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

		switch msComm.RequestType {
		case "sqlDataRequest":
			err := handleSqlDataRequest(ctx, msComm)
			if err != nil {
				logger.Sugar().Errorf("Failed to process %s message: %v", msComm.RequestType, err)
			}

		default:
			logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
		}

		// TODO: the third party needs to handle more incoming data before it is stopped, it needs to handle it from the second data provider as well.
		// so, it should not be stopped just yet, add custom logic to allow for receiving more.
		// TODO: I think SendData can be used, but the next service should also wait then for the next response I think.

		config.NextClient.SendData(ctx, msComm)

		// Stop the microservice gracefully
		close(config.StopMicroservice)
		return nil
	}
}