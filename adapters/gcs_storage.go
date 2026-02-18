package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/nick5616/holodeck-art-api/models"
	"google.golang.org/api/iterator"
)

type GCSStorage struct {
	client     *storage.Client
	bucketName string
}

func NewGCSStorage(ctx context.Context, bucketName string) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	
	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (g *GCSStorage) Save(ctx context.Context, data []byte) (string, error) {
	objectID := uuid.New().String() + ".png"
	
	wc := g.client.Bucket(g.bucketName).Object(objectID).NewWriter(ctx)
	wc.ContentType = "image/png"
	
	if _, err := wc.Write(data); err != nil {
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}
	
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}
	
	return objectID, nil
}

func (g *GCSStorage) SetMetadata(ctx context.Context, objectID string, meta models.Metadata) error {
	obj := g.client.Bucket(g.bucketName).Object(objectID)
	
	_, err := obj.Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"title":      meta.Title,
			"tags":       meta.Tags,
			"isFavorite": meta.IsFavorite,
			"uploadedAt": meta.UploadedAt,
		},
	})
	
	return err
}

func (g *GCSStorage) ListFavorites(ctx context.Context) ([]models.ArtPiece, error) {
	var pieces []models.ArtPiece
	
	it := g.client.Bucket(g.bucketName).Objects(ctx, nil)
	objectCount := 0
	favoriteCount := 0
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}
		
		objectCount++
		fmt.Printf("Found object: %s, isFavorite=%s\n", attrs.Name, attrs.Metadata["isFavorite"])
		
		// Filter by isFavorite metadata
		if attrs.Metadata["isFavorite"] != "true" {
			continue
		}
		
		favoriteCount++
		
		// Generate signed URL
		url, err := g.GetSignedURL(ctx, attrs.Name)
		if err != nil {
			fmt.Printf("ERROR generating signed URL for %s: %v\n", attrs.Name, err)
			continue // Skip this one if URL generation fails
		}
		
		fmt.Printf("Successfully generated URL for %s\n", attrs.Name)
		
		uploadedAt, _ := time.Parse(time.RFC3339, attrs.Metadata["uploadedAt"])
		
		pieces = append(pieces, models.ArtPiece{
			ID:         attrs.Name,
			URL:        url,
			Title:      attrs.Metadata["title"],
			Tags:       strings.Split(attrs.Metadata["tags"], ","),
			UploadedAt: uploadedAt,
		})
	}
	
	fmt.Printf("Total objects: %d, Favorites: %d, Pieces returned: %d\n", objectCount, favoriteCount, len(pieces))
	return pieces, nil
}

func (g *GCSStorage) GetSignedURL(ctx context.Context, objectID string) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(1 * time.Hour),
	}
	
	url, err := g.client.Bucket(g.bucketName).SignedURL(objectID, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	
	return url, nil
}

func (g *GCSStorage) Close() error {
	return g.client.Close()
}