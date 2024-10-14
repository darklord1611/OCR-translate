
package ocr

import (
	"fmt"
	"strings"
	"github.com/otiai10/gosseract/v2"
)




func OCRFilter(imagePath string) string {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(imagePath)
	text, _ := client.Text()
	fmt.Println(text)
	// Hello, World!
    return strings.ReplaceAll(text, "\n", "")
}