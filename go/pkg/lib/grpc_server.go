package lib

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

type SharedServer struct {
	pb.UnimplementedMicroserviceServer
	pb.UnimplementedHealthServer
}

func (s *SharedServer) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}
