
package main


import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
)


func main(translatedText string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, translatedText)
	const new_uuid string = uuid.New().String()
	err := pdf.OutputFileAndClose(fmt.Sprintf("./$s.pdf", new_uuid))
	fmt.Println(err)
}