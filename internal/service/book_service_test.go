package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"testing"

	"gorm.io/gorm"
)

// MockBookRepository for service testing
type MockBookRepository struct {
	books      map[uint]*models.Book
	nextID     uint
	createErr  error
	getErr     error
	updateErr  error
	deleteErr  error
	searchErr  error
}

func NewMockBookRepository() *MockBookRepository {
	return &MockBookRepository{
		books:  make(map[uint]*models.Book),
		nextID: 1,
	}
}

func (m *MockBookRepository) Create(book *models.Book) error {
	if m.createErr != nil {
		return m.createErr
	}
	book.ID = m.nextID
	m.nextID++
	m.books[book.ID] = book
	return nil
}

func (m *MockBookRepository) GetByID(id uint) (*models.Book, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	book, exists := m.books[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return book, nil
}

func (m *MockBookRepository) GetByIDWithMetadata(id uint) (*models.Book, error) {
	return m.GetByID(id)
}

func (m *MockBookRepository) GetAll(options repository.BookQueryOptions) ([]*models.Book, int64, error) {
	var result []*models.Book
	for _, book := range m.books {
		result = append(result, book)
	}
	return result, int64(len(result)), nil
}

func (m *MockBookRepository) GetAllBooks() ([]*models.Book, error) {
	var result []*models.Book
	for _, book := range m.books {
		result = append(result, book)
	}
	return result, nil
}

func (m *MockBookRepository) Update(book *models.Book) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.books[book.ID] = book
	return nil
}

func (m *MockBookRepository) Delete(id uint) error {
	return m.SoftDelete(id)
}

func (m *MockBookRepository) SoftDelete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.books, id)
	return nil
}

func (m *MockBookRepository) Search(query string, limit, offset int) ([]*models.Book, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.GetAll(repository.BookQueryOptions{})
}

func (m *MockBookRepository) GetByGenre(genre string, limit, offset int) ([]*models.Book, error) {
	return nil, nil
}

func (m *MockBookRepository) GetByAuthor(author string, limit, offset int) ([]*models.Book, error) {
	return nil, nil
}

func (m *MockBookRepository) GetBookStats(bookID uint) (*repository.BookStats, error) {
	return &repository.BookStats{BookID: bookID}, nil
}

func (m *MockBookRepository) CreateMetadata(metadata *models.BookMetadata) error {
	return nil
}

func (m *MockBookRepository) UpdateMetadata(metadata *models.BookMetadata) error {
	return nil
}

func (m *MockBookRepository) GetMetadataByBookID(bookID uint) ([]*models.BookMetadata, error) {
	return nil, nil
}

func (m *MockBookRepository) GetBooksByRatingRange(minRating, maxRating int, limit, offset int) ([]*models.Book, error) {
	return nil, nil
}

func (m *MockBookRepository) Count() (int64, error) {
	return int64(len(m.books)), nil
}

func TestBookService_CreateBook(t *testing.T) {
	tests := map[string]struct {
		request *BookCreateRequest
		userID  uint
		repoErr error
		wantErr bool
	}{
		"valid_book": {
			request: &BookCreateRequest{
				Title:  "Test Book",
				Author: "Test Author",
				Genre:  "Fiction",
			},
			userID:  1,
			wantErr: false,
		},
		"empty_title": {
			request: &BookCreateRequest{
				Title:  "",
				Author: "Test Author",
				Genre:  "Fiction",
			},
			userID:  1,
			wantErr: true,
		},
		"repository_error": {
			request: &BookCreateRequest{
				Title:  "Test Book",
				Author: "Test Author",
				Genre:  "Fiction",
			},
			userID:  1,
			repoErr: gorm.ErrInvalidDB,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := NewMockBookRepository()
			mockRepo.createErr = tc.repoErr
			service := NewBookService(mockRepo)

			book, err := service.CreateBook(tc.request, tc.userID)
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateBook() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if book.Title != tc.request.Title {
					t.Errorf("Expected title %s, got %s", tc.request.Title, book.Title)
				}
				if book.CreatedBy != tc.userID {
					t.Errorf("Expected CreatedBy %d, got %d", tc.userID, book.CreatedBy)
				}
			}
		})
	}
}

func TestBookService_GetBookByID(t *testing.T) {
	mockRepo := NewMockBookRepository()
	service := NewBookService(mockRepo)

	// Create test book
	testBook := &models.Book{
		ID:     1,
		Title:  "Test Book",
		Author: "Test Author",
		Genre:  "Fiction",
	}
	mockRepo.books[1] = testBook

	tests := map[string]struct {
		id      uint
		repoErr error
		wantErr bool
	}{
		"existing_book": {
			id:      1,
			wantErr: false,
		},
		"nonexistent_book": {
			id:      999,
			wantErr: true,
		},
		"repository_error": {
			id:      1,
			repoErr: gorm.ErrInvalidDB,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo.getErr = tc.repoErr
			book, err := service.GetBookByID(tc.id)
			mockRepo.getErr = nil

			if (err != nil) != tc.wantErr {
				t.Errorf("GetBookByID() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && book.ID != tc.id {
				t.Errorf("Expected book ID %d, got %d", tc.id, book.ID)
			}
		})
	}
}

func TestBookService_UpdateBook(t *testing.T) {
	mockRepo := NewMockBookRepository()
	service := NewBookService(mockRepo)

	// Create test book
	testBook := &models.Book{
		ID:        1,
		Title:     "Original Title",
		Author:    "Original Author",
		Genre:     "Fiction",
		CreatedBy: 1,
	}
	mockRepo.books[1] = testBook

	newTitle := "Updated Title"
	tests := map[string]struct {
		id      uint
		request *BookUpdateRequest
		userID  uint
		wantErr bool
	}{
		"valid_update": {
			id: 1,
			request: &BookUpdateRequest{
				Title: &newTitle,
			},
			userID:  1,
			wantErr: false,
		},
		"unauthorized_update": {
			id: 1,
			request: &BookUpdateRequest{
				Title: &newTitle,
			},
			userID:  2,
			wantErr: true,
		},
		"nonexistent_book": {
			id: 999,
			request: &BookUpdateRequest{
				Title: &newTitle,
			},
			userID:  1,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			book, err := service.UpdateBook(tc.id, tc.request, tc.userID)
			if (err != nil) != tc.wantErr {
				t.Errorf("UpdateBook() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && tc.request.Title != nil && book.Title != *tc.request.Title {
				t.Errorf("Expected title %s, got %s", *tc.request.Title, book.Title)
			}
		})
	}
}

func TestBookService_DeleteBook(t *testing.T) {
	mockRepo := NewMockBookRepository()
	service := NewBookService(mockRepo)

	// Create test book
	testBook := &models.Book{
		ID:        1,
		Title:     "Test Book",
		Author:    "Test Author",
		Genre:     "Fiction",
		CreatedBy: 1,
	}
	mockRepo.books[1] = testBook

	tests := map[string]struct {
		id      uint
		userID  uint
		wantErr bool
	}{
		"valid_delete": {
			id:      1,
			userID:  1,
			wantErr: false,
		},
		"unauthorized_delete": {
			id:      1,
			userID:  2,
			wantErr: true,
		},
		"nonexistent_book": {
			id:      999,
			userID:  1,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := service.DeleteBook(tc.id, tc.userID)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteBook() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestBookService_SearchBooks(t *testing.T) {
	mockRepo := NewMockBookRepository()
	service := NewBookService(mockRepo)

	tests := map[string]struct {
		query   string
		limit   int
		offset  int
		wantErr bool
	}{
		"valid_search": {
			query:   "test",
			limit:   10,
			offset:  0,
			wantErr: false,
		},
		"empty_query": {
			query:   "",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
		"whitespace_query": {
			query:   "   ",
			limit:   10,
			offset:  0,
			wantErr: true,
		},
		"limit_adjustment": {
			query:   "test",
			limit:   0,
			offset:  0,
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := service.SearchBooks(tc.query, tc.limit, tc.offset)
			if (err != nil) != tc.wantErr {
				t.Errorf("SearchBooks() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if result == nil {
					t.Error("Expected search result to be non-nil")
				}
				if tc.limit <= 0 && result.PageSize != 20 {
					t.Errorf("Expected default page size 20, got %d", result.PageSize)
				}
			}
		})
	}
}

func TestBookService_GetBooks(t *testing.T) {
	mockRepo := NewMockBookRepository()
	service := NewBookService(mockRepo)

	// Add some test books
	for i := 1; i <= 5; i++ {
		book := &models.Book{
			ID:     uint(i),
			Title:  "Test Book " + string(rune(i)),
			Author: "Test Author",
			Genre:  "Fiction",
		}
		mockRepo.books[uint(i)] = book
	}

	tests := map[string]struct {
		options      repository.BookQueryOptions
		expectedSize int
		wantErr      bool
	}{
		"default_options": {
			options:      repository.BookQueryOptions{},
			expectedSize: 20, // default page size
			wantErr:      false,
		},
		"custom_limit": {
			options:      repository.BookQueryOptions{Limit: 10},
			expectedSize: 10,
			wantErr:      false,
		},
		"limit_too_high": {
			options:      repository.BookQueryOptions{Limit: 200},
			expectedSize: 20, // should be capped at default
			wantErr:      false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := service.GetBooks(tc.options)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetBooks() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if result.PageSize != tc.expectedSize {
					t.Errorf("Expected page size %d, got %d", tc.expectedSize, result.PageSize)
				}
			}
		})
	}
}