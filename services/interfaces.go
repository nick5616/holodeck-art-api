package services

import (
	"context"

	"github.com/nick5616/holodeck-art-api/models"
)

// StorageProvider handles GCS operations
type StorageProvider interface {
	Save(ctx context.Context, data []byte) (objectID string, err error)
	SetMetadata(ctx context.Context, objectID string, meta models.Metadata) error
	ListFavorites(ctx context.Context) ([]models.ArtPiece, error)
	GetSignedURL(ctx context.Context, objectID string) (string, error)
}

// AIProvider handles image analysis
type AIProvider interface {
	AnalyzeImage(ctx context.Context, data []byte) (models.AIAnalysis, error)
}