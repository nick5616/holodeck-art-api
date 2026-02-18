package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nick5616/holodeck-art-api/services"
)

type ArtHandler struct {
	service     *services.ArtService
	rateLimiter RateLimiter
	logger      *slog.Logger
}

type RateLimiter interface {
	Allow(ip string) bool
}

func NewArtHandler(service *services.ArtService, rateLimiter RateLimiter, logger *slog.Logger) *ArtHandler {
	return &ArtHandler{
		service:     service,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (h *ArtHandler) SubmitArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Rate limiting
	ip := getIP(r)
	if !h.rateLimiter.Allow(ip) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Rate limit exceeded. Please wait 1 minute.", http.StatusTooManyRequests)
		return
	}
	
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		h.logger.Error("Failed to parse form", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	
	file, _, err := r.FormFile("image")
	if err != nil {
		h.logger.Error("Failed to get image file", "error", err)
		http.Error(w, "Missing or invalid image field", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Read image data
	imgData, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read image", "error", err)
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}
	
	// Validate image size
	if len(imgData) > 5<<20 { // 5MB
		http.Error(w, "Image too large (max 5MB)", http.StatusBadRequest)
		return
	}
	
	// Process artwork
	result, err := h.service.SubmitArt(r.Context(), imgData)
	if err != nil {
		h.logger.Error("Failed to submit art", "error", err)
		http.Error(w, "Failed to process artwork", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *ArtHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	pieces, err := h.service.ListFavorites(r.Context())
	if err != nil {
		h.logger.Error("Failed to list favorites", "error", err)
		http.Error(w, "Failed to retrieve favorites", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pieces": pieces,
	})
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}