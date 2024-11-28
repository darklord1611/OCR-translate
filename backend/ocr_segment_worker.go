package main


import (
	"fmt"
	"log"
	"time"
	"encoding/json"
	"backend/pkg/ocr"
	"backend/pkg/segmentation"
	"backend/models"
	"github.com/joho/godotenv"
	"backend/pkg/aws_utils"
	"backend/pkg/rabbitmq"
)


func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mode := "SPLIT_IMAGE"

	if mode == "CLIENT_POOL" {
		ocr.Initialize()
		defer ocr.Cleanup()
	}
	
	conn, err := rabbitmq_utils.ConnectRabbitMQ()
	rabbitmq_utils.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	rabbitmq_utils.FailOnError(err, "Failed to open a channel")
	defer channel.Close()

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	rabbitmq_utils.FailOnError(err, "Failed to set QoS")

	ocr_queue, err := rabbitmq_utils.InitQueue(channel, "ocr-queue")
	rabbitmq_utils.FailOnError(err, "Failed to declare a queue")

	msgs, err := rabbitmq_utils.ConsumeMessage(channel, ocr_queue.Name)
	rabbitmq_utils.FailOnError(err, "Failed to register a consumer")

	var req_count int = 0

	var forever chan struct{}

	go func() {
		for d := range msgs {
			start_time := time.Now()
			var job models.Job
			err := json.Unmarshal(d.Body, &job)
			rabbitmq_utils.FailOnError(err, "Failed to unmarshal job")

			err = processMessage(&job, mode)
			if err != nil {
				log.Printf("Failed to process image: %v", err)
			}

			new_msg, err := json.Marshal(job)
			rabbitmq_utils.FailOnError(err, "Failed to marshal job")
			req_count++
			log.Printf("Processed %dth requests", req_count)
			rabbitmq_utils.PublishMessage(channel, "translation-queue", new_msg)
			log.Printf("OCR job completed in %v", time.Since(start_time))
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}


func processMessage(job *models.Job, mode string) error {

	var err error
	var text string
	if job.ImageDownloadURL == "" {
		err = aws_utils.DownloadFile(job.ImageDownloadURL, job.ImagePath)
	}

	segmentPaths := segmentation.SplitImage(job.ImagePath, job.JobID)
	text, err = ocr.OCRFilterConcurrent(segmentPaths)

	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}
	job.ExtractedText = text
	return nil
}


