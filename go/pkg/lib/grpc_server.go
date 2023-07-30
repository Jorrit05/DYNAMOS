package lib

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SharedServer struct {
	pb.UnimplementedMicroserviceServer
	pb.UnimplementedHealthServer
	pb.UnimplementedGenericServer
	callbacks map[string]func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)
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

func (s *SharedServer) RegisterCallback(msgType string, callback func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)) {
	if s.callbacks == nil {
		s.callbacks = make(map[string]func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error))
	}
	s.callbacks[msgType] = callback
}

// TODO: This go function is mostly to get an accurate feel for data transfer speeds.
// It's probably better to just remove the Go func in the long run
func (s *SharedServer) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Sugar().Debugf("Starting lib.SendData: %v", data.RequestMetadata.DestinationQueue)

	ctx, span, err := StartRemoteParentSpan(ctx, "sidecar SendData/func:", data.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	span.End()
	// logger.Sugar().Debugf("data.Type: %v", data.Type)
	// logger.Sugar().Debugf("data.RequestType: %v", data.RequestType)
	callback, ok := s.callbacks[data.Type]
	if !ok {
		logger.Warn("no callback registered for this message type")
		// span.End()
		return &emptypb.Empty{}, nil
	}

	// go func(data *pb.MicroserviceCommunication, callback func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)) {
	logger.Sugar().Debugf("In go routine SendData:")
	// span.End()
	if _, err := callback(ctx, data); err != nil {
		logger.Sugar().Errorf("Callback Error: %v", err)
		return &emptypb.Empty{}, nil
	}
	// }(data, callback)

	return &emptypb.Empty{}, nil
}
