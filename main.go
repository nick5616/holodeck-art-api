package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nick5616/holodeck-art-api/adapters"
	"github.com/nick5616/holodeck-art-api/config"
	"github.com/nick5616/holodeck-art-api/handlers"
	"github.com/nick5616/holodeck-art-api/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	
	// Load configuration
	cfg := config.Load()
	
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	ctx := context.Background()
	
	// Initialize adapters
	storage, err := adapters.NewGCSStorage(ctx, cfg.GCSBucketName)
	if err != nil {
		log.Fatalf("Failed to initialize GCS storage: %v", err)
	}
	defer storage.Close()
	
	aiAnalyzer := adapters.NewOpenAIAnalyzer(cfg.OpenAIAPIKey)
	rateLimiter := adapters.NewRateLimiter(1 * time.Minute)
	
	// Initialize service
	artService := services.NewArtService(storage, aiAnalyzer)
	
	// Initialize handler
	artHandler := handlers.NewArtHandler(artService, rateLimiter, logger)
	
	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/art", artHandler.SubmitArt)
	mux.HandleFunc("/api/v1/art/favorites", artHandler.ListFavorites)
	
	// CORS middleware for browser access
	handler := corsMiddleware(mux)
	
	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Graceful shutdown
	go func() {
		logger.Info("Server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Server shutting down...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	logger.Info("Server exited")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}