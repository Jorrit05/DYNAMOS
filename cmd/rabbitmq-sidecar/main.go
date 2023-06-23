package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	port       = flag.Int("port", 3005, "The server port")
	logger     = lib.InitLogger(logLevel)
	channel    *amqp.Channel
	conn       *amqp.Connection
	messages   <-chan amqp.Delivery
	routingKey string
)

type server struct {
	pb.UnimplementedSideCarServer
}

func (s *server) StartService(ctx context.Context, in *pb.ServiceRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("Received:", "Servicename", in.ServiceName, "RoutingKey", in.RoutingKey)

	var err error
	// Call the SetupConnection function and handle the message consumption inside this function
	_, conn, channel, err = setupConnection(in.ServiceName, in.RoutingKey, in.QueueAutoDelete)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// func (s *server) Consume(ctx context.Context, in *pb.ConsumeRequest) (*emptypb.Empty, error) {
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
		// Handle other message types...
		default:
			logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
			return status.Error(codes.Unknown, fmt.Sprintf("Unknown message type: %s", msg.Type))
		}
	}

	return status.Error(codes.Internal, err.Error())
}

func send(message amqp.Publishing, target string) (*emptypb.Empty, error) {
	logger.Sugar().Infow("Sending message: ", "My routingKey", routingKey, "exchangeName", exchangeName, "target", target)

	err := channel.PublishWithContext(context.Background(), exchangeName, target, false, false, message)
	if err != nil {
		logger.Sugar().Errorf("Publish failed: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.Sugar().Info("2")
	return &emptypb.Empty{}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Sugar().Fatalw("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSideCarServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}
}
