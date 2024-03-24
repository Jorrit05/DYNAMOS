package main

import (
	"fmt"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) ChainConsume(in *pb.ConsumeRequest, stream pb.SideCar_ChainConsumeServer) error {
	var err error
	logger.Sugar().Infow("Received:", "QueueName", in.QueueName, "AutoAck", in.AutoAck)
	messages, err = channel.Consume(
		in.QueueName, // queue
		"",           // consumer
		in.AutoAck,   // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	logger.Sugar().Infof("Started consuming from %s", in.QueueName)

	for msg := range messages {
		logger.Sugar().Debugw("switchin: ", "msg,Type", msg.Type, "Port:", port)
		switch msg.Type {

		case "microserviceCommunication":
			if err := s.handleMicroserviceCommunication(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling microserviceCommunication: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
			logger.Sugar().Debug("Handled s.handleMicroserviceCommunication")

		// Handle other message types...
		default:
			logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
			return status.
				Error(codes.Unknown, fmt.Sprintf("Unknown message type: %s", msg.Type))
		}

	}
	logger.Sugar().Debug("returning nil")
	return nil
}
