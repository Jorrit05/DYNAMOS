package lib

import (
	"context"
	"fmt"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SharedServer struct {
	pb.UnimplementedMicroserviceServer
	pb.UnimplementedHealthServer
	pb.UnimplementedGenericServer
	ServiceName string
	Callback    func(ctx context.Context, data *pb.MicroserviceCommunication) error
}

func (s *SharedServer) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}

func (s *SharedServer) InitTracer(ctx context.Context, in *pb.ServiceName) (*emptypb.Empty, error) {

	_, err := InitTracer(in.ServiceName + "/sidecar")
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// SharedServer implementation for SendData, this is called by the previous microservice
// or by the sidecar itself if a message is received from rabbitMQ
//
// The different SendData functions are picked by either registring the 'serverInstance' or the 'sharedServer' instances in your
// gRPC server.
//
// Parameters:
//   - ctx: The context of the request
//   - data: MicroserviceCommunication messages
//
// Returns:
//   - ContinueReceiving: A boolean indicating if the sidecar should continue receiving messages
//   - error: An error if the function fails
func (s *SharedServer) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*pb.ContinueReceiving, error) {
	logger.Sugar().Debugf("Starting (to next MS) lib.SendData: %v", data.RequestMetadata.DestinationQueue)

	ctx, span, err := StartRemoteParentSpan(ctx, fmt.Sprintf("%s SendData/func:", s.ServiceName), data.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	if data.Type == "microserviceCommunication" {
		err = s.Callback(ctx, data)
		if err != nil {
			logger.Sugar().Errorf("Failed to process message: %v", err)
		}
	} else {
		logger.Sugar().Errorf("Unknown message type: %v", data.Type)
		return &pb.ContinueReceiving{ContinueReceiving: false}, fmt.Errorf("unknown message type: %s", data.Type)
	}
	return &pb.ContinueReceiving{ContinueReceiving: false}, err
}
