package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	logger.Debug("Start msCommunication handleSqlDataRequest")
	// Look for appropiate agent
	// Convert to AMQ
	// Throw on queue
	// Graceful exit
	return nil
}
