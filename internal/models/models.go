// Package models contains all GORM model definitions for the BookRank application
package models

import "gorm.io/gorm"

// AllModels returns a slice of all model structs for auto-migration
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Book{},
		&Ranking{},
		&Comparison{},
		&Friendship{},
		&BookMetadata{},
	}
}

// AutoMigrate runs auto-migration for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}

// CreateIndexes creates additional indexes for better performance
func CreateIndexes(db *gorm.DB) error {
	// Create composite indexes for better query performance

	// Composite index for user rankings queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_rankings_user_score ON rankings(user_id, score DESC)").Error; err != nil {
		return err
	}

	// Composite index for book rankings queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_rankings_book_score ON rankings(book_id, score DESC)").Error; err != nil {
		return err
	}

	// Composite index for user comparisons
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comparisons_user_created ON comparisons(user_id, created_at DESC)").Error; err != nil {
		return err
	}

	// Composite index for book comparisons
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comparisons_books ON comparisons(book_a_id, book_b_id)").Error; err != nil {
		return err
	}

	// Composite index for friendships to prevent duplicates
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_friendships_unique ON friendships(user_id, friend_id)").Error; err != nil {
		return err
	}

	// Composite index for book metadata source and external_id
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_book_metadata_source_external ON book_metadata(source, external_id) WHERE external_id IS NOT NULL").Error; err != nil {
		return err
	}

	// Index for book search by title and author
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_books_title_author ON books(title, author)").Error; err != nil {
		return err
	}

	return nil
}