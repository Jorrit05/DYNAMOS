package lib

import (
	"context"

	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection

type serviceFunc func(message amqp.Delivery) (amqp.Publishing, error)

type connectionOptions struct {
	queueAutoDelete bool
}

type ConnectionOptions func(*connectionOptions)

func QueueAutoDelete(queueAutoDelete bool) ConnectionOptions {
	return func(options *connectionOptions) {
		options.queueAutoDelete = queueAutoDelete
	}
}

func getConnectionToRabbitMq() (*amqp.Connection, *amqp.Channel, error) {
	connectionString, err := GetAMQConnectionString()
	if err != nil {
		log.Fatalf("Failed to get an AMQ connectionString: %v", err)
	}

	var conn *amqp.Connection
	var channel *amqp.Channel

	for i := 1; i <= 7; i++ { // maximum of 7 retries
		conn, channel, err = Connect(connectionString)
		if err == nil {
			break // no error, break out of loop
		}

		log.Printf("Failed to connect to RabbitMQ: %v", err)
		time.Sleep(10 * time.Second) // wait for 10 seconds before retrying
	}

	if err != nil {
		log.Fatalf("Failed to setup proper connection to RabbitMQ after 7 attempts: %v", err)
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
func SetupConnection(serviceName string, routingKey string, startConsuming bool, opts ...ConnectionOptions) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	options := &connectionOptions{
		queueAutoDelete: false,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	conn, channel, _ := getConnectionToRabbitMq()

	err := Exchange(channel)
	if err != nil {
		log.Fatalf("Failed to create exchange: %v", err)
		return nil, nil, nil, err
	}

	queue, err := DeclareQueue(serviceName, channel, options.queueAutoDelete)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
		return nil, nil, nil, err
	}

	// Bind queue to "topic_exchange"
	// TODO: Make "topic_exchange" flexible?
	if err := channel.QueueBind(
		queue.Name,       // name
		routingKey,       // key
		"topic_exchange", // exchange
		false,            // noWait
		nil,              // args
	); err != nil {
		log.Fatalf("Queue Bind: %s", err)
		return nil, nil, nil, err
	}

	// Start listening to queue defined by environment var INPUT_QUEUE
	if startConsuming {
		messages, err := Consume(os.Getenv("INPUT_QUEUE"), channel)
		if err != nil {
			log.Fatalf("Failed to register consumer: %v", err)
			return nil, nil, nil, err
		} else {
			log.Printf("%s listening to (consumer of): %s", serviceName, os.Getenv("INPUT_QUEUE"))
		}

		return messages, conn, channel, nil
	}

	return nil, conn, channel, nil
}

func StartNewConsumer() (<-chan amqp.Delivery, *amqp.Connection) {
	conn, channel, _ := getConnectionToRabbitMq()
	var messages <-chan amqp.Delivery
	var err error
	var consumer = os.Getenv("INPUT_QUEUE")
	for i := 1; i <= 7; i++ { // maximum of 7 retries
		messages, err = Consume(consumer, channel)
		if err == nil {
			break // no error, break out of loop
		}

		log.Printf("Failed to register consumer %s, retrying... %v", consumer, err)
		time.Sleep(10 * time.Second) // wait for 10 seconds before retrying
	}

	if err != nil {
		log.Fatalf("Failed to setup proper connection to RabbitMQ after 7 attempts: %v", err)
	}

	log.Printf("Registered consumer: %s", os.Getenv("INPUT_QUEUE"))
	return messages, conn
}

func StartMessageLoop(fn serviceFunc, messages <-chan amqp.Delivery, channel *amqp.Channel, routingKey string, exchangeName string) {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}

	log.Printf("before messageloop of %s", routingKey)
	// Message loop stays alive
	for msg := range messages {
		log.Printf("StartMessageLoop: Received message: %v", string(msg.Body))

		newMsg, err := fn(msg)

		if err != nil {
			publishing := amqp.Publishing{
				Body: []byte("Error executing query: " + err.Error()),
			}
			err := channel.PublishWithContext(context.Background(), "dead-letter-exchange", routingKey, false, false, publishing)
			if err != nil {
				log.Fatalf("StartMessageLoop: Error publishing message: %v", err)
			}
		} else {
			err := channel.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, newMsg)
			if err != nil {
				log.Printf("StartMessageLoop: Error publishing message: %v", err)
			}
		}
	}
}

func Connect(connectionString string) (*amqp.Connection, *amqp.Channel, error) {
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

func DeclareQueue(name string, channel *amqp.Channel, autoDelete bool) (*amqp.Queue, error) {
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

func Close(channel *amqp.Channel) {
	channel.Close()
	conn.Close()
}

func Exchange(channel *amqp.Channel) error {
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

func Consume(queueName string, channel *amqp.Channel) (<-chan amqp.Delivery, error) {
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

func Publish(chann *amqp.Channel, routingKey string, message amqp.Publishing, exchangeName string) error {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}
	log.Printf("Publish: exchangeName: %s, routingKey: %s", exchangeName, routingKey)

	err := chann.PublishWithContext(context.Background(), exchangeName, routingKey, false, false, message)
	if err != nil {
		log.Printf("Publish: 2 %s", err)
		return err
	}
	log.Println("Publish: 3")

	return nil
}
