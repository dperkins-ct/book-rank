package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// GetUserRankings retrieves the current user's book rankings
func (h *Handlers) GetUserRankings(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get user rankings logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Get user rankings endpoint - not implemented yet",
	})
}

// CompareBooks handles book comparison for ELO rating
func (h *Handlers) CompareBooks(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement book comparison logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Compare books endpoint - not implemented yet",
	})
}

// GetRecommendations retrieves book recommendations for the user
func (h *Handlers) GetRecommendations(w http.ResponseWriter, r *http.Request) {
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
	recommendations, err := h.services.Recommendation.GetRecommendations(userID, limit)
	if err != nil {
		http.Error(w, "Failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
		"total_count":     len(recommendations),
	})
}