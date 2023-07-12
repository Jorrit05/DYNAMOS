package main

import (
	"context"
	"errors"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	bo "github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	channel    *amqp.Channel
	conn       *amqp.Connection
	messages   <-chan amqp.Delivery
	routingKey string
)

func send(message amqp.Publishing, target string, opts ...etcd.Option) (*emptypb.Empty, error) {
	// Start with default options
	retryOpts := etcd.DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}

	// Create a returns channel
	returns := make(chan amqp.Return)
	channel.NotifyReturn(returns)

	operation := func() error {
		// Log before sending message
		logger.Sugar().Infow("Sending message: ", "My routingKey", routingKey, "exchangeName", exchangeName, "target", target)

		errChan := make(chan error)

		go func() {
			// Create a context with a timeout
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := channel.PublishWithContext(ctx, exchangeName, target, true, false, message)
			errChan <- err
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return err
			}
		case r := <-returns:
			if r.ReplyText == "NO_ROUTE" {
				logger.Sugar().Infof("No route to job yet: %v", target)
				// This will trigger a retry
				return errors.New("no route to target")
			} else {
				logger.Sugar().Errorf("Unknown reason message returned: %v", r)
				return bo.Permanent(errors.New("unknown error"))
			}
		}

		return nil
	}

	// Create a new exponential backoff
	backoff := bo.NewExponentialBackOff()
	backoff.InitialInterval = retryOpts.InitialInterval
	backoff.MaxInterval = retryOpts.MaxInterval
	backoff.MaxElapsedTime = retryOpts.MaxElapsedTime
	err := bo.Retry(operation, backoff)
	if err != nil {
		logger.Sugar().Errorf("Publish failed after %v seconds, err: %s", backoff.MaxElapsedTime, err)
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) SendRequestApproval(ctx context.Context, in *pb.RequestApproval) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal requestApproval failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "requestApproval",
	}

	return send(message, "policyEnforcer-in")
}

func (s *server) SendValidationResponse(ctx context.Context, in *pb.ValidationResponse) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal ValidationResponse failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "validationResponse",
	}

	return send(message, "orchestrator-in")
}

func (s *server) SendCompositionRequest(ctx context.Context, in *pb.CompositionRequest) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal CompositionRequest failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "compositionRequest",
	}

	return send(message, in.DestinationQueue)
}

func (s *server) SendSqlDataRequest(ctx context.Context, in *pb.SqlDataRequest) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal requestApproval failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: in.CorrelationId,
		Body:          data,
		Type:          "sqlDataRequest",
	}

	return send(message, in.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second))
}

func (s *server) SendSqlDataRequestResponse(ctx context.Context, in *pb.SqlDataRequestResponse) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal requestApproval failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: in.CorrelationId,
		Body:          data,
		Type:          "sqlDataRequestResponse",
	}

	return send(message, in.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second))
}

func (s *server) SendTest(ctx context.Context, in *pb.SqlDataRequest) (*emptypb.Empty, error) {

	// Create a returns channel
	// returns := make(chan amqp.Return)
	// channel.NotifyReturn(returns)

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: "in.CorrelationId",
		Body:          []byte("data"),
		Type:          "TestMessage",
	}
	in.DestinationQueue = "donnon"
	return send(message, in.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second))

	// Start a separate goroutine to handle returned messages
	// go func() {
	// 	for r := range returns {
	// 		logger.Sugar().Errorf("Message returned: %v", r.ReplyText)
	// 		if strings.EqualFold(r.ReplyText, "NO_ROUTE") {
	// 			logger.Sugar().Info("NO_ROUTE returned: if statement")

	// 		}
	// 		// Handle the returned message (e.g. by retrying, logging an error, etc.)
	// 	}
	// }()

	// Do other stuff
	// message := amqp.Publishing{
	// 	CorrelationId: "in.CorrelationId",
	// 	Body:          []byte("data"),
	// 	Type:          "TestMessage",
	// }

	// err := channel.PublishWithContext(context.Background(), exchangeName, "doesnotexist", true, false, message)
	// if err != nil {
	// 	logger.Sugar().Errorf("Publish failed: %s", err)
	// 	return &emptypb.Empty{}, err
	// }

	return &emptypb.Empty{}, nil
}
