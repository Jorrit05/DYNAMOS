package main

import (
	"context"
	"os"
	"strconv"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger               = lib.InitLogger(logLevel)
	COORDINATOR          = make(chan struct{})
	NR_OF_DATA_PROVIDERS = getNrOfDataProviders()
	msCommList           = []*pb.MicroserviceCommunication{}
)

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

		msCommList = append(msCommList, msComm)
		logger.Sugar().Infof("msCommList: %v", msCommList)
		logger.Sugar().Infof("amount of data providers %v", NR_OF_DATA_PROVIDERS)
		logger.Sugar().Infof("Length of msCommList %v", len(msCommList))

		// If NR_OF_DATA_PROVIDERS == 0 aggregate won't actually function and pass on the message.
		// This can happen at this moment if the aggregate flag is set to True, but it is not allowed by policy.
		if len(msCommList) == NR_OF_DATA_PROVIDERS || NR_OF_DATA_PROVIDERS == 0 {
			logger.Sugar().Debugf(msCommList[0].Data.String())
			// All messages have arrived
			logger.Sugar().Infof("All messages have arrived, %v", len(msCommList))

			switch msComm.RequestType {
			case "sqlDataRequest":
				ctx, msComm, err = handleSqlDataRequest(ctx, msCommList)
				if err != nil {
					logger.Sugar().Errorf("Failed to process %s message: %v", msComm.RequestType, err)
				}

			default:
				logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
			}
		} else {
			logger.Sugar().Infof("Waiting for %v messages", NR_OF_DATA_PROVIDERS-len(msCommList))
			return nil
		}

		config.NextClient.SendData(ctx, msComm)

		close(config.StopMicroservice)
		return nil
	}
}
