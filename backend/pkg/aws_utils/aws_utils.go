package aws_utils

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"io"
	"io/ioutil"
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	globalS3Session *session.Session
	bucketName      string
)

func InitS3Session(region, access_key, secret_access_key string) {

	globalS3Session = session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(access_key, secret_access_key, ""),
	}))

}



func GenerateUploadURL(bucket, key string, expiration time.Duration) (string, error) {
	
	s3Client := s3.New(globalS3Session)

	req, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return urlStr, nil
}

func GenerateDownloadURL(bucket, key string, expiration time.Duration) (string, error) {
	s3Client := s3.New(globalS3Session)

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return urlStr, nil
}

func UploadStream(file io.Reader, uploadURL string) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	req, err := http.NewRequest("PUT", uploadURL, buf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", body)
	}

	return nil
}

func UploadFile(filePath, uploadURL string) error {
	// Read the file to be uploaded
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create a new PUT request with the file content
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content-type header (this can vary based on your file type)
	req.Header.Set("Content-Type", "application/octet-stream")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check if the upload was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file, status code: %d", resp.StatusCode)
	}

	fmt.Println("File uploaded successfully")
	return nil
}



func DownloadFile(downloadURL, destinationPath string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s", body)
	}

	outFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// 3 URLs for one job, 1 URL for upload, 1 URL for download 

func GeneratePresignedURL(bucket, key string, expiration time.Duration) (string, string, error) {
	svc := s3.New(globalS3Session)

	GetReq, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	GetUrlStr, err := GetReq.Presign(expiration)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}


	PutReq, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	PutUrlStr, err := PutReq.Presign(expiration)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return GetUrlStr, PutUrlStr, nil
}