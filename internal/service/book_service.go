package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)

// BookService handles business logic for book operations
type BookService struct {
	bookRepo    repository.BookRepository
	externalAPI *ExternalAPIService
}

// NewBookService creates a new BookService
func NewBookService(bookRepo repository.BookRepository) *BookService {
	return &BookService{
		bookRepo:    bookRepo,
		externalAPI: NewExternalAPIService(),
	}
}

// BookCreateRequest represents a request to create a book
type BookCreateRequest struct {
	Title           string     `json:"title" validate:"required,max=255"`
	Author          string     `json:"author" validate:"required,max=255"`
	Genre           string     `json:"genre,omitempty" validate:"omitempty,max=100"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Description     string     `json:"description,omitempty" validate:"max=5000"`
	FetchMetadata   bool       `json:"fetch_metadata,omitempty"`
}

// BookUpdateRequest represents a request to update a book
type BookUpdateRequest struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,max=255"`
	Author          *string    `json:"author,omitempty" validate:"omitempty,max=255"`
	Genre           *string    `json:"genre,omitempty" validate:"omitempty,max=100"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
}

// BookListResponse represents paginated book results
type BookListResponse struct {
	Books      []*models.Book `json:"books"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// CreateBook creates a new book with validation and optional metadata fetching
func (s *BookService) CreateBook(req *BookCreateRequest, userID uint) (*models.Book, error) {
	// Create book model
	book := &models.Book{
		Title:           req.Title,
		Author:          req.Author,
		Genre:           req.Genre,
		PublicationDate: req.PublicationDate,
		Description:     req.Description,
		CreatedBy:       userID,
	}

	// Validate book data
	if err := s.externalAPI.ValidateBookData(book); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create book in database
	if err := s.bookRepo.Create(book); err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	// Fetch and store metadata if requested
	if req.FetchMetadata {
		go s.fetchAndStoreMetadata(book.ID, book.Title, book.Author)
	}

	return book, nil
}

// GetBookByID retrieves a book by ID with full details
func (s *BookService) GetBookByID(id uint) (*models.Book, error) {
	book, err := s.bookRepo.GetByIDWithMetadata(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}
	return book, nil
}

// GetBooks retrieves books with filtering, sorting, and pagination
func (s *BookService) GetBooks(options repository.BookQueryOptions) (*BookListResponse, error) {
	// Set default limits
	if options.Limit <= 0 || options.Limit > 100 {
		options.Limit = 20
	}

	books, total, err := s.bookRepo.GetAll(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}

	page := (options.Offset / options.Limit) + 1
	totalPages := int((total + int64(options.Limit) - 1) / int64(options.Limit))

	return &BookListResponse{
		Books:      books,
		Total:      total,
		Page:       page,
		PageSize:   options.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateBook updates a book's information
func (s *BookService) UpdateBook(id uint, req *BookUpdateRequest, userID uint) (*models.Book, error) {
	// Get existing book
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	// Check permissions (only creator can update)
	if book.CreatedBy != userID {
		return nil, errors.New("unauthorized: only the book creator can update it")
	}

	// Update fields if provided
	if req.Title != nil {
		book.Title = *req.Title
	}
	if req.Author != nil {
		book.Author = *req.Author
	}
	if req.Genre != nil {
		book.Genre = *req.Genre
	}
	if req.PublicationDate != nil {
		book.PublicationDate = req.PublicationDate
	}
	if req.Description != nil {
		book.Description = *req.Description
	}

	// Validate updated data
	if err := s.externalAPI.ValidateBookData(book); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Save changes
	if err := s.bookRepo.Update(book); err != nil {
		return nil, fmt.Errorf("failed to update book: %w", err)
	}

	return book, nil
}

// DeleteBook soft deletes a book
func (s *BookService) DeleteBook(id uint, userID uint) error {
	// Get existing book
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("book not found")
		}
		return fmt.Errorf("failed to get book: %w", err)
	}

	// Check permissions (only creator can delete)
	if book.CreatedBy != userID {
		return errors.New("unauthorized: only the book creator can delete it")
	}

	// Soft delete the book
	if err := s.bookRepo.SoftDelete(id); err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	return nil
}

// SearchBooks searches for books by title or author
func (s *BookService) SearchBooks(query string, limit, offset int) (*BookListResponse, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query cannot be empty")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	books, total, err := s.bookRepo.Search(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search books: %w", err)
	}

	page := (offset / limit) + 1
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &BookListResponse{
		Books:      books,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

// GetBookStats retrieves statistics for a specific book
func (s *BookService) GetBookStats(id uint) (*repository.BookStats, error) {
	// Verify book exists
	_, err := s.bookRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	stats, err := s.bookRepo.GetBookStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get book stats: %w", err)
	}

	return stats, nil
}

// RefreshMetadata fetches and updates external metadata for a book
func (s *BookService) RefreshMetadata(id uint) error {
	book, err := s.bookRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("book not found")
		}
		return fmt.Errorf("failed to get book: %w", err)
	}

	return s.fetchAndStoreMetadata(book.ID, book.Title, book.Author)
}

// fetchAndStoreMetadata fetches metadata from external APIs and stores it
func (s *BookService) fetchAndStoreMetadata(bookID uint, title, author string) error {
	metadata, source, err := s.externalAPI.FetchBookMetadata(title, author)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	// Convert to database model
	bookMetadata := &models.BookMetadata{
		BookID:         bookID,
		Source:         source,
		AdditionalData: models.JSON(metadata.AdditionalData),
	}

	// Set external ID if available
	if source == models.SourceOpenLibrary {
		if key, ok := metadata.AdditionalData["openlibrary_key"]; ok {
			if keyStr, ok := key.(string); ok {
				bookMetadata.ExternalID = keyStr
			}
		}
	} else if source == models.SourceGoogleBooks {
		if id, ok := metadata.AdditionalData["google_books_id"]; ok {
			if idStr, ok := id.(string); ok {
				bookMetadata.ExternalID = idStr
			}
		}
	}

	// Check if metadata already exists for this book and source
	// If so, update it instead of creating new
	return s.bookRepo.UpsertMetadata(bookMetadata)
}