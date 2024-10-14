package main

import (
    "fmt"
    "net/http"
    "log"
    "time"
    "github.com/gin-gonic/gin"
    "backend/pkg/ocr"
    "backend/pkg/pdf"
    "backend/pkg/translation"
    "backend/models"
    "github.com/google/uuid"
)

const numWorkers = 5

var jobQueue = make(chan models.Job, 100) // Job queue channel with buffer size 100


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

    r.GET("/", func (c *gin.Context) {
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

        // pipeline

        job := models.Job{
            ImagePath:  imagePath,
            JobID:      uuid.New().String(),
        }

        jobQueue <- job

        // Respond with a success message
        c.JSON(200, gin.H{"message": "Job submitted", "jobID": job.JobID})
    })

    // Start the server on port 8080
    r.Run(":8080")
}



func worker(id int, jobs <-chan models.Job) {
    for job := range jobs {
        start_time := time.Now()
        log.Printf("Worker %d started job %s", id, job.JobID)

        // Perform the OCR, translation, and PDF generation here
        originalText, err := ocr.OCRFilter(job.ImagePath)
        if err != nil {
            log.Printf("Job %s failed", id, job.JobID)
        }
        translatedText := translation.TranslateFilter(originalText)
        result, err := pdf.ExportPDF(translatedText, job.JobID)
        // Additional steps...
        if err != nil {
            log.Printf("Job %s failed", id, job.JobID)
        }

        job.OutFilePath = result

        log.Printf("Worker %d finished job %s", id, job.JobID)
        elapsed_time := time.Since(start_time)
        log.Printf("Request took %v\n", elapsed_time)
    }
}



