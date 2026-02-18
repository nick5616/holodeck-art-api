package models

import "time"

type ArtPiece struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	Tags       []string  `json:"tags"`
	UploadedAt time.Time `json:"uploadedAt"`
}

type AIAnalysis struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type Metadata struct {
	Title      string
	Tags       string // comma-separated
	IsFavorite string
	UploadedAt string
}

type SubmissionResult struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Tags     []string   `json:"tags"`
	Status   string     `json:"status"`
}