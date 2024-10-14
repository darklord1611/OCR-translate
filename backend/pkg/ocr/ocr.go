
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



func OCRFilter(imagePath string) (string, error) {
	// Get a Tesseract client from the pool
    client := tesseractPool.Get().(*gosseract.Client)

    // Always put the client back into the pool after use
    defer tesseractPool.Put(client)

	err := client.SetImage(imagePath)
	if err != nil {
        return "", fmt.Errorf("failed to set image: %v", err)
    }

	text, err := client.Text()
	if err != nil {
        return "", fmt.Errorf("failed to extract text: %v", err)
    }

	// Hello, World!
    return strings.ReplaceAll(text, "\n", ""), nil
}