package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
    "backend/pkg/ocr"
    "backend/pkg/pdf"
    "backend/pkg/translation"
    "backend/models"

)

func main() {
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

        job = models.Job{} {
            ImagePath:  imagePath,
            JobID:      uuid.New().String()
        }

        originalText := ocr.OCRFilter(imagePath)
        translatedText := translation.TranslateFilter(originalText)
        result := pdf.ExportPDF(translatedText)

        if result != "Export file successfully" {
            c.String(http.StatusInternalServerError, result)
            return
        }

        // Respond with a success message
        c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded successfully!", file.Filename))
    })

    // Start the server on port 8080
    r.Run(":8080")
}



func worker(id int, jobs <-chan Job) {
    for job := range jobs {
        log.Printf("Worker %d started job %s", id, job.JobID)

        // Perform the OCR, translation, and PDF generation here
        processOCR(job.ImagePath)
        // Additional steps...

        log.Printf("Worker %d finished job %s", id, job.JobID)
    }
}



