package dto

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"time"
)

// CreateBookRequest represents the request to create a book
type CreateBookRequest struct {
	Title           string     `json:"title" validate:"required,min=1,max=255"`
	Author          string     `json:"author" validate:"required,min=1,max=255"`
	Genre           string     `json:"genre,omitempty" validate:"omitempty,min=1,max=100"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Description     string     `json:"description,omitempty" validate:"max=5000"`
	FetchMetadata   bool       `json:"fetch_metadata,omitempty"`
}

// UpdateBookRequest represents the request to update a book
type UpdateBookRequest struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Author          *string    `json:"author,omitempty" validate:"omitempty,min=1,max=255"`
	Genre           *string    `json:"genre,omitempty" validate:"omitempty,min=1,max=100"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
}

// BookResponse represents a book in API responses
type BookResponse struct {
	ID              uint                   `json:"id"`
	Title           string                 `json:"title"`
	Author          string                 `json:"author"`
	Genre           string                 `json:"genre"`
	PublicationDate *time.Time             `json:"publication_date,omitempty"`
	Description     string                 `json:"description,omitempty"`
	CreatedBy       uint                   `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Creator         *UserSummary           `json:"creator,omitempty"`
	PersonalRating  *int                   `json:"personal_rating,omitempty"`
	AverageRating   *float64               `json:"average_rating,omitempty"`
	TotalRatings    *int64                 `json:"total_ratings,omitempty"`
	Metadata        []*BookMetadataResponse `json:"metadata,omitempty"`
}

// BookSummary represents a condensed book response
type BookSummary struct {
	ID            uint       `json:"id"`
	Title         string     `json:"title"`
	Author        string     `json:"author"`
	Genre         string     `json:"genre"`
	AverageRating *float64   `json:"average_rating,omitempty"`
	TotalRatings  *int64     `json:"total_ratings,omitempty"`
	CoverURL      *string    `json:"cover_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// UserSummary represents a condensed user response
type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// BookMetadataResponse represents book metadata in API responses
type BookMetadataResponse struct {
	ID             uint                   `json:"id"`
	Source         string                 `json:"source"`
	ExternalID     string                 `json:"external_id,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// BookListResponse represents paginated book results
type BookListResponse struct {
	Books      []*BookResponse `json:"books"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// BookStatsResponse represents book statistics
type BookStatsResponse struct {
	BookID             uint            `json:"book_id"`
	TotalRatings       int64           `json:"total_ratings"`
	AverageRating      float64         `json:"average_rating"`
	HighestRating      int             `json:"highest_rating"`
	LowestRating       int             `json:"lowest_rating"`
	RatingDistribution map[int]int64   `json:"rating_distribution"`
}

// BookSearchResponse represents search results
type BookSearchResponse struct {
	Books      []*BookSummary `json:"books"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	Query      string         `json:"query"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

// ToBookResponse converts a Book model to BookResponse DTO
func ToBookResponse(book *models.Book) *BookResponse {
	response := &BookResponse{
		ID:              book.ID,
		Title:           book.Title,
		Author:          book.Author,
		Genre:           book.Genre,
		PublicationDate: book.PublicationDate,
		Description:     book.Description,
		CreatedBy:       book.CreatedBy,
		CreatedAt:       book.CreatedAt,
		UpdatedAt:       book.UpdatedAt,
	}

	// Add creator information if loaded
	if book.Creator.ID != 0 {
		response.Creator = &UserSummary{
			ID:       book.Creator.ID,
			Username: book.Creator.Username,
			Email:    "", // User model doesn't have email field
		}
	}

	// Add metadata if loaded
	if len(book.Metadata) > 0 {
		response.Metadata = make([]*BookMetadataResponse, len(book.Metadata))
		for i, metadata := range book.Metadata {
			response.Metadata[i] = &BookMetadataResponse{
				ID:             metadata.ID,
				Source:         string(metadata.Source),
				ExternalID:     metadata.ExternalID,
				AdditionalData: map[string]interface{}(metadata.AdditionalData),
				CreatedAt:      metadata.CreatedAt,
				UpdatedAt:      metadata.UpdatedAt,
			}
		}
	}

	return response
}

// ToBookSummary converts a Book model to BookSummary DTO
func ToBookSummary(book *models.Book) *BookSummary {
	return &BookSummary{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		Genre:     book.Genre,
		CreatedAt: book.CreatedAt,
	}
}

// ToBookStatsResponse converts BookStats to BookStatsResponse DTO
func ToBookStatsResponse(stats *repository.BookStats) *BookStatsResponse {
	return &BookStatsResponse{
		BookID:             stats.BookID,
		TotalRatings:       stats.TotalRatings,
		AverageRating:      stats.AverageRating,
		HighestRating:      stats.HighestRating,
		LowestRating:       stats.LowestRating,
		RatingDistribution: stats.RatingDistribution,
	}
}