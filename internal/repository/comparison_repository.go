package repository

import (
	"bookrank/internal/models"
	"gorm.io/gorm"
)

// ComparisonRepository defines the interface for comparison data operations
type ComparisonRepository interface {
	Create(comparison *models.Comparison) error
	GetByUserID(userID uint) ([]*models.Comparison, error)
	GetByUserAndBooks(userID, bookAID, bookBID uint) (*models.Comparison, error)
	GetUserComparisonsForBook(userID, bookID uint) ([]*models.Comparison, error)
	GetPendingComparisons(userID uint, limit int) ([]PendingComparison, error)
	HasUserComparedBooks(userID, bookAID, bookBID uint) (bool, error)
}

// PendingComparison represents a pair of books that need comparison
type PendingComparison struct {
	BookA models.Book `json:"book_a"`
	BookB models.Book `json:"book_b"`
}

// comparisonRepository implements ComparisonRepository
type comparisonRepository struct {
	db *gorm.DB
}

// NewComparisonRepository creates a new ComparisonRepository
func NewComparisonRepository(db *gorm.DB) ComparisonRepository {
	return &comparisonRepository{db: db}
}

// Create creates a new comparison in the database
func (r *comparisonRepository) Create(comparison *models.Comparison) error {
	return r.db.Create(comparison).Error
}

// GetByUserID retrieves all comparisons for a specific user
func (r *comparisonRepository) GetByUserID(userID uint) ([]*models.Comparison, error) {
	var comparisons []*models.Comparison
	err := r.db.Preload("BookA").Preload("BookB").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&comparisons).Error
	return comparisons, err
}

// GetByUserAndBooks retrieves a specific comparison for a user and two books
func (r *comparisonRepository) GetByUserAndBooks(userID, bookAID, bookBID uint) (*models.Comparison, error) {
	var comparison models.Comparison
	err := r.db.Where(
		"user_id = ? AND ((book_a_id = ? AND book_b_id = ?) OR (book_a_id = ? AND book_b_id = ?))",
		userID, bookAID, bookBID, bookBID, bookAID,
	).First(&comparison).Error
	if err != nil {
		return nil, err
	}
	return &comparison, nil
}

// GetUserComparisonsForBook retrieves all comparisons involving a specific book for a user
func (r *comparisonRepository) GetUserComparisonsForBook(userID, bookID uint) ([]*models.Comparison, error) {
	var comparisons []*models.Comparison
	err := r.db.Preload("BookA").Preload("BookB").
		Where("user_id = ? AND (book_a_id = ? OR book_b_id = ?)", userID, bookID, bookID).
		Order("created_at DESC").
		Find(&comparisons).Error
	return comparisons, err
}

// GetPendingComparisons finds pairs of books that need comparison for a user
func (r *comparisonRepository) GetPendingComparisons(userID uint, limit int) ([]PendingComparison, error) {
	var pending []PendingComparison

	// Find books that the user has ranked but not compared against each other
	query := `
		SELECT DISTINCT b1.id as book_a_id, b1.title as book_a_title, b1.author as book_a_author, b1.genre as book_a_genre,
		                b2.id as book_b_id, b2.title as book_b_title, b2.author as book_b_author, b2.genre as book_b_genre
		FROM books b1
		CROSS JOIN books b2
		INNER JOIN rankings r1 ON b1.id = r1.book_id AND r1.user_id = ?
		INNER JOIN rankings r2 ON b2.id = r2.book_id AND r2.user_id = ?
		LEFT JOIN comparisons c ON c.user_id = ?
		    AND ((c.book_a_id = b1.id AND c.book_b_id = b2.id) OR (c.book_a_id = b2.id AND c.book_b_id = b1.id))
		WHERE b1.id < b2.id -- Avoid duplicate pairs and self-comparison
		  AND c.id IS NULL   -- No existing comparison
		ORDER BY RANDOM()
		LIMIT ?
	`

	type bookPair struct {
		BookAID     uint   `db:"book_a_id"`
		BookATitle  string `db:"book_a_title"`
		BookAAuthor string `db:"book_a_author"`
		BookAGenre  string `db:"book_a_genre"`
		BookBID     uint   `db:"book_b_id"`
		BookBTitle  string `db:"book_b_title"`
		BookBAuthor string `db:"book_b_author"`
		BookBGenre  string `db:"book_b_genre"`
	}

	var pairs []bookPair
	err := r.db.Raw(query, userID, userID, userID, limit).Scan(&pairs).Error
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		pending = append(pending, PendingComparison{
			BookA: models.Book{
				ID:     pair.BookAID,
				Title:  pair.BookATitle,
				Author: pair.BookAAuthor,
				Genre:  pair.BookAGenre,
			},
			BookB: models.Book{
				ID:     pair.BookBID,
				Title:  pair.BookBTitle,
				Author: pair.BookBAuthor,
				Genre:  pair.BookBGenre,
			},
		})
	}

	return pending, nil
}

// HasUserComparedBooks checks if a user has already compared two books
func (r *comparisonRepository) HasUserComparedBooks(userID, bookAID, bookBID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Comparison{}).Where(
		"user_id = ? AND ((book_a_id = ? AND book_b_id = ?) OR (book_a_id = ? AND book_b_id = ?))",
		userID, bookAID, bookBID, bookBID, bookAID,
	).Count(&count).Error
	return count > 0, err
}