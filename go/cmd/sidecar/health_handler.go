package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func (s *server) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}
