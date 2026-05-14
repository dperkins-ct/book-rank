package book

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookrank/internal/api/dto"
	"bookrank/internal/auth"
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"bookrank/internal/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// MockBookRepository implements BookRepository for testing
type MockBookRepository struct {
	books []models.Book
	nextID uint
}

func NewMockBookRepository() *MockBookRepository {
	return &MockBookRepository{
		books:  make([]models.Book, 0),
		nextID: 1,
	}
}

func (m *MockBookRepository) Create(book *models.Book) error {
	book.ID = m.nextID
	m.nextID++
	m.books = append(m.books, *book)
	return nil
}

func (m *MockBookRepository) GetByID(id uint) (*models.Book, error) {
	for _, book := range m.books {
		if book.ID == id {
			return &book, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockBookRepository) GetByIDWithMetadata(id uint) (*models.Book, error) {
	return m.GetByID(id)
}

func (m *MockBookRepository) GetAll(options repository.BookQueryOptions) ([]*models.Book, int64, error) {
	result := make([]*models.Book, 0)
	for i := range m.books {
		result = append(result, &m.books[i])
	}
	return result, int64(len(result)), nil
}

func (m *MockBookRepository) Update(book *models.Book) error {
	for i, b := range m.books {
		if b.ID == book.ID {
			m.books[i] = *book
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *MockBookRepository) Delete(id uint) error {
	return m.SoftDelete(id)
}

func (m *MockBookRepository) SoftDelete(id uint) error {
	for i, book := range m.books {
		if book.ID == id {
			m.books = append(m.books[:i], m.books[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *MockBookRepository) Search(query string, limit, offset int) ([]*models.Book, int64, error) {
	return m.GetAll(repository.BookQueryOptions{})
}

func (m *MockBookRepository) GetByGenre(genre string, limit, offset int) ([]*models.Book, error) {
	result := make([]*models.Book, 0)
	for i, book := range m.books {
		if book.Genre == genre {
			result = append(result, &m.books[i])
		}
	}
	return result, nil
}

func (m *MockBookRepository) GetByAuthor(author string, limit, offset int) ([]*models.Book, error) {
	result := make([]*models.Book, 0)
	for i, book := range m.books {
		if book.Author == author {
			result = append(result, &m.books[i])
		}
	}
	return result, nil
}

func (m *MockBookRepository) GetAllBooks() ([]*models.Book, error) {
	result := make([]*models.Book, len(m.books))
	for i := range m.books {
		result[i] = &m.books[i]
	}
	return result, nil
}

func (m *MockBookRepository) GetBookStats(bookID uint) (*repository.BookStats, error) {
	return &repository.BookStats{
		BookID:        bookID,
		TotalRatings:  0,
		AverageRating: 0,
	}, nil
}

func (m *MockBookRepository) CreateMetadata(metadata *models.BookMetadata) error {
	return nil
}

func (m *MockBookRepository) UpdateMetadata(metadata *models.BookMetadata) error {
	return nil
}

func (m *MockBookRepository) GetMetadataByBookID(bookID uint) ([]*models.BookMetadata, error) {
	return []*models.BookMetadata{}, nil
}

func (m *MockBookRepository) UpsertMetadata(metadata *models.BookMetadata) error {
	return nil
}

func (m *MockBookRepository) GetBooksByRatingRange(minRating, maxRating int, limit, offset int) ([]*models.Book, error) {
	return []*models.Book{}, nil
}

func (m *MockBookRepository) Count() (int64, error) {
	return int64(len(m.books)), nil
}

func setupTestHandlers() *Handlers {
	mockRepo := NewMockBookRepository()
	bookService := service.NewBookService(mockRepo)
	return NewHandlers(bookService)
}

func createTestContext(userID uint) context.Context {
	claims := &auth.Claims{
		UserID:   userID,
		Username: "testuser",
	}
	return context.WithValue(context.Background(), "user", claims)
}

func TestCreateBook(t *testing.T) {
	handlers := setupTestHandlers()

	createReq := dto.CreateBookRequest{
		Title:  "Test Book",
		Author: "Test Author",
		Genre:  "Fiction",
	}

	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewBuffer(reqBody))
	req = req.WithContext(createTestContext(1))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handlers.CreateBook(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}

	var response dto.BookResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Title != createReq.Title {
		t.Errorf("Expected title %s, got %s", createReq.Title, response.Title)
	}
	if response.Author != createReq.Author {
		t.Errorf("Expected author %s, got %s", createReq.Author, response.Author)
	}
}

func TestGetBooks(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req = req.WithContext(createTestContext(1))

	rr := httptest.NewRecorder()
	handlers.GetBooks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response dto.BookListResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Books == nil {
		t.Error("Expected books array to be present")
	}
}

func TestGetBook(t *testing.T) {
	handlers := setupTestHandlers()

	// First create a book
	createReq := dto.CreateBookRequest{
		Title:  "Test Book",
		Author: "Test Author",
		Genre:  "Fiction",
	}
	reqBody, _ := json.Marshal(createReq)
	createRequest := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewBuffer(reqBody))
	createRequest = createRequest.WithContext(createTestContext(1))
	createRequest.Header.Set("Content-Type", "application/json")

	createRR := httptest.NewRecorder()
	handlers.CreateBook(createRR, createRequest)

	// Now get the book
	req := httptest.NewRequest(http.MethodGet, "/api/books/1", nil)
	req = req.WithContext(createTestContext(1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handlers.GetBook(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response dto.BookResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Title != createReq.Title {
		t.Errorf("Expected title %s, got %s", createReq.Title, response.Title)
	}
}

func TestSearchBooks(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/books/search?q=test", nil)
	req = req.WithContext(createTestContext(1))

	rr := httptest.NewRecorder()
	handlers.SearchBooks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response dto.BookSearchResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Query != "test" {
		t.Errorf("Expected query %s, got %s", "test", response.Query)
	}
}