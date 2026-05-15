package repository

import (
	"bookrank/internal/models"
	"gorm.io/gorm"
)

// RankingRepository defines the interface for ranking data operations
type RankingRepository interface {
	Create(ranking *models.Ranking) error
	GetByUserID(userID uint) ([]*models.Ranking, error)
	GetByUserAndBook(userID, bookID uint) (*models.Ranking, error)
	Update(ranking *models.Ranking) error
	GetTopRanked(userID uint, limit int) ([]*models.Ranking, error)
	GetAllByBookID(bookID uint) ([]*models.Ranking, error)
	GetAllUserRankings() (map[uint][]*models.Ranking, error)
	GetUsersByGenre(genre string) ([]uint, error)
	GetAverageRatingByGenre(userID uint) (map[string]float64, error)
	GetUnratedBooksByUser(userID uint) ([]*models.Book, error)
	GetTopRatedBooksByGenre(genre string, limit int) ([]*models.Book, error)
	GetUsersWhoRatedBook(bookID uint) ([]uint, error)
}

// rankingRepository implements RankingRepository
type rankingRepository struct {
	db *gorm.DB
}

// NewRankingRepository creates a new RankingRepository
func NewRankingRepository(db *gorm.DB) RankingRepository {
	return &rankingRepository{db: db}
}

// Create creates a new ranking in the database
func (r *rankingRepository) Create(ranking *models.Ranking) error {
	return r.db.Create(ranking).Error
}

// GetByUserID retrieves all rankings for a specific user
func (r *rankingRepository) GetByUserID(userID uint) ([]*models.Ranking, error) {
	var rankings []*models.Ranking
	err := r.db.Preload("Book").
		Where("user_id = ?", userID).
		Order("score DESC").
		Find(&rankings).Error
	return rankings, err
}

// GetByUserAndBook retrieves a specific ranking for a user and book
func (r *rankingRepository) GetByUserAndBook(userID, bookID uint) (*models.Ranking, error) {
	var ranking models.Ranking
	err := r.db.Where("user_id = ? AND book_id = ?", userID, bookID).
		First(&ranking).Error
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}

// Update updates a ranking
func (r *rankingRepository) Update(ranking *models.Ranking) error {
	return r.db.Save(ranking).Error
}

// GetTopRanked retrieves the highest-ranked books for a user
func (r *rankingRepository) GetTopRanked(userID uint, limit int) ([]*models.Ranking, error) {
	var rankings []*models.Ranking
	err := r.db.Preload("Book").
		Where("user_id = ?", userID).
		Order("score DESC").
		Limit(limit).
		Find(&rankings).Error
	return rankings, err
}

// GetAllByBookID retrieves all rankings for a specific book across all users
func (r *rankingRepository) GetAllByBookID(bookID uint) ([]*models.Ranking, error) {
	var rankings []*models.Ranking
	err := r.db.Preload("User").
		Where("book_id = ?", bookID).
		Find(&rankings).Error
	return rankings, err
}

// GetAllUserRankings retrieves all rankings grouped by user ID
func (r *rankingRepository) GetAllUserRankings() (map[uint][]*models.Ranking, error) {
	var rankings []*models.Ranking
	err := r.db.Preload("Book").Find(&rankings).Error
	if err != nil {
		return nil, err
	}

	userRankings := make(map[uint][]*models.Ranking)
	for _, ranking := range rankings {
		userRankings[ranking.UserID] = append(userRankings[ranking.UserID], ranking)
	}

	return userRankings, nil
}

// GetUsersByGenre finds users who have rated books in a specific genre
func (r *rankingRepository) GetUsersByGenre(genre string) ([]uint, error) {
	var userIDs []uint
	err := r.db.Table("rankings").
		Select("DISTINCT user_id").
		Joins("JOIN books ON rankings.book_id = books.id").
		Where("books.genre = ?", genre).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// GetAverageRatingByGenre calculates average rating per genre for a user
func (r *rankingRepository) GetAverageRatingByGenre(userID uint) (map[string]float64, error) {
	type GenreAvg struct {
		Genre   string  `json:"genre"`
		Average float64 `json:"average"`
	}

	var results []GenreAvg
	err := r.db.Table("rankings").
		Select("books.genre, AVG(rankings.score) as average").
		Joins("JOIN books ON rankings.book_id = books.id").
		Where("rankings.user_id = ?", userID).
		Group("books.genre").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	genreAvgs := make(map[string]float64)
	for _, result := range results {
		genreAvgs[result.Genre] = result.Average
	}

	return genreAvgs, nil
}

// GetUnratedBooksByUser finds books that a user hasn't rated yet
func (r *rankingRepository) GetUnratedBooksByUser(userID uint) ([]*models.Book, error) {
	var books []*models.Book
	err := r.db.Table("books").
		Where("id NOT IN (?)",
			r.db.Table("rankings").
				Select("book_id").
				Where("user_id = ?", userID)).
		Find(&books).Error
	return books, err
}

// GetTopRatedBooksByGenre finds top-rated books in a specific genre
func (r *rankingRepository) GetTopRatedBooksByGenre(genre string, limit int) ([]*models.Book, error) {
	var books []*models.Book
	err := r.db.Table("books").
		Select("books.*, AVG(rankings.score) as avg_score").
		Joins("JOIN rankings ON books.id = rankings.book_id").
		Where("books.genre = ?", genre).
		Group("books.id").
		Order("avg_score DESC").
		Limit(limit).
		Find(&books).Error
	return books, err
}

// GetUsersWhoRatedBook finds all users who have rated a specific book
func (r *rankingRepository) GetUsersWhoRatedBook(bookID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Table("rankings").
		Select("user_id").
		Where("book_id = ?", bookID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}