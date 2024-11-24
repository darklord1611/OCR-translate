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
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"backend/pkg/aws_utils"
	"mime/multipart"
	"path/filepath"
)

var redisClient *redis.Client
var redisCtx = context.Background()
var rabbitConn *amqp.Connection
var s3_bucket_name string

// Retrieve the average response time from Redis
func getAverageResponseTime() (float64, error) {
	// Retrieve updated total response time and total requests
	totalResponseTime, err := redisClient.Get(redisCtx, "total_response_time").Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to get total response time: %v", err)
	}

	totalRequests, err := redisClient.Get(redisCtx, "total_requests").Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to get total requests: %v", err)
	}

	if err == redis.Nil {
		// Return 0 if the key doesn't exist
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get average response time: %v", err)
	}

	// Calculate the average
	average := totalResponseTime / float64(totalRequests)

	return average, nil
}



func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ch, err := initRabbitMQ()
	failOnError(err, "Failed to connect to RabbitMQ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Initialize Redis client
	initRedis()

	initS3()
	s3_bucket_name = os.Getenv("AWS_BUCKET_NAME")

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

		// Generate a UUID for the jobID
		jobID := uuid.New().String()

		// Generate a new filename with the UUID
		newFileName := generateNewFileName(file, jobID)
		imagePath := "./uploads/" + newFileName
		key := "uploads/" + newFileName

		// Generate presign URLs to upload and download the image
		ImageDownloadURL, ImageUploadURL, err := aws_utils.GeneratePresignedURL(s3_bucket_name, key, 15*time.Minute)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("failed to generate download pre-signed URL: %s", err.Error()))
			return
		}

		// Generate presign URL for translation worker to upload the pdf
		out_key := "output/" + jobID + ".pdf"
		PDFUploadURL, err := aws_utils.GenerateUploadURL(s3_bucket_name, out_key, 15*time.Minute)

		// Stream the image file to S3 using the pre-signed URL
		src, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("file open err: %s", err.Error()))
			return
		}
		defer src.Close()
		err = aws_utils.UploadStream(src, ImageUploadURL)

		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("upload to S3 failed: %s", err.Error()))
			return
		}

		

		// Create a new job
		job := &models.Job{
			ImagePath: imagePath,
			ImageDownloadURL: ImageDownloadURL,
			PDFUploadURL: PDFUploadURL,
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
		c.JSON(200, gin.H{"message": "Job submitted", "jobID": jobID})
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

	r.GET("/cloud_download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		// Generate a presigned URL for downloading the file
		filePath := "output/" + filename
		url, err := aws_utils.GenerateDownloadURL(s3_bucket_name, filePath, 15*time.Minute)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate presigned URL"})
			return
		}

		log.Printf("Generated presigned URL: %s", url)

		// Return the URL to the client
		c.JSON(200, gin.H{
			"url": url,
		})
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
		avgTime, err := getAverageResponseTime()
		if err != nil {
			log.Printf("Failed to calculate average response time: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate average response time"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"average_response_time": avgTime})
	})

	port := ":" + os.Getenv("MQ_ASYNC_PORT")
	r.Run(port)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
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

	return ch, nil
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CONNECTION"), // Redis connection string
		Password: "",                            // No password set
		DB:       0,                             // Use default DB
	})

	removeAllHashKeys("*")
}


func initS3() {
	aws_utils.InitS3Session(os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))
}

// Remove all hash keys
func removeAllHashKeys(pattern string) error {
	// Find keys matching the pattern
	keys, err := redisClient.Keys(redisCtx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %v", err)
	}

	if len(keys) == 0 {
		log.Println("No keys match the given pattern")
		return nil
	}

	// Delete the keys
	deletedCount, err := redisClient.Del(redisCtx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %v", err)
	}

	log.Printf("Deleted %d keys matching the pattern '%s'", deletedCount, pattern)
	return nil
}

// Function to generate a new filename with UUID appended before the extension
func generateNewFileName(file *multipart.FileHeader, uuid string) string {

	// Extract the base name and extension
	baseName := file.Filename[:len(file.Filename)-len(filepath.Ext(file.Filename))]
	ext := filepath.Ext(file.Filename)

	// Append the UUID to the base name
	newFileName := fmt.Sprintf("%s-%s%s", baseName, uuid, ext)

	return newFileName
}

func addExtensionToFile(filename, ext string) string {
	// Get the file name without extension
	extExisting := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(extExisting)]

	// Add the new extension
	return baseName + ext
}