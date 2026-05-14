package ranking

import (
	"bookrank/internal/api/middleware"
	"bookrank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler handles ranking-related HTTP requests
type Handler struct {
	rankingService *service.RankingService
}

// NewHandler creates a new ranking handler
func NewHandler(rankingService *service.RankingService) *Handler {
	return &Handler{
		rankingService: rankingService,
	}
}

// GetUserRankings handles GET /api/rankings/user/{userId}
func (h *Handler) GetUserRankings(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	authUserID := claims.UserID

	// Get requested user ID from URL
	vars := mux.Vars(r)
	requestedUserID, err := strconv.ParseUint(vars["userId"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// For now, only allow users to view their own rankings
	// TODO: Add privacy settings to allow sharing rankings
	if uint(requestedUserID) != authUserID {
		http.Error(w, "Forbidden: can only view your own rankings", http.StatusForbidden)
		return
	}

	rankings, err := h.rankingService.GetUserRankings(uint(requestedUserID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := h.rankingService.GetUserRankingStats(uint(requestedUserID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rankings": rankings,
		"stats":    stats,
		"user_id":  requestedUserID,
	})
}

// GetMyRankings handles GET /api/rankings/me
func (h *Handler) GetMyRankings(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	rankings, err := h.rankingService.GetUserRankings(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := h.rankingService.GetUserRankingStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rankings": rankings,
		"stats":    stats,
		"user_id":  userID,
	})
}

// GetTopRanked handles GET /api/rankings/top
func (h *Handler) GetTopRanked(w http.ResponseWriter, r *http.Request) {
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
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	rankings, err := h.rankingService.GetTopRankedBooks(userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"top_rankings": rankings,
		"limit":        limit,
		"count":        len(rankings),
	})
}

// GetRankingStats handles GET /api/rankings/stats
func (h *Handler) GetRankingStats(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	stats, err := h.rankingService.GetUserRankingStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetRankingPosition handles GET /api/rankings/position/{bookId}
func (h *Handler) GetRankingPosition(w http.ResponseWriter, r *http.Request) {
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

	position, err := h.rankingService.GetRankingPosition(userID, uint(bookID))
	if err != nil {
		if err.Error() == "book not found in user's rankings" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ranking, err := h.rankingService.GetRanking(userID, uint(bookID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position": position,
		"ranking":  ranking,
		"book_id":  bookID,
	})
}

// CompareRankings handles GET /api/rankings/compare
func (h *Handler) CompareRankings(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get book IDs from query parameters
	bookAIDStr := r.URL.Query().Get("book_a")
	bookBIDStr := r.URL.Query().Get("book_b")

	if bookAIDStr == "" || bookBIDStr == "" {
		http.Error(w, "book_a and book_b parameters are required", http.StatusBadRequest)
		return
	}

	bookAID, err := strconv.ParseUint(bookAIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid book_a ID", http.StatusBadRequest)
		return
	}

	bookBID, err := strconv.ParseUint(bookBIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid book_b ID", http.StatusBadRequest)
		return
	}

	if bookAID == bookBID {
		http.Error(w, "Cannot compare a book with itself", http.StatusBadRequest)
		return
	}

	comparison, err := h.rankingService.GetRankingComparison(userID, uint(bookAID), uint(bookBID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// InitializeRanking handles POST /api/rankings/initialize
func (h *Handler) InitializeRanking(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	var req struct {
		BookID  uint   `json:"book_id" validate:"required"`
		BookIDs []uint `json:"book_ids,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle single book initialization
	if req.BookID != 0 {
		err := h.rankingService.InitializeRanking(userID, req.BookID)
		if err != nil {
			if err.Error() == "ranking already exists" {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Ranking initialized successfully",
			"book_id": req.BookID,
		})
		return
	}

	// Handle bulk initialization
	if len(req.BookIDs) > 0 {
		err := h.rankingService.BulkInitializeRankings(userID, req.BookIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Rankings initialized successfully",
			"book_ids":      req.BookIDs,
			"books_count":   len(req.BookIDs),
		})
		return
	}

	http.Error(w, "Either book_id or book_ids must be provided", http.StatusBadRequest)
}