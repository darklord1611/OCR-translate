package main

import (
	"backend/models"
	"backend/pkg/ocr"
	"backend/pkg/pdf"
	"backend/pkg/translation"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const numWorkers = 5

var ocrQueue = make(chan models.Job, 100)        // OCR queue channel
var translationQueue = make(chan models.Job, 100) // Translation queue channel
var pdfQueue = make(chan models.Job, 100)         // PDF queue channel

var jobStatusMap = make(map[string]string)
var jobStatusMutex = &sync.Mutex{}

func main() {

	// Initialize the Tesseract client
	ocr.Initialize()
	defer ocr.Cleanup() // Ensure the client is closed when the server shuts down

	// Start workers for each queue
	for i := 0; i < numWorkers * 2; i++ {
		go ocrWorker(i, ocrQueue)
	}

	for i := 0; i < numWorkers; i++ {
		go translationWorker(i, translationQueue)
		go pdfWorker(i, pdfQueue)
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
			print("Error")
			c.String(http.StatusBadRequest, fmt.Sprintf("get file err: %s", err.Error()))
			return
		}
		imagePath := "./uploads/" + file.Filename
		// Save the file to a specific location
		err = c.SaveUploadedFile(file, imagePath)
		if err != nil {
			print("Error")

			c.String(http.StatusInternalServerError, fmt.Sprintf("save file err: %s", err.Error()))
			return
		}

		// Generate a UUID for the jobID
		jobID := uuid.New().String()

		// Initialize job
		job := models.Job{
			ImagePath: imagePath,
			JobID:     jobID,
		}

		// Initialize job status to "pending"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "pending"
		jobStatusMutex.Unlock()

		// Enqueue the job in the OCR queue
		ocrQueue <- job

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
	r.Run(":8081")
}

// Worker function for OCR processing
func ocrWorker(id int, jobs <-chan models.Job) {
	for job := range jobs {
		start_time := time.Now()
		log.Printf("OCR Worker %d started job %s", id, job.JobID)

		// Update job status to "OCR in-progress"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "OCR in-progress"
		jobStatusMutex.Unlock()

		// Perform OCR
		originalText, err := ocr.OCRFilter(job.ImagePath)
		if err != nil {
			log.Printf("Job %s failed during OCR", job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			continue
		}
		elapsed_time := time.Since(start_time)
		log.Printf("OCR took %v\n", elapsed_time)

		job.ExtractedText = originalText
		translationQueue <- job // Pass job to the translation queue
	}
}

// Worker function for Translation processing
func translationWorker(id int, jobs <-chan models.Job) {
	for job := range jobs {
		start_time := time.Now()
		log.Printf("Translation Worker %d started job %s", id, job.JobID)

		// Update job status to "Translation in-progress"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "Translation in-progress"
		jobStatusMutex.Unlock()

		// Perform translation
		translatedText := translation.TranslateFilter(job.ExtractedText)
		job.TranslatedText = translatedText

		elapsed_time := time.Since(start_time)
		log.Printf("Translation took %v\n", elapsed_time)

		pdfQueue <- job // Pass job to the PDF generation queue
	}
}

// Worker function for PDF generation
func pdfWorker(id int, jobs <-chan models.Job) {
	for job := range jobs {
		start_time := time.Now()
		log.Printf("PDF Worker %d started job %s", id, job.JobID)

		// Update job status to "PDF generation in-progress"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "PDF generation in-progress"
		jobStatusMutex.Unlock()

		// Generate PDF
		result, err := pdf.ExportPDF(job.TranslatedText, job.JobID)
		if err != nil {
			log.Printf("Job %s failed during PDF generation", job.JobID)
			jobStatusMutex.Lock()
			jobStatusMap[job.JobID] = "failed"
			jobStatusMutex.Unlock()
			continue
		}

		elapsed_time := time.Since(start_time)
		log.Printf("PDF generation took %v\n", elapsed_time)

		job.OutFilePath = result

		// Update job status to "completed"
		jobStatusMutex.Lock()
		jobStatusMap[job.JobID] = "completed"
		jobStatusMutex.Unlock()

		log.Printf("PDF Worker %d finished job %s", id, job.JobID)
	}
}
