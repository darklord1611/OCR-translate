package main


import (
	"fmt"
	"log"
	"os"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
	"encoding/json"
	"backend/pkg/ocr"
	"backend/models"
	"github.com/joho/godotenv"
	"backend/pkg/aws_utils"
	"github.com/shirou/gopsutil/cpu"
)


func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mode := "CLIENT_POOL"

	if mode == "CLIENT_POOL" {
		ocr.Initialize()
		defer ocr.Cleanup()
	}
	
	conn, err := connectRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	ocr_queue, err := initQueue(channel, "ocr-queue")
	failOnError(err, "Failed to declare a queue")

	msgs, err := consumeMessage(channel, ocr_queue.Name)
	failOnError(err, "Failed to register a consumer")

	// go monitorCPUUsage(channel)

	var req_count int = 0

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var job models.Job
			err := json.Unmarshal(d.Body, &job)
			failOnError(err, "Failed to unmarshal job")

			err = processMessage(&job, mode)
			if err != nil {
				log.Printf("Failed to process image: %v", err)
			}

			new_msg, err := json.Marshal(job)
			failOnError(err, "Failed to marshal job")
			req_count++
			log.Printf("Processed %dth requests", req_count)
			publishMessage(channel, "translation-queue", new_msg)
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

func connectRabbitMQ() (*amqp.Connection, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
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
	log.Printf("Message published to queue %s", queueName)
	return nil
}

func consumeMessage(channel *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
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

func processMessage(job *models.Job, mode string) error {

	var text string
	var err error

	if job.ImageDownloadURL != "" {
		err = aws_utils.DownloadFile(job.ImageDownloadURL, job.ImagePath)
	}

	text, err = ocr.OCRFilter(job.ImagePath)

	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}
	job.ExtractedText = text
	return nil
}


func monitorCPUUsage(channel *amqp.Channel) {
	for {
		// Get the current CPU usage as a percentage
		usage, err := cpu.Percent(0, false)
		if err != nil {
			log.Printf("Error fetching CPU usage: %v", err)
			continue
		}

		// Adjust QoS based on CPU usage
		currentUsage := usage[0]
		var prefetchCount int
		if currentUsage < 50 {
			prefetchCount = 5 // Low CPU usage: allow up to 5 messages
		} else if currentUsage < 80 {
			prefetchCount = 2 // Medium CPU usage: allow up to 2 messages
		} else {
			prefetchCount = 1 // High CPU usage: allow only 1 message
		}

		// Set the new QoS
		err = channel.Qos(prefetchCount, 0, true)
		if err != nil {
			log.Printf("Error setting QoS: %v", err)
		}

		// Wait for a few seconds before checking again
		time.Sleep(5 * time.Second)
	}
}
