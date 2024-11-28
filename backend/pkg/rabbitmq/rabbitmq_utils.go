package rabbitmq_utils


import (
	"fmt"
	"log"
	"os"
	amqp "github.com/rabbitmq/amqp091-go"
)



func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}




func ConnectRabbitMQ() (*amqp.Connection, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

func InitQueue(channel *amqp.Channel, queueName string) (amqp.Queue, error) {
	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return queue, fmt.Errorf("failed to declare queue: %w", err)
	}
	return queue, nil
}

func PublishMessage(channel *amqp.Channel, queueName string, messageBody []byte) error {
	err := channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "encoding/json",
			Body:        messageBody,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	log.Printf("Message published to queue %s", queueName)
	return nil
}

func ConsumeMessage(channel *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}
	return msgs, nil
}