package lib

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MbtMessage defines the structure for the message body.
type MbtMessage struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

// Queue this function is dedicated to
const (
	mbtTestingQueueName = "mbt_testing_queue"
	rabbitPort          = "5672"
	rabbitDNS           = "rabbitmq.core.svc.cluster.local"
)

func SendToTestQueue(
	ctx context.Context,
	msgType string,
	msgBody interface{}, // expected to be a proto.Message
) *emptypb.Empty {
	logger.Debug("Starting SendToTestQueue", zap.String("queue", mbtTestingQueueName))

	connectionString, err := getAMQConnectionString()
	if err != nil {
		logger.Error("Failed to get RabbitMQ connection string", zap.Error(err))
		return &emptypb.Empty{}
	}

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return &emptypb.Empty{}
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open channel", zap.Error(err))
		return &emptypb.Empty{}
	}
	defer channel.Close()

	if !queueExists(channel, mbtTestingQueueName) {
		logger.Error("Queue does not exist", zap.String("queue", mbtTestingQueueName))
		return &emptypb.Empty{}
	}

	pbMsg, ok := msgBody.(proto.Message)
	if !ok {
		logger.Error("msgBody is not a proto.Message")
		return &emptypb.Empty{}
	}

	binaryBody, err := proto.Marshal(pbMsg)
	if err != nil {
		logger.Error("Failed to marshal protobuf message", zap.Error(err))
		return &emptypb.Empty{}
	}

	payload := MbtMessage{
		Type: msgType,
		Body: base64.StdEncoding.EncodeToString(binaryBody),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal payload to JSON", zap.Error(err))
		return &emptypb.Empty{}
	}

	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        bodyBytes,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	logger.Info("Publishing message",
		zap.String("queue", mbtTestingQueueName),
		zap.String("type", msgType),
	)

	err = channel.PublishWithContext(
		ctxWithTimeout,
		"",                  // exchange
		mbtTestingQueueName, // routing key
		true,                // mandatory
		false,               // immediate
		publishing,
	)
	if err != nil {
		logger.Error("Failed to publish message",
			zap.String("queue", mbtTestingQueueName),
			zap.Error(err),
		)
	}

	logger.Debug("Finished SendToTestQueue", zap.String("queue", mbtTestingQueueName))
	return &emptypb.Empty{}
}

func getAMQConnectionString() (string, error) {
	user := os.Getenv("AMQ_USER")
	pw := os.Getenv("AMQ_PASSWORD")

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pw, rabbitDNS, rabbitPort), nil
}

func queueExists(ch *amqp.Channel, queueName string) bool {
	_, err := ch.QueueDeclarePassive(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	return err == nil
}
