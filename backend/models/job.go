package models

import (
	"time"
)

type Job struct {
	ImagePath 	string
	ImageDownloadURL	string
	PDFUploadURL	string
	JobID		string
	ExtractedText string
	TranslatedText string
	OutFilePath	string
	SubmittedAt  time.Time `json:"submitted_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	ResponseTime time.Duration `json:"-"`
}


