package translation

import (
	"fmt"
	gt "github.com/bas24/googletranslatefree"
)


func TranslateFilter(text string) string {
	// you can use "auto" for source language
	// so, translator will detect language
	result, _ := gt.Translate(text, "en", "vi")
	fmt.Println("Translated successfully")
	// Output: "Hola, Mundo!"
    return result
}