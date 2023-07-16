package main

import (
	"context"
	"fmt"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *server) InitRabbitMq(ctx context.Context, in *pb.InitRequest) (*emptypb.Empty, error) {
	logger.Debug("Starting InitRabbitMq")
	logger.Sugar().Infow("Received:", "Servicename", in.ServiceName, "RoutingKey", in.RoutingKey)

	var err error
	// Call the SetupConnection function and handle the message consumption inside this function
	_, conn, channel, err = setupConnection(in.ServiceName, in.RoutingKey, in.QueueAutoDelete)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *server) CreateQueue(ctx context.Context, in *pb.QueueInfo) (*emptypb.Empty, error) {
	queue, err := declareQueue(in.QueueName, channel, true)
	if err != nil {
		logger.Sugar().Fatalw("Failed to declare queue: %v", err)
		return nil, err
	}
	if err := channel.QueueBind(
		queue.Name,       // name
		in.QueueName,     // key
		"topic_exchange", // exchange
		false,            // noWait
		nil,              // args
	); err != nil {
		logger.Sugar().Fatalw("Queue Bind: %s", err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) Consume(in *pb.ConsumeRequest, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("Starting Consume")
	var err error
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
		case "validationResponse":
			if err := s.handleValidationResponse(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling validation response: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		case "requestApproval":
			if err := s.handleRequestApprovalResponse(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling validation response: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		case "compositionRequest":
			if err := s.handleCompositionRequestResponse(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling validation response: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		case "sqlDataRequest":
			if err := s.handleSqlDataRequest(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling sqlData request: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		case "microserviceCommunication":
			if err := s.handleMicroserviceCommunication(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling microserviceCommunication: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		// case "sqlDataRequestResponse":
		// 	if err := s.handleSqlDataRequestResponse(msg, stream); err != nil {
		// 		logger.Sugar().Errorf("Error handling SqlDataRequestResponse response: %v", err)
		// 		return status.Error(codes.Internal, err.Error())
		// 	}

		// Handle other message types...
		default:
			logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
			return status.Error(codes.Unknown, fmt.Sprintf("Unknown message type: %s", msg.Type))
		}
	}

	return nil
}

// func send(message amqp.Publishing, target string) (*emptypb.Empty, error) {
// 	logger.Sugar().Infow("Sending message: ", "My routingKey", routingKey, "exchangeName", exchangeName, "target", target)

// 	err := channel.PublishWithContext(context.Background(), exchangeName, target, false, false, message)
// 	if err != nil {
// 		logger.Sugar().Errorf("Publish failed: %s", err)
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}

// 	return &emptypb.Empty{}, nil
// }
