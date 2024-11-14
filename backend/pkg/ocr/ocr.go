package ocr

import (
	"fmt"
	"strings"
	"sync"

	"github.com/otiai10/gosseract/v2"
)

var tesseractPool sync.Pool

// Initialize initializes the Tesseract pool (call only once at startup)
func Initialize() {
	tesseractPool = sync.Pool{
		New: func() interface{} {
			// Create a new Tesseract client when needed
			return gosseract.NewClient()
		},
	}

	fmt.Println("Tesseract client pool initialized")
}

// Cleanup closes the Tesseract client (call on shutdown)
func Cleanup() {
	fmt.Println("No explicit cleanup needed, sync.Pool handles cleanup")
}

func OneShotOCR(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	err := client.SetImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to set image: %v", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %v", err)
	}

	return strings.ReplaceAll(text, "\n", ""), nil
}

// OCRFilter processes OCR on a single image
func OCRFilter(imagePath string) (string, error) {
	// Get a Tesseract client from the pool
	client := tesseractPool.Get().(*gosseract.Client)
	defer tesseractPool.Put(client)

	err := client.SetImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to set image: %v", err)
	}

	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %v", err)
	}

	return strings.ReplaceAll(text, "\n", ""), nil
}

// OCRFilterConcurrent performs OCR on a list of image paths concurrently
func OCRFilterConcurrent(imagePaths []string) (string, error) {
	var wg sync.WaitGroup
	var result string = ""
	var mu sync.Mutex
	var errOccurred error

	// Process each image path in a separate goroutine
	for _, path := range imagePaths {
		wg.Add(1)

		go func(imgPath string) {
			defer wg.Done()

			text, err := OneShotOCR(imgPath)
			if err != nil {
				errOccurred = err // capture the first error (optional)
				return
			}
			// Ensure safe access to the shared result map
			mu.Lock()
			result = result + "     " + strings.TrimSpace(text) + "\n"
			mu.Unlock()
		}(path)
	}
	// Wait for all goroutines to complete
	wg.Wait()

	// Return results and any error that occurred
	return result, errOccurred
}
