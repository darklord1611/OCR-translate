package pdf


import (
	"fmt"
	"github.com/google/uuid"
    gofpdf "github.com/jung-kurt/gofpdf"
)





func ExportPDF(translatedText string) string {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

    pdf.AddUTF8Font("DejaVu", "", "./fonts/DejaVuSans.ttf")
    pdf.SetFont("DejaVu", "", 14)

    pdf.MoveTo(0, 20)
    width, _ := pdf.GetPageSize()
    pdf.MultiCell(width, 10, translatedText, "", "", false)
	new_uuid := uuid.New().String()
    fmt.Println(new_uuid)
	// err := pdf.OutputFileAndClose(fmt.Sprintf("./%s.pdf", new_uuid))

    err := pdf.OutputFileAndClose("./output/sample.pdf")

	if err != nil {
        return fmt.Sprintf("export file err: %s", err.Error())
    }
    return "Export file successfully"
}