package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *serverInstance) InitRabbitMq(ctx context.Context, in *pb.InitRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("Received:", "Servicename", in.ServiceName, "RoutingKey", in.RoutingKey)

	// Call the SetupConnection function and handle the message consumption inside this function
	_, conn, channel, err := setupConnection(in.ServiceName, in.RoutingKey, in.QueueAutoDelete)
	s.channel = channel
	s.conn = conn

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *serverInstance) CreateQueue(ctx context.Context, in *pb.QueueInfo) (*emptypb.Empty, error) {
	queue, err := declareQueue(in.QueueName, s.channel, in.AutoDelete)
	if err != nil {
		logger.Sugar().Fatalw("Failed to declare queue: %v", err)
		return nil, err
	}
	if err := s.channel.QueueBind(
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

func (s *serverInstance) DeleteQueue(ctx context.Context, in *pb.QueueInfo) (*emptypb.Empty, error) {
	logger.Sugar().Debugf("Delete Queue: %s", in.QueueName)

	purgedMessages, err := s.channel.QueueDelete(in.QueueName, false, false, false)
	if err != nil {
		logger.Sugar().Warnf("Error deleting queue: %v", err)
		return &emptypb.Empty{}, err
	}

	logger.Sugar().Infof("Deleted queue %s, purged %d, messages", in.QueueName, purgedMessages)
	return &emptypb.Empty{}, nil
}

func (s *serverInstance) Consume(in *pb.ConsumeRequest, stream pb.SideCar_ConsumeServer) error {
	messages, err := s.channel.Consume(
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
				logger.Sugar().Errorf("Error handling requestApproval response: %v", err)
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
		case "policyUpdate":
			if err := s.handlePolicyUpdate(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling policyUpdate: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		case "requestApprovalResponse":
			if err := s.handleRequestApprovalToApiResponse(msg, stream); err != nil {
				logger.Sugar().Errorf("Error handling requestApprovalResponse: %v", err)
				return status.Error(codes.Internal, err.Error())
			}
		// Handle other message types...
		default:
			logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
			return status.Error(codes.Unknown, fmt.Sprintf("Unknown message type: %s", msg.Type))
		}
	}
	s.channel.Close()
	return nil
}

func (s *serverInstance) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*pb.ContinueReceiving, error) {
	logger.Sugar().Debugf("Starting lib.SendData: %v", data.RequestMetadata.DestinationQueue)

	ctx, span, err := lib.StartRemoteParentSpan(ctx, "sidecar SendData/func:", data.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	if _, err := SendDataThroughAMQ(ctx, data, s); err != nil {
		logger.Sugar().Errorf("Callback Error: %v", err)
		return &pb.ContinueReceiving{ContinueReceiving: false}, nil
	}

	return &pb.ContinueReceiving{ContinueReceiving: false}, nil
}
