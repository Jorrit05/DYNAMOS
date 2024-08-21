package main

import (
	"context"
	"os"
	"strconv"
	"sync"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

var (
	logger               = lib.InitLogger(logLevel)
	config               *msinit.Configuration
	COORDINATOR          = make(chan struct{})
	NR_OF_DATA_PROVIDERS = getNrOfDataProviders()
	receiveMutex         = &sync.Mutex{}
	mscommList           = []*pb.MicroserviceCommunication{}
)

func main() {
	logger.Sugar().Debugf("Starting %s service", serviceName)

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

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

		config.NextClient.SendData(ctx, msComm)

		close(config.StopMicroservice)
		return nil
	}
}

func getNrOfDataProviders() int {
	nr_of_data_providers_int := 0
	nr_of_data_providers := os.Getenv("NR_OF_DATA_PROVIDERS")
	var err error
	if nr_of_data_providers != "" {
		nr_of_data_providers_int, err = strconv.Atoi(nr_of_data_providers)
		if err != nil {
			logger.Sugar().Errorf("Error converting nr_of_data_providers to int: %v", err)
		}
	}
	return nr_of_data_providers_int
}

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, msCommList []*pb.MicroserviceCommunication) (context.Context, *pb.MicroserviceCommunication, error) {
	ctx, span := trace.StartSpan(ctx, "aggregtate/handleSqlDataRequest")
	defer span.End()
	logger.Sugar().Infof("Start %s handleSqlDataRequest", serviceName)

	if len(msCommList) < 2 {
		return ctx, msCommList[0], nil
	}

	// Coordinator ensures all services are started before further processing messages
	msCommList[0].Traces["binaryTrace"] = propagation.Binary(span.SpanContext())
	mergedMsComm := mergeData(msCommList)

	return ctx, mergedMsComm, nil
}
