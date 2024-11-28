package main

import (
	"backend/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"backend/pkg/aws_utils"
	"backend/pkg/rabbitmq"
	"backend/pkg/redis"
	"backend/pkg/utils"
)

var redisClient *redis.Client
var redisCtx context.Context
var rabbitConn *amqp.Connection
var s3_bucket_name string

// Retrieve the average response time from Redis
func getAverageResponseTime() (int64, float64, error) {
	// Retrieve updated total response time and total requests
	totalResponseTime, err := redisClient.Get(redisCtx, "total_response_time").Float64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get total response time: %v", err)
	}

	totalRequests, err := redisClient.Get(redisCtx, "total_requests").Int64()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get total requests: %v", err)
	}

	if err == redis.Nil {
		// Return 0, 0 if the key doesn't exist
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, fmt.Errorf("failed to get average response time: %v", err)
	}

	// Calculate the average
	average := totalResponseTime / float64(totalRequests)

	return totalRequests, average, nil
}



func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ch, err := initRabbitMQ()
	rabbitmq_utils.FailOnError(err, "Failed to connect to RabbitMQ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Initialize Redis client
	redisClient, redisCtx = redis_utils.InitRedis(false)

	defer ch.Close()
	defer rabbitConn.Close()
	defer cancel()


	// Create a Gin router
	r := gin.Default()

	// Config CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust this to match your frontend's origin
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "index page, nothing here")
	})

	// Route to handle file upload
	r.POST("/upload", func(c *gin.Context) {
		// Get the uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get file err: %s", err.Error()))
			return
		}
		// compute the hash key for the file
		hash, err := utils.GenerateHashFromFormFile(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate hash"})
			return
		}

		// check if the file content is already processed?
		
		status, err := redisClient.HGet(redisCtx, hash, "status").Result()

		if err == nil && status != "" {
			// Respond with a success message
			c.JSON(200, gin.H{"message": "Job submitted", "jobID": hash})
			return
		}

		// save the file to local for further processing
		imagePath := "./uploads/" + file.Filename
		c.SaveUploadedFile(file, imagePath)


		
		// Create a new job
		job := &models.Job{
			ImagePath: imagePath,
			ImageDownloadURL: "",
			PDFUploadURL: "",
			JobID:     hash,
			SubmittedAt: time.Now(),
		}

		body, err := json.Marshal(job)

		err = ch.PublishWithContext(ctx,
			"",     // exchange
			"ocr-queue", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "encoding/json",
				Body:       body,
			})
	
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}

		data := map[string]interface{}{
			"response_time": 0,
			"status":        "submitted",
		}
		err = redisClient.HSet(redisCtx, job.JobID, data).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set status"})
			return
		}

		// Respond with a success message
		c.JSON(200, gin.H{"message": "Job submitted", "jobID": hash})
	})

	// Status endpoint
	r.GET("/status/:jobID", func(c *gin.Context) {
		jobID := c.Param("jobID")
		status, err := redisClient.HGet(redisCtx, jobID, "status").Result()

		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": status})
	})
	// Serve files normally
	r.Static("/uploads", "./output")

	// Download endpoint with Content-Disposition header
	r.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		filePath := "./output/" + filename

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.File(filePath)
	})


	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		healthStatus := map[string]string{
			"redis":     "ok",
			"rabbitmq":  "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		}
	
		// Check Redis
		err := redisClient.Ping(redisCtx).Err()
		if err != nil {
			healthStatus["redis"] = "unhealthy"
			healthStatus["redis_error"] = err.Error()
		}
	
		conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION"))
		defer conn.Close()
		if err != nil {
			healthStatus["rabbitmq"] = "unhealthy"
			healthStatus["rabbitmq_error"] = "connection is closed"
		}
	
		c.JSON(http.StatusOK, healthStatus)
	})

	// Endpoint to get average response time
	r.GET("/average-response-time", func(c *gin.Context) {
		totalReq, avgTime, err := getAverageResponseTime()
		if err != nil {
			log.Printf("Failed to calculate average response time: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate average response time"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"average_request_processing_time": avgTime, "total_requests": totalReq})
	})

	port := ":" + os.Getenv("MQ_ASYNC_PORT")
	r.Run(port)
}


func initRabbitMQ() (*amqp.Channel, error) {
	rabbitConn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}


	ch, err := rabbitConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"translation-queue", // name
		true,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	fmt.Println(q.Name)

	rabbitmq_utils.FailOnError(err, "Failed to declare translation queue")

	q, err = ch.QueueDeclare(
		"ocr-queue", // name
		true,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	rabbitmq_utils.FailOnError(err, "Failed to declare ocr queue")

	return ch, nil
}


func initS3() {
	aws_utils.InitS3Session(os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
}