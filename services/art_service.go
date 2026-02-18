package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nick5616/holodeck-art-api/models"
	"golang.org/x/sync/errgroup"
)

type ArtService struct {
	storage StorageProvider
	ai      AIProvider
}

func NewArtService(storage StorageProvider, ai AIProvider) *ArtService {
	return &ArtService{
		storage: storage,
		ai:      ai,
	}
}

// SubmitArt handles the concurrent upload + AI analysis pipeline
func (s *ArtService) SubmitArt(ctx context.Context, imgData []byte) (*models.SubmissionResult, error) {
	g, gctx := errgroup.WithContext(ctx)  // Rename to gctx
	
	var objectID string
	var analysis models.AIAnalysis
	
	// Concurrent execution: Upload to GCS + Analyze with AI
	g.Go(func() error {
		id, err := s.storage.Save(gctx, imgData)  // Use gctx here
		objectID = id
		return err
	})
	
	g.Go(func() error {
		res, err := s.ai.AnalyzeImage(gctx, imgData)  // Use gctx here
		analysis = res
		return err
	})
	
	// Wait for both operations to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to process artwork: %w", err)
	}
	
	// Update metadata after both operations succeed
	// Use the ORIGINAL context, not the errgroup context
	meta := models.Metadata{
		Title:      analysis.Title,
		Tags:       strings.Join(analysis.Tags, ","),
		IsFavorite: "false",
		UploadedAt: time.Now().Format(time.RFC3339),
	}
	
	if err := s.storage.SetMetadata(ctx, objectID, meta); err != nil {  // Use ctx, not gctx
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}
	
	return &models.SubmissionResult{
		ID:     objectID,
		Title:  analysis.Title,
		Tags:   analysis.Tags,
		Status: "processed",
	}, nil
}

// ListFavorites returns all favorited artworks
func (s *ArtService) ListFavorites(ctx context.Context) ([]models.ArtPiece, error) {
	return s.storage.ListFavorites(ctx)
}