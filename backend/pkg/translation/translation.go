package translation

import (
	gt "github.com/bas24/googletranslatefree"
)


func TranslateFilter(text string) string {
	// you can use "auto" for source language
	// so, translator will detect language
	result, _ := gt.Translate(text, "en", "vi")
	// Output: "Hola, Mundo!"
    return result
}