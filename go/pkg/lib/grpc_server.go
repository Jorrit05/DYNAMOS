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
	callbacks map[string]func(ctx context.Context, data *pb.MicroserviceCommunication) error
}

func (s *SharedServer) RegisterCallback(msgType string, callback func(ctx context.Context, data *pb.MicroserviceCommunication) error) {
	if s.callbacks == nil {
		s.callbacks = make(map[string]func(ctx context.Context, data *pb.MicroserviceCommunication) error)
	}
	s.callbacks[msgType] = callback
}

func (s *SharedServer) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Debug("Starting lib.SendData")
	logger.Sugar().Debugf("data.Type: %v", data.Type)
	logger.Sugar().Debugf("data.RequestType: %v", data.RequestType)
	callback, ok := s.callbacks[data.RequestType]
	if !ok {
		logger.Warn("no callback registered for this message type")
		return nil, fmt.Errorf("no callback registered for this message type")
	}

	if err := callback(ctx, data); err != nil {
		logger.Sugar().Errorf("Callback Error: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
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
