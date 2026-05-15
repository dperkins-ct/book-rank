package comparison

import (
	"bookrank/internal/api/middleware"
	"bookrank/internal/models"
	"bookrank/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler handles comparison-related HTTP requests
type Handler struct {
	comparisonService *service.ComparisonService
}

// NewHandler creates a new comparison handler
func NewHandler(comparisonService *service.ComparisonService) *Handler {
	return &Handler{
		comparisonService: comparisonService,
	}
}

// FrontendComparisonRequest represents the request format from the frontend
type FrontendComparisonRequest struct {
	BookAID  uint `json:"book_a_id" validate:"required"`
	BookBID  uint `json:"book_b_id" validate:"required"`
	WinnerID uint `json:"winner_id" validate:"required"`
}

// SubmitComparison handles POST /api/comparisons
func (h *Handler) SubmitComparison(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var frontendReq FrontendComparisonRequest
	if err := json.NewDecoder(r.Body).Decode(&frontendReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert frontend request to service request format
	var preference models.ComparisonPreference
	if frontendReq.WinnerID == frontendReq.BookAID {
		preference = models.PreferenceBookA
	} else if frontendReq.WinnerID == frontendReq.BookBID {
		preference = models.PreferenceBookB
	} else {
		http.Error(w, "Winner ID must match either book_a_id or book_b_id", http.StatusBadRequest)
		return
	}

	req := service.ComparisonRequest{
		UserID:     claims.UserID,
		BookAID:    frontendReq.BookAID,
		BookBID:    frontendReq.BookBID,
		Preference: preference,
	}

	response, err := h.comparisonService.SubmitComparison(&req)
	if err != nil {
		if err.Error() == "comparison already exists for these books" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPendingComparisons handles GET /api/comparisons/pending
func (h *Handler) GetPendingComparisons(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Parse limit parameter
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	pending, err := h.comparisonService.GetPendingComparisons(userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_comparisons": pending,
		"count":              len(pending),
	})
}

// GetComparisonHistory handles GET /api/comparisons and GET /api/comparisons/history
func (h *Handler) GetComparisonHistory(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Parse limit parameter
	limit := 0 // 0 means no limit for history
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	comparisons, err := h.comparisonService.GetComparisonHistory(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply limit if specified
	if limit > 0 && len(comparisons) > limit {
		comparisons = comparisons[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comparisons": comparisons,
		"count":       len(comparisons),
	})
}

// GetBookComparisons handles GET /api/comparisons/book/{bookId}
func (h *Handler) GetBookComparisons(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get book ID from URL
	vars := mux.Vars(r)
	bookID, err := strconv.ParseUint(vars["bookId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	comparisons, err := h.comparisonService.GetBookComparisons(userID, uint(bookID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comparisons": comparisons,
		"book_id":     bookID,
		"count":       len(comparisons),
	})
}

// RecalculateRatings handles POST /api/comparisons/recalculate (admin endpoint)
func (h *Handler) RecalculateRatings(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// TODO: Add admin check here if needed
	// For now, allow any authenticated user to recalculate their own ratings

	err := h.comparisonService.RecalculateUserRatings(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Ratings recalculated successfully",
	})
}

// CheckOnboardingStatus handles GET /api/comparisons/onboarding-status
func (h *Handler) CheckOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	isComplete, err := h.comparisonService.IsOnboardingComplete(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"onboarding_complete": isComplete,
		"user_id":            userID,
	})
}

// GetRandomBookPair handles GET /api/comparisons/random-pair
func (h *Handler) GetRandomBookPair(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	bookPair, err := h.comparisonService.GetRandomBookPairWithContext(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return // Request was cancelled
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if bookPair == nil {
		// No book pairs available - return empty result with 200 status
		json.NewEncoder(w).Encode(map[string]interface{}{
			"book_pair": nil,
			"message":   "No book pairs available for comparison",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"book_pair": bookPair,
	})
}