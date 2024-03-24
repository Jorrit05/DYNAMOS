package main

import (
	"context"
	"fmt"
	"os"

	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func getConnectionToRabbitMq() (*amqp.Connection, *amqp.Channel, error) {
	logger.Debug("Start getConnectionToRabbitMq")
	connectionString, err := getAMQConnectionString()

	if err != nil {
		logger.Sugar().Fatalw("Failed to get an AMQ connectionString: %v", err)
	}

	var conn *amqp.Connection
	var channel *amqp.Channel
	conn, channel, err = connect(connectionString)

	// for i := 1; i <= 7; i++ { // maximum of 7 retries
	// 	conn, channel, err = connect(connectionString)
	// 	if err == nil {
	// 		break // no error, break out of loop
	// 	}

	// 	logger.Sugar().Infof("Failed to connect to RabbitMQ: %v", err)
	// 	time.Sleep(10 * time.Second) // wait for 10 seconds before retrying
	// }

	if err != nil {
		logger.Sugar().Fatalw("Failed to setup proper connection to RabbitMQ after 7 attempts: %v", err)
	}
	return conn, channel, nil
}

// SetupConnection establishes a connection to RabbitMQ and sets up a topic exchange, queue, and consumer to listen to
// messages with the specified routing key.
// It returns a channel to receive delivery messages, the AMQP connection, and channel objects, and an error if any occurs during the setup.
// The connection string and the routing key are passed as arguments.
// The service name in format '<name>_service' is used to declare the queue.
//
// The routingKey service.<name> is used when binding the queue to the exchange, the exchange will publish messages to all queues that match the routingkey pattern
func setupConnection(queueName string, routingKey string, queueAutoDelete bool) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {

	conn, channel, _ := getConnectionToRabbitMq()

	err := exchange(channel)
	if err != nil {
		logger.Sugar().Fatalw("Failed to create exchange: %v", err)
		return nil, nil, nil, err
	}

	queue, err := declareQueue(queueName, channel, queueAutoDelete)
	if err != nil {
		logger.Sugar().Fatalw("Failed to declare queue: %v", err)
		return nil, nil, nil, err
	}

	// Bind queue to "topic_exchange"
	// TODO: Make "topic_exchange" flexible?

	// We are going to assume the queue has been created by the composition request
	for i := 1; i <= 7; i++ { // maximum of 7 retries
		err := channel.QueueBind(
			queue.Name,       // name
			routingKey,       // key
			"topic_exchange", // exchange
			false,            // noWait
			nil,              // args
		)
		if err == nil {
			break // no error, break out of loop
		}

		if i == 7 {
			logger.Sugar().Fatalw("Queue Bind: %s", err)
			return nil, nil, nil, err
		}
		// if err != nil {
		// 	logger.Sugar().Fatalw("Queue Bind: %s", err)
		// 	return nil, nil, nil, err
		// }

		logger.Sugar().Infof("Failed to connect to QueueBind: %v", err)
		time.Sleep(2 * time.Second)
	}
	return nil, conn, channel, nil
}

func connect(connectionString string) (*amqp.Connection, *amqp.Channel, error) {
	var err error
	conn, err = amqp.Dial(connectionString)
	if err != nil {
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	return conn, channel, nil
}

func declareQueue(name string, channel *amqp.Channel, autoDelete bool) (*amqp.Queue, error) {
	queue, err := channel.QueueDeclare(
		name,       // name
		true,       // durable
		autoDelete, // delete when unused
		false,      // exclusive
		false,      // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "dead-letter-exchange",
		},
	)
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func exchange(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		"topic_exchange",
		"topic",
		true,  // durable
		false, // auto delete
		false, // internal
		false, // no-wait
		nil);  // arguments
	err != nil {
		return err
	}
	return nil
}

func close_channel(channel *amqp.Channel) {
	channel.Close()
	conn.Close()
}

func getAMQConnectionString() (string, error) {
	user := os.Getenv("AMQ_USER")
	pw := os.Getenv("AMQ_PASSWORD")

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pw, rabbitDNS, rabbitPort), nil
}

// ------------------------------------------------------------------------------------------------
// OLD:
// ------------------------------------------------------------------------------------------------

func consume(queueName string, channel *amqp.Channel) (<-chan amqp.Delivery, error) {
	messages, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func publish(chann *amqp.Channel, routingKey string, message amqp.Publishing, exchangeName string) error {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}
	logger.Sugar().Infof("Publish: exchangeName: %s, routingKey: %s", exchangeName, routingKey)

	err := chann.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, message)
	if err != nil {
		logger.Sugar().Infof("Publish: 2 %s", err)
		return err
	}
	logger.Info("Publish: 3")

	return nil
}

func startNewConsumer() (<-chan amqp.Delivery, *amqp.Connection) {
	conn, channel, _ := getConnectionToRabbitMq()
	var messages <-chan amqp.Delivery
	var err error
	var consumer = os.Getenv("INPUT_QUEUE")
	for i := 1; i <= 7; i++ { // maximum of 7 retries
		messages, err = consume(consumer, channel)
		if err == nil {
			break // no error, break out of loop
		}

		logger.Sugar().Infof("Failed to register consumer %s, retrying... %v", consumer, err)
		time.Sleep(10 * time.Second) // wait for 10 seconds before retrying
	}

	if err != nil {
		logger.Sugar().Fatalw("Failed to setup proper connection to RabbitMQ after 7 attempts: %v", err)
	}

	logger.Sugar().Infof("Registered consumer: %s", os.Getenv("INPUT_QUEUE"))
	return messages, conn
}

type serviceFunc func(message amqp.Delivery) (amqp.Publishing, error)

func startMessageLoop(fn serviceFunc, messages <-chan amqp.Delivery, channel *amqp.Channel, routingKey string, exchangeName string) {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}

	logger.Sugar().Infof("before messageloop of %s", routingKey)
	// Message loop stays alive
	for msg := range messages {
		logger.Sugar().Infof("StartMessageLoop: Received message: %v", string(msg.Body))

		newMsg, err := fn(msg)

		if err != nil {
			publishing := amqp.Publishing{
				Body: []byte("Error executing query: " + err.Error()),
			}
			err := channel.PublishWithContext(context.Background(), "dead-letter-exchange", routingKey, false, false, publishing)
			if err != nil {
				logger.Sugar().Fatalw("StartMessageLoop: Error publishing message: %v", err)
			}
		} else {
			err := channel.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, newMsg)
			if err != nil {
				logger.Sugar().Infof("StartMessageLoop: Error publishing message: %v", err)
			}
		}
	}
}
