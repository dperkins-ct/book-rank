package repository

import (
	"bookrank/internal/models"
	"gorm.io/gorm"
)

// UserStats represents user statistics
type UserStats struct {
	UserID          uint `json:"user_id"`
	TotalRankings   int  `json:"total_rankings"`
	TotalComparisons int `json:"total_comparisons"`
	AverageRating   float64 `json:"average_rating"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetAll(limit, offset int) ([]*models.User, error)
	Search(query string, limit, offset int) ([]*models.User, error)
	UpdateLastLogin(userID uint) error
	GetUserStats(userID uint) (*UserStats, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user's information
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAll retrieves all users with pagination
func (r *userRepository) GetAll(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	query := r.db.Model(&models.User{})

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&users).Error
	return users, err
}

// Search searches for users by username or email
func (r *userRepository) Search(query string, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	dbQuery := r.db.Model(&models.User{}).Where("username ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%")

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	err := dbQuery.Find(&users).Error
	return users, err
}

// UpdateLastLogin updates the user's last login timestamp
func (r *userRepository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("updated_at", "now()").Error
}

// GetUserStats retrieves user statistics
func (r *userRepository) GetUserStats(userID uint) (*UserStats, error) {
	stats := &UserStats{UserID: userID}

	// Count total rankings
	r.db.Model(&models.Ranking{}).Where("user_id = ?", userID).Count(&[]int64{int64(stats.TotalRankings)}[0])

	// Count total comparisons
	r.db.Model(&models.Comparison{}).Where("user_id = ?", userID).Count(&[]int64{int64(stats.TotalComparisons)}[0])

	// Calculate average rating (this is a simplified version)
	r.db.Model(&models.Ranking{}).Where("user_id = ?", userID).Select("AVG(rating)").Scan(&stats.AverageRating)

	return stats, nil
}