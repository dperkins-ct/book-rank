package book

import (
	"bookrank/internal/api/dto"
	"bookrank/internal/api/middleware"
	"bookrank/internal/repository"
	"bookrank/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Handlers contains all book-related HTTP handlers
type Handlers struct {
	bookService *service.BookService
}

// NewHandlers creates new book handlers
func NewHandlers(bookService *service.BookService) *Handlers {
	return &Handlers{
		bookService: bookService,
	}
}

// CreateBook handles POST /api/books
func (h *Handlers) CreateBook(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context (set by auth middleware)
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req dto.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Convert DTO to service request
	serviceReq := &service.BookCreateRequest{
		Title:           req.Title,
		Author:          req.Author,
		Genre:           req.Genre,
		PublicationDate: req.PublicationDate,
		Description:     req.Description,
		FetchMetadata:   req.FetchMetadata,
	}

	book, err := h.bookService.CreateBook(serviceReq, claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "validation error") {
			writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create book", err.Error())
		return
	}

	response := dto.ToBookResponse(book)
	writeJSONResponse(w, http.StatusCreated, response)
}

// GetBooks handles GET /api/books
func (h *Handlers) GetBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse pagination
	limit := parseIntQuery(query.Get("limit"), 20)
	page := parseIntQuery(query.Get("page"), 1)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Parse filters
	filter := &repository.BookFilter{}
	if genre := query.Get("genre"); genre != "" {
		filter.Genre = genre
	}
	if author := query.Get("author"); author != "" {
		filter.Author = author
	}
	if search := query.Get("search"); search != "" {
		filter.Search = search
	}
	if minRating := query.Get("min_rating"); minRating != "" {
		if rating := parseIntQuery(minRating, 0); rating > 0 {
			filter.MinRating = &rating
		}
	}
	if maxRating := query.Get("max_rating"); maxRating != "" {
		if rating := parseIntQuery(maxRating, 0); rating > 0 {
			filter.MaxRating = &rating
		}
	}

	// Parse sorting
	sort := &repository.BookSort{
		Field:     "created_at",
		Direction: "desc",
	}
	if sortBy := query.Get("sort"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortDir := query.Get("order"); sortDir != "" {
		sort.Direction = sortDir
	}

	options := repository.BookQueryOptions{
		Limit:  limit,
		Offset: offset,
		Filter: filter,
		Sort:   sort,
	}

	result, err := h.bookService.GetBooks(options)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get books", err.Error())
		return
	}

	// Convert to DTO
	response := &dto.BookListResponse{
		Books:      make([]*dto.BookResponse, len(result.Books)),
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	for i, book := range result.Books {
		response.Books[i] = dto.ToBookResponse(book)
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GetBook handles GET /api/books/{id}
func (h *Handlers) GetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "")
		return
	}

	book, err := h.bookService.GetBookByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, "Book not found", "")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get book", err.Error())
		return
	}

	response := dto.ToBookResponse(book)
	writeJSONResponse(w, http.StatusOK, response)
}

// UpdateBook handles PUT /api/books/{id}
func (h *Handlers) UpdateBook(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "")
		return
	}

	var req dto.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Convert DTO to service request
	serviceReq := &service.BookUpdateRequest{
		Title:           req.Title,
		Author:          req.Author,
		Genre:           req.Genre,
		PublicationDate: req.PublicationDate,
		Description:     req.Description,
	}

	book, err := h.bookService.UpdateBook(uint(id), serviceReq, claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, "Book not found", "")
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			writeErrorResponse(w, http.StatusForbidden, "Unauthorized", err.Error())
			return
		}
		if strings.Contains(err.Error(), "validation error") {
			writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update book", err.Error())
		return
	}

	response := dto.ToBookResponse(book)
	writeJSONResponse(w, http.StatusOK, response)
}

// DeleteBook handles DELETE /api/books/{id}
func (h *Handlers) DeleteBook(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "")
		return
	}

	err = h.bookService.DeleteBook(uint(id), claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, "Book not found", "")
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			writeErrorResponse(w, http.StatusForbidden, "Unauthorized", err.Error())
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete book", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBookStats handles GET /api/books/{id}/stats
func (h *Handlers) GetBookStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "")
		return
	}

	stats, err := h.bookService.GetBookStats(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, "Book not found", "")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get book stats", err.Error())
		return
	}

	response := dto.ToBookStatsResponse(stats)
	writeJSONResponse(w, http.StatusOK, response)
}

// RefreshMetadata handles POST /api/books/{id}/metadata
func (h *Handlers) RefreshMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid book ID", "")
		return
	}

	err = h.bookService.RefreshMetadata(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErrorResponse(w, http.StatusNotFound, "Book not found", "")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to refresh metadata", err.Error())
		return
	}

	response := dto.SuccessResponse{
		Message: "Metadata refresh initiated successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// SearchBooks handles GET /api/books/search
func (h *Handlers) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	searchQuery := query.Get("q")
	if strings.TrimSpace(searchQuery) == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Search query is required", "")
		return
	}

	// Parse pagination
	limit := parseIntQuery(query.Get("limit"), 20)
	page := parseIntQuery(query.Get("page"), 1)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	result, err := h.bookService.SearchBooks(searchQuery, limit, offset)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}

	// Convert to search response DTO
	response := &dto.BookSearchResponse{
		Books:      make([]*dto.BookSummary, len(result.Books)),
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		Query:      searchQuery,
	}

	for i, book := range result.Books {
		response.Books[i] = dto.ToBookSummary(book)
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// Helper functions

func parseIntQuery(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}
	return defaultValue
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeErrorResponse(w http.ResponseWriter, status int, message, details string) {
	response := dto.ErrorResponse{
		Error:   message,
		Message: details,
	}
	writeJSONResponse(w, status, response)
}