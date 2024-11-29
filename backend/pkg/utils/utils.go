package utils


import (
	"mime/multipart"
	"path/filepath"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"log"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
	"crypto/sha256"
	"io"
)



func MonitorCPUUsage(channel *amqp.Channel) {
	for {
		// Get the current CPU usage as a percentage
		usage, err := cpu.Percent(0, false)
		if err != nil {
			log.Printf("Error fetching CPU usage: %v", err)
			continue
		}

		// Adjust QoS based on CPU usage
		currentUsage := usage[0]
		var prefetchCount int
		if currentUsage < 50 {
			prefetchCount = 5 // Low CPU usage: allow up to 5 messages
		} else if currentUsage < 80 {
			prefetchCount = 2 // Medium CPU usage: allow up to 2 messages
		} else {
			prefetchCount = 1 // High CPU usage: allow only 1 message
		}

		// Set the new QoS
		err = channel.Qos(prefetchCount, 0, true)
		if err != nil {
			log.Printf("Error setting QoS: %v", err)
		}

		// Wait for a few seconds before checking again
		time.Sleep(5 * time.Second)
	}
}


// Function to generate a new filename with UUID appended before the extension
func GenerateNewFileName(file *multipart.FileHeader, uuid string) string {

	// Extract the base name and extension
	baseName := file.Filename[:len(file.Filename)-len(filepath.Ext(file.Filename))]
	ext := filepath.Ext(file.Filename)

	// Append the UUID to the base name
	newFileName := fmt.Sprintf("%s-%s%s", baseName, uuid, ext)

	return newFileName
}

func AddExtensionToFile(filename, ext string) string {
	// Get the file name without extension
	extExisting := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(extExisting)]

	// Add the new extension
	return baseName + ext
}


func GenerateHashFromFormFile(fileHeader *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a SHA-256 hash
	hash := sha256.New()

	// Read the file content in chunks and update the hash
	buffer := make([]byte, 8192) // 8 KB buffer
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		hash.Write(buffer[:n])
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
