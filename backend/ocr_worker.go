package main


import (
	"fmt"
	"log"
	"encoding/json"
	"backend/pkg/ocr"
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

	mode := "CLIENT_POOL"

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

	ocr_queue, err := rabbitmq_utils.InitQueue(channel, "ocr-queue")
	rabbitmq_utils.FailOnError(err, "Failed to declare a queue")

	msgs, err := rabbitmq_utils.ConsumeMessage(channel, ocr_queue.Name)
	rabbitmq_utils.FailOnError(err, "Failed to register a consumer")

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	rabbitmq_utils.FailOnError(err, "Failed to set QoS")

	var req_count int = 0

	var forever chan struct{}

	go func() {
		for d := range msgs {
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
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}


func processMessage(job *models.Job, mode string) error {

	var text string
	var err error

	if job.ImageDownloadURL != "" {
		err = aws_utils.DownloadFile(job.ImageDownloadURL, job.ImagePath)
	}

	if mode == "CLIENT_POOL" {
		text, err = ocr.OCRFilter(job.ImagePath)
	} else {
		text, err = ocr.OneShotOCR(job.ImagePath)
	}

	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}
	job.ExtractedText = text
	return nil
}

