package main

import (
	"fmt"
	"log"
	"context"
	"time"
	"encoding/json"
	"backend/pkg/translation"
	"backend/pkg/pdf"
	"backend/models"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"backend/pkg/rabbitmq"
	"backend/pkg/redis"
)

var margins = map[string]float64{
	"left":  30,
	"top":   50,
	"right": 30,
}

var redisClient *redis.ClusterClient
var redisCtx context.Context


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
	
	conn, err := rabbitmq_utils.ConnectRabbitMQ()
	rabbitmq_utils.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	redisClient, redisCtx = redis_utils.InitRedisCluster(false)

	channel, err := conn.Channel()
	rabbitmq_utils.FailOnError(err, "Failed to open a channel")
	defer channel.Close()

	// err = channel.Qos(
	// 	5,     // prefetch count
	// 	0,     // prefetch size
	// 	false, // global
	// )
	// rabbitmq_utils.FailOnError(err, "Failed to set QoS")

	translate_queue, err := rabbitmq_utils.InitQueue(channel, "translation-queue")
	rabbitmq_utils.FailOnError(err, "Failed to declare a queue")

	msgs, err := rabbitmq_utils.ConsumeMessage(channel, translate_queue.Name)
	rabbitmq_utils.FailOnError(err, "Failed to register a consumer")

	// go monitorCPUUsage(channel)
	
	var forever chan struct{}

	go func() {
		for d := range msgs {
			var job models.Job
			err := json.Unmarshal(d.Body, &job)
			rabbitmq_utils.FailOnError(err, "Failed to unmarshal job")

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
			rabbitmq_utils.FailOnError(err, "Failed to set response time Redis")

			log.Printf("Total processing time: %v", job.ResponseTime)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
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

 


