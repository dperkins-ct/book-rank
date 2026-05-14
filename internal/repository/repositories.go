package repository

import (
	"gorm.io/gorm"
)

// Repositories holds all data access repositories
type Repositories struct {
	User       UserRepository
	Book       BookRepository
	Ranking    RankingRepository
	Comparison ComparisonRepository
}

// NewRepositories creates a new Repositories instance
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:       NewUserRepository(db),
		Book:       NewBookRepository(db),
		Ranking:    NewRankingRepository(db),
		Comparison: NewComparisonRepository(db),
	}
}