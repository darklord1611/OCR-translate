package main

import (
	"backend/models"
	"backend/pkg/ocr"
	"backend/pkg/pdf"
	"backend/pkg/segmentation"
	"backend/pkg/translation"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	totalResponseTime time.Duration
	totalJobs         int
	averageMutex      = &sync.Mutex{}
)

var margins = map[string]float64{
	"left":  30,
	"top":   50,
	"right": 30}

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


var jobStatusMap = make(map[string]string)
var jobStatusMutex = &sync.Mutex{}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the Tesseract client
	ocr.Initialize()
	defer ocr.Cleanup() // Ensure the client is closed when the server shuts down

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

		// Initialize job status to "pending"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "pending"
		jobStatusMutex.Unlock()

		// process immediately

		originalText, err := ocr.OCRFilter(imagePath)
		if err != nil {
			log.Printf("Job %s failed", job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		}

		translatedText := translation.TranslateFilter(originalText)
		result, err := pdf.ExportPDF(translatedText, job.JobID, margins)
		if err != nil {
			log.Printf("Job %s failed", job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		}

		job.OutFilePath = result
		job.CompletedAt = time.Now()
		job.ResponseTime = job.CompletedAt.Sub(job.SubmittedAt)
		// Update average response time
		updateAverageResponseTime(job.ResponseTime)

		filename := job.JobID + ".pdf"
		// Respond with a success message
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.File(result)
	})

	// Status endpoint
	r.GET("/status/:jobID", func(c *gin.Context) {
		jobID := c.Param("jobID")
		jobStatusMutex.Lock()
		status, exists := jobStatusMap[jobID]
		jobStatusMutex.Unlock()
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
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
	r.GET("/average-response-time", func(c *gin.Context) {
		avgTime := getAverageResponseTime()
		c.JSON(http.StatusOK, gin.H{"average_response_time": avgTime.Seconds()})
	})

	port := ":" + os.Getenv("SYNC_PORT")
	// Start the server on port 8080
	r.Run(port)
}

func worker(id int, jobs <-chan *models.Job) {
	for job := range jobs {
		start_time := time.Now()
		log.Printf("Worker %d started job %s", id, job.JobID)

		// Update job status to "in-progress"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "in-progress"
		jobStatusMutex.Unlock()

		splitTime := time.Now()
		segmentPaths := segmentation.SplitImage(job.ImagePath, job.JobID)
		log.Printf("Image Spliting took %v\n", time.Since(splitTime))
		// Perform the OCR, translation, and PDF generation here
		OCRTime := time.Now()
		originalText, err := ocr.OCRFilterConcurrent(segmentPaths)
		if err != nil {
			log.Printf("Job %s failed", id, job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			continue
		}

		log.Printf("OCR took %v\n", time.Since(OCRTime))
		TranslationTime := time.Now()
		translatedText := translation.TranslateFilter(originalText)
		log.Printf("Translation took %v\n", time.Since(TranslationTime))
		margins := map[string]float64{
			"left":  30,
			"top":   50,
			"right": 30}
		result, err := pdf.ExportPDF(translatedText, job.JobID, margins)
		if err != nil {
			log.Printf("Job %s failed", id, job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			continue
		}

		elapsed_time := time.Since(start_time)
		log.Printf("PDF generation took %v\n", elapsed_time)

		// Update job status to "completed"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "completed"
		jobStatusMutex.Unlock()

		job.OutFilePath = result
		job.CompletedAt = time.Now()
		job.ResponseTime = job.CompletedAt.Sub(job.SubmittedAt)
		// Update average response time
		updateAverageResponseTime(job.ResponseTime)

		log.Printf("Worker %d finished job %s", id, job.JobID)
	}
}
