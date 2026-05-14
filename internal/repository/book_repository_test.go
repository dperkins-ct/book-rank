package repository

import (
	"bookrank/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.BookMetadata{},
		&models.Ranking{},
		&models.Comparison{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestUser(db *gorm.DB) *models.User {
	user := &models.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
	}
	db.Create(user)
	return user
}

func TestBookRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	tests := map[string]struct {
		book    *models.Book
		wantErr bool
	}{
		"valid_book": {
			book: &models.Book{
				Title:     "Test Book",
				Author:    "Test Author",
				Genre:     "Fiction",
				CreatedBy: user.ID,
			},
			wantErr: false,
		},
		"missing_title": {
			book: &models.Book{
				Author:    "Test Author",
				Genre:     "Fiction",
				CreatedBy: user.ID,
			},
			wantErr: true,
		},
		"missing_author": {
			book: &models.Book{
				Title:     "Test Book",
				Genre:     "Fiction",
				CreatedBy: user.ID,
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := repo.Create(tc.book)
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && tc.book.ID == 0 {
				t.Error("Expected book ID to be set after creation")
			}
		})
	}
}

func TestBookRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	// Create test book
	testBook := &models.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	repo.Create(testBook)

	tests := map[string]struct {
		id      uint
		wantErr bool
	}{
		"existing_book": {
			id:      testBook.ID,
			wantErr: false,
		},
		"non_existing_book": {
			id:      9999,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			book, err := repo.GetByID(tc.id)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if book.ID != tc.id {
					t.Errorf("Expected book ID %d, got %d", tc.id, book.ID)
				}
				if book.Creator.ID != user.ID {
					t.Error("Expected creator to be preloaded")
				}
			}
		})
	}
}

func TestBookRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	// Create test books
	books := []*models.Book{
		{Title: "Book A", Author: "Author A", Genre: "Fiction", CreatedBy: user.ID},
		{Title: "Book B", Author: "Author B", Genre: "Science", CreatedBy: user.ID},
		{Title: "Book C", Author: "Author A", Genre: "Fiction", CreatedBy: user.ID},
	}
	for _, book := range books {
		repo.Create(book)
	}

	tests := map[string]struct {
		options BookQueryOptions
		wantLen int
		wantErr bool
	}{
		"no_filters": {
			options: BookQueryOptions{Limit: 10, Offset: 0},
			wantLen: 3,
			wantErr: false,
		},
		"genre_filter": {
			options: BookQueryOptions{
				Limit:  10,
				Offset: 0,
				Filter: &BookFilter{Genre: "Fiction"},
			},
			wantLen: 2,
			wantErr: false,
		},
		"author_filter": {
			options: BookQueryOptions{
				Limit:  10,
				Offset: 0,
				Filter: &BookFilter{Author: "Author A"},
			},
			wantLen: 2,
			wantErr: false,
		},
		"pagination": {
			options: BookQueryOptions{Limit: 2, Offset: 1},
			wantLen: 2,
			wantErr: false,
		},
		"search_filter": {
			options: BookQueryOptions{
				Limit:  10,
				Offset: 0,
				Filter: &BookFilter{Search: "Book A"},
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, total, err := repo.GetAll(tc.options)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if len(result) != tc.wantLen {
					t.Errorf("Expected %d books, got %d", tc.wantLen, len(result))
				}
				if total < int64(tc.wantLen) {
					t.Errorf("Expected total >= %d, got %d", tc.wantLen, total)
				}
			}
		})
	}
}

func TestBookRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	// Create test books
	books := []*models.Book{
		{Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Genre: "Fiction", CreatedBy: user.ID},
		{Title: "To Kill a Mockingbird", Author: "Harper Lee", Genre: "Fiction", CreatedBy: user.ID},
		{Title: "1984", Author: "George Orwell", Genre: "Dystopian", CreatedBy: user.ID},
	}
	for _, book := range books {
		repo.Create(book)
	}

	tests := map[string]struct {
		query   string
		limit   int
		offset  int
		wantLen int
		wantErr bool
	}{
		"title_search": {
			query:   "Great",
			limit:   10,
			offset:  0,
			wantLen: 1,
			wantErr: false,
		},
		"author_search": {
			query:   "Orwell",
			limit:   10,
			offset:  0,
			wantLen: 1,
			wantErr: false,
		},
		"partial_match": {
			query:   "Kill",
			limit:   10,
			offset:  0,
			wantLen: 1,
			wantErr: false,
		},
		"no_results": {
			query:   "NonexistentBook",
			limit:   10,
			offset:  0,
			wantLen: 0,
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, total, err := repo.Search(tc.query, tc.limit, tc.offset)
			if (err != nil) != tc.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if len(result) != tc.wantLen {
					t.Errorf("Expected %d books, got %d", tc.wantLen, len(result))
				}
				if total != int64(tc.wantLen) {
					t.Errorf("Expected total %d, got %d", tc.wantLen, total)
				}
			}
		})
	}
}

func TestBookRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	// Create test book
	testBook := &models.Book{
		Title:     "Original Title",
		Author:    "Original Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	repo.Create(testBook)

	tests := map[string]struct {
		book    *models.Book
		wantErr bool
	}{
		"valid_update": {
			book: &models.Book{
				ID:        testBook.ID,
				Title:     "Updated Title",
				Author:    "Updated Author",
				Genre:     "Updated Genre",
				CreatedBy: user.ID,
			},
			wantErr: false,
		},
		"nonexistent_book": {
			book: &models.Book{
				ID:        9999,
				Title:     "Updated Title",
				Author:    "Updated Author",
				Genre:     "Fiction",
				CreatedBy: user.ID,
			},
			wantErr: false, // GORM doesn't error on updating non-existent records
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := repo.Update(tc.book)
			if (err != nil) != tc.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && tc.book.ID == testBook.ID {
				// Verify update worked
				updated, err := repo.GetByID(tc.book.ID)
				if err != nil {
					t.Errorf("Failed to fetch updated book: %v", err)
				} else if updated.Title != tc.book.Title {
					t.Errorf("Expected title %s, got %s", tc.book.Title, updated.Title)
				}
			}
		})
	}
}

func TestBookRepository_GetBookStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	// Create test book
	testBook := &models.Book{
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: user.ID,
	}
	repo.Create(testBook)

	// Create test rankings
	rankings := []*models.Ranking{
		{UserID: user.ID, BookID: testBook.ID, Score: 1500},
		{UserID: user.ID, BookID: testBook.ID, Score: 1600},
		{UserID: user.ID, BookID: testBook.ID, Score: 1400},
	}
	for _, ranking := range rankings {
		db.Create(ranking)
	}

	tests := map[string]struct {
		bookID  uint
		wantErr bool
	}{
		"existing_book_with_stats": {
			bookID:  testBook.ID,
			wantErr: false,
		},
		"nonexistent_book": {
			bookID:  9999,
			wantErr: false, // Stats calculation doesn't fail for non-existent books
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stats, err := repo.GetBookStats(tc.bookID)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetBookStats() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && tc.bookID == testBook.ID {
				if stats.BookID != tc.bookID {
					t.Errorf("Expected book ID %d, got %d", tc.bookID, stats.BookID)
				}
				if stats.TotalRatings != 3 {
					t.Errorf("Expected 3 total ratings, got %d", stats.TotalRatings)
				}
			}
		})
	}
}

func TestBookRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBookRepository(db)
	user := createTestUser(db)

	tests := map[string]struct {
		setupBooks int
		wantCount  int64
		wantErr    bool
	}{
		"empty_database": {
			setupBooks: 0,
			wantCount:  0,
			wantErr:    false,
		},
		"with_books": {
			setupBooks: 5,
			wantCount:  5,
			wantErr:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Clean database
			db.Exec("DELETE FROM books")

			// Create test books
			for i := 0; i < tc.setupBooks; i++ {
				book := &models.Book{
					Title:     "Book " + string(rune(i)),
					Author:    "Author " + string(rune(i)),
					Genre:     "Fiction",
					CreatedBy: user.ID,
				}
				repo.Create(book)
			}

			count, err := repo.Count()
			if (err != nil) != tc.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && count != tc.wantCount {
				t.Errorf("Expected count %d, got %d", tc.wantCount, count)
			}
		})
	}
}