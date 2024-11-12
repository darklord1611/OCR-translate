package main

import (
	"fmt"
	"log"
	"os"
	amqp "github.com/rabbitmq/amqp091-go"
	"encoding/json"
	"backend/pkg/translation"
	"backend/pkg/pdf"
	"backend/models"
	"github.com/joho/godotenv"
)

var margins = map[string]float64{
	"left":  30,
	"top":   50,
	"right": 30,
}

func main() {

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	conn, err := connectRabbitMQ(os.Getenv("RABBITMQ_CONNECTION"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	rdb := initRedis()
	redis_ctx := context.Background()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	translate_queue, err := initQueue(channel, "translation-queue")
	failOnError(err, "Failed to declare a queue")

	msgs, err := consumeMessage(channel, translate_queue.Name)
	failOnError(err, "Failed to register a consumer")


	var forever chan struct{}

	go func() {
		for d := range msgs {
			var job models.Job
			err := json.Unmarshal(d.Body, &job)
			failOnError(err, "Failed to unmarshal job")

			filePath, err := processMessage(&job)
			if err != nil {
				log.Printf("Failed to translate: %v", err)
			}
			rdb.Set(redis_ctx, job.JobID, "completed", 0)
			log.Printf("OutFilePath: %s", filePath)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}



func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func connectRabbitMQ(RABBITMQ_CONNECTION string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(RABBITMQ_CONNECTION)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_CONNECTION"),
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	return rdb
}


func initQueue(channel *amqp.Channel, queueName string) (amqp.Queue, error) {
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

func publishMessage(channel *amqp.Channel, queueName string, messageBody []byte) error {
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
	log.Printf("Message published to queue %s: %s", queueName, messageBody)
	return nil
}

func consumeMessage(channel *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
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

func processMessage(job *models.Job) (string, error) {
	translatedText := translation.TranslateFilter(job.ExtractedText)
	job.TranslatedText = translatedText

	OutFilePath, err := pdf.ExportPDF(job.TranslatedText, job.JobID, margins)
	if err != nil {
		return OutFilePath, fmt.Errorf("failed to generate PDF: %w", err)
	}
	
	return OutFilePath, nil
}


