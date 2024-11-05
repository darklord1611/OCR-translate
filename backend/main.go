package main

import (
	"backend/models"
	"backend/pkg/ocr"
	"backend/pkg/pdf"
	"backend/pkg/translation"
	"backend/pkg/segmentation"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const numWorkers = 4

var jobQueue = make(chan models.Job, 100) // Job queue channel with buffer size 100
var jobStatusMap = make(map[string]string)
var jobStatusMutex = &sync.Mutex{}

func main() {

	// Initialize the Tesseract client
	ocr.Initialize()
	defer ocr.Cleanup() // Ensure the client is closed when the server shuts down

	// Start workers to process jobs concurrently
	for i := 0; i < numWorkers; i++ {
		go worker(i, jobQueue)
	}

	// Create a Gin router
	r := gin.Default()

	// Config CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Adjust this to match your frontend's origin
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

		job := models.Job{
			ImagePath: imagePath,
			JobID:     jobID,
		}

		jobQueue <- job

		// Initialize job status to "pending"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "pending"
		jobStatusMutex.Unlock()

		// Respond with a success message
		c.JSON(200, gin.H{"message": "Job submitted", "jobID": jobID})
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

	// Start the server on port 8080
	r.Run(":8080")
}

func worker(id int, jobs <-chan models.Job) {
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
		result, err := pdf.ExportPDF(translatedText, job.JobID)
		if err != nil {
			log.Printf("Job %s failed", id, job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			continue
		}

		job.OutFilePath = result

		// Update job status to "completed"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "completed"
		jobStatusMutex.Unlock()

		log.Printf("Worker %d finished job %s", id, job.JobID)
		elapsed_time := time.Since(start_time)
		log.Printf("Request took %v\n", elapsed_time)
	}
}
