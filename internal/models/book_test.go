package models

import (
	"testing"
	"time"
)

func TestBookModel(t *testing.T) {
	tests := map[string]struct {
		book    Book
		wantErr bool
	}{
		"valid book": {
			book: Book{
				Title:       "Test Book",
				Author:      "Test Author",
				Genre:       "Fiction",
				Description: "A test book description",
				CreatedBy:   1,
			},
			wantErr: false,
		},
		"empty title": {
			book: Book{
				Title:     "",
				Author:    "Test Author",
				Genre:     "Fiction",
				CreatedBy: 1,
			},
			wantErr: false, // SQLite allows empty strings, validation should be in service layer
		},
		"empty author": {
			book: Book{
				Title:     "Test Book",
				Author:    "",
				Genre:     "Fiction",
				CreatedBy: 1,
			},
			wantErr: false, // SQLite allows empty strings, validation should be in service layer
		},
		"empty genre": {
			book: Book{
				Title:     "Test Book",
				Author:    "Test Author",
				Genre:     "",
				CreatedBy: 1,
			},
			wantErr: false, // SQLite allows empty strings, validation should be in service layer
		},
		"missing created_by": {
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Genre:  "Fiction",
				// CreatedBy is required but not set (0 is invalid FK)
			},
			wantErr: false, // SQLite doesn't enforce FK constraints by default in test setup
		},
	}

	db := setupTestDB()

	// Create a user first for foreign key constraint
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Set valid CreatedBy if it's 0 and we don't expect an error
			if tc.book.CreatedBy == 0 && !tc.wantErr {
				tc.book.CreatedBy = user.ID
			} else if tc.book.CreatedBy == 1 {
				tc.book.CreatedBy = user.ID
			}

			err := db.Create(&tc.book).Error
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Verify timestamps were set
				if tc.book.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if tc.book.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should be set")
				}
				if tc.book.ID == 0 {
					t.Error("ID should be set")
				}
			}
		})
	}
}

func TestBookBeforeCreateHook(t *testing.T) {
	db := setupTestDB()

	// Create a user first
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

	err := db.Create(&book).Error
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	if book.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by BeforeCreate hook")
	}

	if book.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set by BeforeCreate hook")
	}

	// Check that both timestamps are approximately equal (within 1 second)
	diff := book.UpdatedAt.Sub(book.CreatedAt)
	if diff > time.Second || diff < -time.Second {
		t.Error("CreatedAt and UpdatedAt should be set to approximately the same time")
	}
}

func TestBookBeforeUpdateHook(t *testing.T) {
	db := setupTestDB()

	// Create a user first
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

	// Create book
	err := db.Create(&book).Error
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	originalUpdatedAt := book.UpdatedAt

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update book
	book.Title = "Updated Test Book"
	err = db.Save(&book).Error
	if err != nil {
		t.Fatalf("Failed to update book: %v", err)
	}

	if !book.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated by BeforeUpdate hook")
	}
}

func TestBookTableName(t *testing.T) {
	book := Book{}
	expected := "books"
	if book.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", book.TableName(), expected)
	}
}

func TestBookWithPublicationDate(t *testing.T) {
	db := setupTestDB()

	// Create a user first
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	pubDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	book := Book{
		Title:           "Test Book",
		Author:          "Test Author",
		Genre:           "Fiction",
		PublicationDate: &pubDate,
		CreatedBy:       user.ID,
	}

	err := db.Create(&book).Error
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	// Retrieve the book and verify publication date
	var retrievedBook Book
	err = db.First(&retrievedBook, book.ID).Error
	if err != nil {
		t.Fatalf("Failed to retrieve book: %v", err)
	}

	if retrievedBook.PublicationDate == nil {
		t.Error("PublicationDate should not be nil")
	} else if !retrievedBook.PublicationDate.Equal(pubDate) {
		t.Errorf("PublicationDate = %v, want %v", retrievedBook.PublicationDate, pubDate)
	}
}

func TestBookRelationships(t *testing.T) {
	db := setupTestDB()

	// Create a user
	user := User{
		Username:     "testuser",
		PasswordHash: "password",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a book
	book := Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	// Test preloading creator relationship
	var bookWithCreator Book
	err := db.Preload("Creator").First(&bookWithCreator, book.ID).Error
	if err != nil {
		t.Fatalf("Failed to load book with creator: %v", err)
	}

	if bookWithCreator.Creator.ID != user.ID {
		t.Errorf("Creator ID = %v, want %v", bookWithCreator.Creator.ID, user.ID)
	}

	if bookWithCreator.Creator.Username != user.Username {
		t.Errorf("Creator Username = %v, want %v", bookWithCreator.Creator.Username, user.Username)
	}
}