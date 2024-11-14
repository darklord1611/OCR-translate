package translation

import (
	"log"

	gt "github.com/bas24/googletranslatefree"
)

func TranslateFilter(text string) string {
	// you can use "auto" for source language
	// so, translator will detect language
	result, err := gt.Translate(text, "en", "vi")
	// Output: "Hola, Mundo!"
	if err != nil {
		log.Println("Translation Error:", err)
	}
	return result
}
