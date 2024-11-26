package main

import (
	"fmt"
	"log"
	"os"
	"context"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
	"encoding/json"
	"backend/pkg/translation"
	"backend/pkg/pdf"
	"backend/models"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/cpu"
)

var margins = map[string]float64{
	"left":  30,
	"top":   50,
	"right": 30,
}

var redisClient *redis.Client
var redisCtx = context.Background()


// Update average response time in Redis
func updateAverageResponseTime(responseTime time.Duration) error {
	// Increment total response time
	err := redisClient.IncrByFloat(redisCtx, "total_response_time", responseTime.Seconds()).Err()
	if err != nil {
		return fmt.Errorf("failed to increment total response time: %v", err)
	}

	// Increment total requests
	err = redisClient.Incr(redisCtx, "total_requests").Err()
	if err != nil {
		return fmt.Errorf("failed to increment total requests: %v", err)
	}

	return nil
}

func main() {

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	conn, err := connectRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	initRedis()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	// err = channel.Qos(
	// 	5,     // prefetch count
	// 	0,     // prefetch size
	// 	false, // global
	// )
	// failOnError(err, "Failed to set QoS")

	translate_queue, err := initQueue(channel, "translation-queue")
	failOnError(err, "Failed to declare a queue")

	msgs, err := consumeMessage(channel, translate_queue.Name)
	failOnError(err, "Failed to register a consumer")

	// go monitorCPUUsage(channel)
	
	var forever chan struct{}

	go func() {
		for d := range msgs {
			var job models.Job
			err := json.Unmarshal(d.Body, &job)
			failOnError(err, "Failed to unmarshal job")

			_, err = processMessage(&job)
			if err != nil {
				log.Printf("Failed to translate: %v", err)
			}
			
			job.CompletedAt = time.Now()
        	job.ResponseTime = job.CompletedAt.Sub(job.SubmittedAt)

			updateAverageResponseTime(job.ResponseTime)

			data := map[string]interface{}{
				"response_time": job.ResponseTime.Milliseconds(), // Store as milliseconds
				"status":        "completed",
			}
			err = redisClient.HSet(redisCtx, job.JobID, data).Err()
			failOnError(err, "Failed to set response time Redis")

			log.Printf("Total processing time: %v", job.ResponseTime)

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

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CONNECTION"), // Redis connection string
		Password: "",                            // No password set
		DB:       0,                             // Use default DB
	})
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

func processMessage(job *models.Job) (string, error) {
	translatedText := translation.TranslateFilter(job.ExtractedText)
	job.TranslatedText = translatedText

	var OutFilePath string
	var err error
	if job.PDFUploadURL != "" {
		OutFilePath, err = pdf.ExportPDFtoS3(job.TranslatedText, job.JobID, margins, job.PDFUploadURL)
	} else {
		OutFilePath, err = pdf.ExportPDF(job.TranslatedText, job.JobID, margins)
	}


	if err != nil {
		return OutFilePath, fmt.Errorf("failed to generate PDF: %w", err)
	}
	
	return OutFilePath, nil
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
 


