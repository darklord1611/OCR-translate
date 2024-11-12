package main

import (
	"backend/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"os"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	totalResponseTime time.Duration
	totalJobs         int
	averageMutex      = &sync.Mutex{}
)

// Function to update average response time
func updateAverageResponseTime(responseTime time.Duration) {
	averageMutex.Lock()
	defer averageMutex.Unlock()

	totalResponseTime += responseTime
	totalJobs++
}

// Function to retrieve the average response time
func getAverageResponseTime() time.Duration {
	averageMutex.Lock()
	defer averageMutex.Unlock()

	if totalJobs == 0 {
		return 0
	}
	return totalResponseTime / time.Duration(totalJobs)
}



func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, ch, err := initRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	redis_ctx := context.Background()

	// Initialize Redis client
	rdb := initRedis()


	defer ch.Close()
	defer conn.Close()
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
		imagePath := "./uploads/" + file.Filename
		// Save the file to a specific location
		err = c.SaveUploadedFile(file, imagePath)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("save file err: %s", err.Error()))
			return
		}

		// Generate a UUID for the jobID
		jobID := uuid.New().String()

		// pipeline

		job := &models.Job{
			ImagePath: imagePath,
			JobID:     jobID,
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

		err = rdb.Set(redis_ctx, job.JobID, "submitted", 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set status"})
			return
		}

		// Respond with a success message
		c.JSON(200, gin.H{"message": "Job submitted", "jobID": jobID})
	})

	// Status endpoint
	r.GET("/status/:jobID", func(c *gin.Context) {
		jobID := c.Param("jobID")
		status, err := rdb.Get(redis_ctx, jobID).Result()

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

	// Endpoint to get average response time
	// r.GET("/average-response-time", func(c *gin.Context) {
	// 	avgTime := getAverageResponseTime()
	// 	c.JSON(http.StatusOK, gin.H{"average_response_time": avgTime.Seconds()})
	// })

	// Start the server on port 8080
	r.Run(":8082")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}


func initRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
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

	failOnError(err, "Failed to declare translation queue")

	q, err = ch.QueueDeclare(
		"ocr-queue", // name
		true,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare ocr queue")

	return conn, ch, nil
}

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_CONNECTION"),
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	return rdb
}