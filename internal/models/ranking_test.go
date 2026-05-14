package models

import (
	"fmt"
	"testing"
	"time"
)

func TestRankingModel(t *testing.T) {
	tests := map[string]struct {
		ranking Ranking
		wantErr bool
	}{
		"valid ranking": {
			ranking: Ranking{
				UserID: 1,
				BookID: 1,
				Score:  1600,
			},
			wantErr: false,
		},
		"default score": {
			ranking: Ranking{
				UserID: 1,
				BookID: 2, // Different book to avoid duplicate
				// Score not set, should default to 1500
			},
			wantErr: false,
		},
		"score too high": {
			ranking: Ranking{
				UserID: 1,
				BookID: 3, // Different book
				Score:  3001, // Above max constraint
			},
			wantErr: true,
		},
		"negative score": {
			ranking: Ranking{
				UserID: 1,
				BookID: 4, // Different book
				Score:  -100, // Below min constraint
			},
			wantErr: true,
		},
	}

	db := setupTestDB()

	// Create a user and multiple books first
	user := User{
		Username:     "testuser_rankings",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	books := make([]Book, 4)
	for i := 0; i < 4; i++ {
		book := Book{
			Title:     fmt.Sprintf("Test Book %d", i+1),
			Author:    "Test Author",
			Genre:     "Fiction",
			CreatedBy: user.ID,
		}
		if err := db.Create(&book).Error; err != nil {
			t.Fatalf("Failed to create test book %d: %v", i+1, err)
		}
		books[i] = book
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Set valid foreign keys
			tc.ranking.UserID = user.ID
			if tc.ranking.BookID <= 4 {
				tc.ranking.BookID = books[tc.ranking.BookID-1].ID
			}

			err := db.Create(&tc.ranking).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamp was set
				if tc.ranking.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should be set")
				}

				// Verify default score if it was 0
				if name == "default score" && tc.ranking.Score != 1500 {
					t.Errorf("Score should default to 1500, got %d", tc.ranking.Score)
				}
			}
		})
	}
}

func TestRankingBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	// Create user and books
	user := User{
		Username:     "testuser_hooks",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book1 := Book{
		Title:     "Test Book Hook 1",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("Failed to create test book 1: %v", err)
	}

	book2 := Book{
		Title:     "Test Book Hook 2",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("Failed to create test book 2: %v", err)
	}

	t.Run("sets default score", func(t *testing.T) {
		ranking := Ranking{
			UserID: user.ID,
			BookID: book1.ID,
			// Score not set
		}

		err := db.Create(&ranking).Error
		if err != nil {
			t.Fatalf("Failed to create ranking: %v", err)
		}

		if ranking.Score != 1500 {
			t.Errorf("Score should default to 1500, got %d", ranking.Score)
		}

		if ranking.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should be set by BeforeCreate hook")
		}
	})

	t.Run("preserves existing score", func(t *testing.T) {
		ranking := Ranking{
			UserID: user.ID,
			BookID: book2.ID, // Use different book
			Score:  1800,
		}

		err := db.Create(&ranking).Error
		if err != nil {
			t.Fatalf("Failed to create ranking: %v", err)
		}

		if ranking.Score != 1800 {
			t.Errorf("Score should preserve existing value 1800, got %d", ranking.Score)
		}
	})
}

func TestRankingBeforeUpdateHook(t *testing.T) {
	db := setupTestDB()

	// Create user and book
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	ranking := Ranking{
		UserID: user.ID,
		BookID: book.ID,
		Score:  1500,
	}

	// Create ranking
	err := db.Create(&ranking).Error
	if err != nil {
		t.Fatalf("Failed to create ranking: %v", err)
	}

	originalUpdatedAt := ranking.UpdatedAt

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update ranking
	ranking.Score = 1700
	err = db.Save(&ranking).Error
	if err != nil {
		t.Fatalf("Failed to update ranking: %v", err)
	}

	if !ranking.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated by BeforeUpdate hook")
	}
}

func TestRankingTableName(t *testing.T) {
	ranking := Ranking{}
	expected := "rankings"
	if ranking.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", ranking.TableName(), expected)
	}
}

func TestRankingRelationships(t *testing.T) {
	db := setupTestDB()

	// Create user and book
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	ranking := Ranking{
		UserID: user.ID,
		BookID: book.ID,
		Score:  1600,
	}
	if err := db.Create(&ranking).Error; err != nil {
		t.Fatalf("Failed to create ranking: %v", err)
	}

	// Test preloading relationships
	var rankingWithRelations Ranking
	err := db.Preload("User").Preload("Book").Where("user_id = ? AND book_id = ?", user.ID, book.ID).First(&rankingWithRelations).Error
	if err != nil {
		t.Fatalf("Failed to load ranking with relations: %v", err)
	}

	if rankingWithRelations.User.ID != user.ID {
		t.Errorf("User ID = %v, want %v", rankingWithRelations.User.ID, user.ID)
	}

	if rankingWithRelations.Book.ID != book.ID {
		t.Errorf("Book ID = %v, want %v", rankingWithRelations.Book.ID, book.ID)
	}
}

func TestRankingCompositeKey(t *testing.T) {
	db := setupTestDB()

	// Create user and book
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	// Create first ranking
	ranking1 := Ranking{
		UserID: user.ID,
		BookID: book.ID,
		Score:  1500,
	}
	if err := db.Create(&ranking1).Error; err != nil {
		t.Fatalf("Failed to create first ranking: %v", err)
	}

	// Try to create duplicate ranking (should fail due to composite primary key)
	ranking2 := Ranking{
		UserID: user.ID,
		BookID: book.ID,
		Score:  1600,
	}
	err := db.Create(&ranking2).Error
	if err == nil {
		t.Error("Should not be able to create duplicate ranking with same user_id and book_id")
	}
}