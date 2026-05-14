package repository

import (
	"bookrank/internal/models"
	"errors"
	"fmt"
	"strings"
	"gorm.io/gorm"
)

// BookFilter represents filtering options for book queries
type BookFilter struct {
	Genre       string
	Author      string
	MinRating   *int
	MaxRating   *int
	CreatedBy   *uint
	Search      string
}

// BookSort represents sorting options for book queries
type BookSort struct {
	Field     string // "title", "author", "created_at", "rating"
	Direction string // "asc", "desc"
}

// BookQueryOptions represents pagination and filtering options
type BookQueryOptions struct {
	Limit  int
	Offset int
	Filter *BookFilter
	Sort   *BookSort
}

// BookRepository defines the interface for book data operations
type BookRepository interface {
	Create(book *models.Book) error
	GetByID(id uint) (*models.Book, error)
	GetByIDWithMetadata(id uint) (*models.Book, error)
	GetAll(options BookQueryOptions) ([]*models.Book, int64, error)
	GetAllBooks() ([]*models.Book, error)
	Update(book *models.Book) error
	Delete(id uint) error
	SoftDelete(id uint) error
	Search(query string, limit, offset int) ([]*models.Book, int64, error)
	GetByGenre(genre string, limit, offset int) ([]*models.Book, error)
	GetByAuthor(author string, limit, offset int) ([]*models.Book, error)
	GetBookStats(bookID uint) (*BookStats, error)
	CreateMetadata(metadata *models.BookMetadata) error
	UpdateMetadata(metadata *models.BookMetadata) error
	UpsertMetadata(metadata *models.BookMetadata) error
	GetMetadataByBookID(bookID uint) ([]*models.BookMetadata, error)
	GetBooksByRatingRange(minRating, maxRating int, limit, offset int) ([]*models.Book, error)
	Count() (int64, error)
}

// BookStats represents statistics for a book
type BookStats struct {
	BookID          uint    `json:"book_id"`
	TotalRatings    int64   `json:"total_ratings"`
	AverageRating   float64 `json:"average_rating"`
	HighestRating   int     `json:"highest_rating"`
	LowestRating    int     `json:"lowest_rating"`
	RatingDistribution map[int]int64 `json:"rating_distribution"`
}

// bookRepository implements BookRepository
type bookRepository struct {
	db *gorm.DB
}

// NewBookRepository creates a new BookRepository
func NewBookRepository(db *gorm.DB) BookRepository {
	return &bookRepository{db: db}
}

// Helper functions for database-agnostic operations
func (r *bookRepository) ilike(field, value string) string {
	// Check if we're using SQLite
	if r.db.Dialector.Name() == "sqlite" {
		return fmt.Sprintf("%s LIKE ?", field)
	}
	// Use ILIKE for PostgreSQL
	return fmt.Sprintf("%s ILIKE ?", field)
}

func (r *bookRepository) floorDiv(field string, divisor int) string {
	// Check if we're using SQLite
	if r.db.Dialector.Name() == "sqlite" {
		return fmt.Sprintf("CAST(%s / %d AS INTEGER)", field, divisor)
	}
	// Use FLOOR for PostgreSQL
	return fmt.Sprintf("FLOOR(%s / %d)", field, divisor)
}

// Create creates a new book in the database
func (r *bookRepository) Create(book *models.Book) error {
	// Validate required fields
	if book.Title == "" {
		return errors.New("title is required")
	}
	if book.Author == "" {
		return errors.New("author is required")
	}

	return r.db.Create(book).Error
}

// GetByID retrieves a book by ID with creator information
func (r *bookRepository) GetByID(id uint) (*models.Book, error) {
	var book models.Book
	err := r.db.Preload("Creator").First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetByIDWithMetadata retrieves a book by ID with creator and metadata information
func (r *bookRepository) GetByIDWithMetadata(id uint) (*models.Book, error) {
	var book models.Book
	err := r.db.Preload("Creator").
		Preload("Metadata").
		Preload("Rankings").
		First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetAll retrieves books with advanced filtering, sorting, and pagination
func (r *bookRepository) GetAll(options BookQueryOptions) ([]*models.Book, int64, error) {
	var books []*models.Book
	var totalCount int64

	// Build base query
	query := r.db.Model(&models.Book{}).Preload("Creator")

	// Apply filters
	if options.Filter != nil {
		query = r.applyFilters(query, options.Filter)
	}

	// Count total records (before pagination)
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if options.Sort != nil {
		query = r.applySorting(query, options.Sort)
	} else {
		// Default sorting by creation date (newest first)
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	// Execute query
	err := query.Find(&books).Error
	return books, totalCount, err
}

// applyFilters applies filter conditions to the query
func (r *bookRepository) applyFilters(query *gorm.DB, filter *BookFilter) *gorm.DB {
	if filter.Genre != "" {
		query = query.Where(r.ilike("genre", filter.Genre), "%"+filter.Genre+"%")
	}

	if filter.Author != "" {
		query = query.Where(r.ilike("author", filter.Author), "%"+filter.Author+"%")
	}

	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if filter.Search != "" {
		searchQuery := "%" + filter.Search + "%"
		searchCondition := fmt.Sprintf("(%s OR %s OR %s)",
			r.ilike("title", filter.Search),
			r.ilike("author", filter.Search),
			r.ilike("description", filter.Search))
		query = query.Where(searchCondition, searchQuery, searchQuery, searchQuery)
	}

	// Rating filters require joining with rankings table
	if filter.MinRating != nil || filter.MaxRating != nil {
		query = query.Joins("LEFT JOIN rankings ON rankings.book_id = books.id").
			Group("books.id")

		if filter.MinRating != nil {
			query = query.Having("AVG(rankings.score) >= ?", *filter.MinRating)
		}
		if filter.MaxRating != nil {
			query = query.Having("AVG(rankings.score) <= ?", *filter.MaxRating)
		}
	}

	return query
}

// applySorting applies sorting to the query
func (r *bookRepository) applySorting(query *gorm.DB, sort *BookSort) *gorm.DB {
	direction := "ASC"
	if strings.ToUpper(sort.Direction) == "DESC" {
		direction = "DESC"
	}

	switch sort.Field {
	case "title":
		return query.Order(fmt.Sprintf("title %s", direction))
	case "author":
		return query.Order(fmt.Sprintf("author %s", direction))
	case "created_at":
		return query.Order(fmt.Sprintf("created_at %s", direction))
	case "updated_at":
		return query.Order(fmt.Sprintf("updated_at %s", direction))
	case "rating":
		// Join with rankings and order by average rating
		return query.Joins("LEFT JOIN rankings ON rankings.book_id = books.id").
			Group("books.id").
			Order(fmt.Sprintf("AVG(rankings.score) %s", direction))
	default:
		return query.Order("created_at DESC")
	}
}

// Update updates a book's information
func (r *bookRepository) Update(book *models.Book) error {
	return r.db.Save(book).Error
}

// Delete hard deletes a book
func (r *bookRepository) Delete(id uint) error {
	return r.db.Unscoped().Delete(&models.Book{}, id).Error
}

// SoftDelete soft deletes a book
func (r *bookRepository) SoftDelete(id uint) error {
	return r.db.Delete(&models.Book{}, id).Error
}

// Search searches for books by title or author with pagination
func (r *bookRepository) Search(query string, limit, offset int) ([]*models.Book, int64, error) {
	var books []*models.Book
	var totalCount int64

	searchQuery := "%" + query + "%"

	searchCondition := fmt.Sprintf("(%s OR %s OR %s)",
		r.ilike("title", query),
		r.ilike("author", query),
		r.ilike("description", query))

	baseQuery := r.db.Model(&models.Book{}).
		Where(searchCondition, searchQuery, searchQuery, searchQuery)

	// Count total results
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := baseQuery.Preload("Creator").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&books).Error

	return books, totalCount, err
}

// GetByGenre retrieves books by genre with pagination
func (r *bookRepository) GetByGenre(genre string, limit, offset int) ([]*models.Book, error) {
	var books []*models.Book
	err := r.db.Preload("Creator").
		Where(r.ilike("genre", genre), "%"+genre+"%").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&books).Error
	return books, err
}

// GetByAuthor retrieves books by author with pagination
func (r *bookRepository) GetByAuthor(author string, limit, offset int) ([]*models.Book, error) {
	var books []*models.Book
	err := r.db.Preload("Creator").
		Where(r.ilike("author", author), "%"+author+"%").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&books).Error
	return books, err
}

// GetBookStats retrieves statistics for a specific book
func (r *bookRepository) GetBookStats(bookID uint) (*BookStats, error) {
	stats := &BookStats{
		BookID:             bookID,
		RatingDistribution: make(map[int]int64),
	}

	// Get basic stats
	var result struct {
		TotalRatings  int64   `json:"total_ratings"`
		AverageRating float64 `json:"average_rating"`
		MaxRating     int     `json:"max_rating"`
		MinRating     int     `json:"min_rating"`
	}

	err := r.db.Model(&models.Ranking{}).
		Select("COUNT(*) as total_ratings, AVG(score) as average_rating, MAX(score) as max_rating, MIN(score) as min_rating").
		Where("book_id = ?", bookID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	stats.TotalRatings = result.TotalRatings
	stats.AverageRating = result.AverageRating
	stats.HighestRating = result.MaxRating
	stats.LowestRating = result.MinRating

	// Get rating distribution
	var distribution []struct {
		ScoreRange int   `json:"score_range"`
		Count      int64 `json:"count"`
	}

	err = r.db.Model(&models.Ranking{}).
		Select(fmt.Sprintf("%s as score_range, COUNT(*) as count", r.floorDiv("score", 300))).
		Where("book_id = ?", bookID).
		Group("score_range").
		Scan(&distribution).Error

	if err != nil {
		return nil, err
	}

	for _, dist := range distribution {
		// Convert ELO score ranges to 1-10 scale
		rating := dist.ScoreRange + 1
		if rating > 10 {
			rating = 10
		}
		stats.RatingDistribution[rating] = dist.Count
	}

	return stats, nil
}

// CreateMetadata creates new book metadata
func (r *bookRepository) CreateMetadata(metadata *models.BookMetadata) error {
	return r.db.Create(metadata).Error
}

// UpdateMetadata updates book metadata
func (r *bookRepository) UpdateMetadata(metadata *models.BookMetadata) error {
	return r.db.Save(metadata).Error
}

// GetMetadataByBookID retrieves all metadata for a book
func (r *bookRepository) GetMetadataByBookID(bookID uint) ([]*models.BookMetadata, error) {
	var metadata []*models.BookMetadata
	err := r.db.Where("book_id = ?", bookID).Find(&metadata).Error
	return metadata, err
}

// UpsertMetadata creates or updates book metadata
func (r *bookRepository) UpsertMetadata(metadata *models.BookMetadata) error {
	// Try to find existing metadata for this book and source
	var existing models.BookMetadata
	err := r.db.Where("book_id = ? AND source = ?", metadata.BookID, metadata.Source).First(&existing).Error

	if err == nil {
		// Update existing metadata
		existing.ExternalID = metadata.ExternalID
		existing.AdditionalData = metadata.AdditionalData
		return r.db.Save(&existing).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new metadata
		return r.db.Create(metadata).Error
	}

	// Return other errors
	return err
}

// GetBooksByRatingRange retrieves books within a specific rating range
func (r *bookRepository) GetBooksByRatingRange(minRating, maxRating int, limit, offset int) ([]*models.Book, error) {
	var books []*models.Book

	err := r.db.Preload("Creator").
		Joins("LEFT JOIN rankings ON rankings.book_id = books.id").
		Group("books.id").
		Having("AVG(rankings.score) BETWEEN ? AND ?", minRating*300, maxRating*300).
		Limit(limit).
		Offset(offset).
		Find(&books).Error

	return books, err
}

// Count returns the total number of books
func (r *bookRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Book{}).Count(&count).Error
	return count, err
}

// GetAllBooks returns all books without pagination (for recommendation algorithms)
func (r *bookRepository) GetAllBooks() ([]*models.Book, error) {
	var books []*models.Book
	err := r.db.Preload("Creator").Find(&books).Error
	return books, err
}