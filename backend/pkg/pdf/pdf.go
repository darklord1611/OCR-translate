package pdf

import (
	"fmt"

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
