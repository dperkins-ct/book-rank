package recommendation

import (
	"bookrank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler handles recommendation-related HTTP requests
type Handler struct {
	recommendationService *service.RecommendationService
}

// NewHandler creates a new recommendation handler
func NewHandler(recommendationService *service.RecommendationService) *Handler {
	return &Handler{
		recommendationService: recommendationService,
	}
}

// GetRecommendationsRequest represents the request for getting recommendations
type GetRecommendationsRequest struct {
	Limit int `json:"limit"`
}

// GetRecommendationsResponse represents the response for recommendations
type GetRecommendationsResponse struct {
	Recommendations []*service.RecommendationResult `json:"recommendations"`
	TotalCount      int                             `json:"total_count"`
	Strategies      []string                        `json:"strategies"`
}

// GetRecommendationsByGenreResponse represents genre-based recommendations response
type GetRecommendationsByGenreResponse struct {
	Genre           string                          `json:"genre"`
	Recommendations []*service.RecommendationResult `json:"recommendations"`
	TotalCount      int                             `json:"total_count"`
}

// SimilarBooksResponse represents similar books response
type SimilarBooksResponse struct {
	BookID      uint                            `json:"book_id"`
	SimilarBooks []*service.RecommendationResult `json:"similar_books"`
	TotalCount  int                             `json:"total_count"`
}

// StatsResponse represents recommendation engine statistics
type StatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// GetRecommendations handles GET /api/recommendations
func (h *Handler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get recommendations
	recommendations, err := h.recommendationService.GetRecommendations(userID, limit)
	if err != nil {
		http.Error(w, "Failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response with strategy information
	strategies := make([]string, 0)
	strategyMap := make(map[string]bool)
	for _, rec := range recommendations {
		if !strategyMap[rec.Strategy] {
			strategies = append(strategies, rec.Strategy)
			strategyMap[rec.Strategy] = true
		}
	}

	response := GetRecommendationsResponse{
		Recommendations: recommendations,
		TotalCount:      len(recommendations),
		Strategies:      strategies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetRecommendationsByGenre handles GET /api/recommendations/genre/{genre}
func (h *Handler) GetRecommendationsByGenre(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genre := vars["genre"]

	if genre == "" {
		http.Error(w, "Genre is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// For now, we'll use the general recommendations and filter by genre
	// In a production system, you'd implement genre-specific algorithms
	allRecommendations, err := h.recommendationService.GetRecommendations(userID, limit*2)
	if err != nil {
		http.Error(w, "Failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter by genre
	var genreRecommendations []*service.RecommendationResult
	for _, rec := range allRecommendations {
		if rec.Book.Genre == genre && len(genreRecommendations) < limit {
			genreRecommendations = append(genreRecommendations, rec)
		}
	}

	response := GetRecommendationsByGenreResponse{
		Genre:           genre,
		Recommendations: genreRecommendations,
		TotalCount:      len(genreRecommendations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSimilarBooks handles GET /api/recommendations/similar/{bookId}
func (h *Handler) GetSimilarBooks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookIDStr := vars["bookId"]

	bookID, err := strconv.ParseUint(bookIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get similar books
	similarBooks, err := h.recommendationService.GetSimilarBooks(uint(bookID), limit)
	if err != nil {
		http.Error(w, "Failed to get similar books: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := SimilarBooksResponse{
		BookID:      uint(bookID),
		SimilarBooks: similarBooks,
		TotalCount:  len(similarBooks),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RefreshRecommendations handles POST /api/recommendations/refresh
func (h *Handler) RefreshRecommendations(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Invalidate user's recommendation cache
	err := h.recommendationService.InvalidateUserRecommendations(userID)
	if err != nil {
		http.Error(w, "Failed to refresh recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Recommendations refreshed successfully",
	})
}

// GetRecommendationStats handles GET /api/recommendations/stats
func (h *Handler) GetRecommendationStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.recommendationService.GetRecommendationStats()
	if err != nil {
		http.Error(w, "Failed to get recommendation stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := StatsResponse{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}