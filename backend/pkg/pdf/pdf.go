package pdf

import (
	"fmt"
	"bytes"
	"io/ioutil"
	"net/http"

	gofpdf "github.com/jung-kurt/gofpdf"
)

func ExportPDF(translatedText, jobID string, margins map[string]float64) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(margins["left"], margins["top"], margins["right"])

	pdf.AddUTF8Font("DejaVu", "", "./fonts/DejaVuSans.ttf")
	pdf.SetFont("DejaVu", "", 14)

	pdf.MoveTo(0, 20)
	width, _ := pdf.GetPageSize()
	pdf.MultiCell(width, 10, translatedText, "", "", false)
	OutFilePath := fmt.Sprintf("./output/%s.pdf", jobID)
	err := pdf.OutputFileAndClose(OutFilePath)

	// err := pdf.OutputFileAndClose("./output/sample.pdf")

	if err != nil {
		return "", fmt.Errorf("failed to export to pdf file: %v", err)
	}
	return OutFilePath, nil
}


// ExportPDFtoS3 generates a PDF and uploads it to S3 using a presigned URL
func ExportPDFtoS3(translatedText, jobID string, margins map[string]float64, presignURL string) (string, error) {
	// Generate the PDF content as a buffer
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(margins["left"], margins["top"], margins["right"])

	pdf.AddUTF8Font("DejaVu", "", "./fonts/DejaVuSans.ttf")
	pdf.SetFont("DejaVu", "", 14)

	pdf.MoveTo(0, 20)
	width, _ := pdf.GetPageSize()
	pdf.MultiCell(width, 10, translatedText, "", "", false)

	// Save PDF content to a buffer instead of file
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF buffer: %v", err)
	}

	// Prepare the request to upload the PDF to the presigned URL
	req, err := http.NewRequest("PUT", presignURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set the necessary headers (assuming presign URL expects them)
	req.Header.Set("Content-Type", "application/pdf")

	// Perform the PUT request to upload the file
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to S3: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful upload (status code 200-299)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload PDF to S3, status code: %d, response: %s", resp.StatusCode, body)
	}

	// Return the URL of the uploaded file (could be the presigned URL itself)
	return presignURL, nil
}
