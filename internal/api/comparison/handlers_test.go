package comparison

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookrank/internal/auth"
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"bookrank/internal/service"

	"gorm.io/gorm"
)

// MockComparisonRepository implements ComparisonRepository for testing
type MockComparisonRepository struct {
	comparisons []models.Comparison
	nextID      uint
}

func NewMockComparisonRepository() *MockComparisonRepository {
	return &MockComparisonRepository{
		comparisons: make([]models.Comparison, 0),
		nextID:      1,
	}
}

func (m *MockComparisonRepository) Create(comparison *models.Comparison) error {
	comparison.ID = m.nextID
	m.nextID++
	m.comparisons = append(m.comparisons, *comparison)
	return nil
}

func (m *MockComparisonRepository) GetByUserAndBooks(userID, bookAID, bookBID uint) (*models.Comparison, error) {
	for _, comp := range m.comparisons {
		if comp.UserID == userID &&
		   ((comp.BookAID == bookAID && comp.BookBID == bookBID) ||
			(comp.BookAID == bookBID && comp.BookBID == bookAID)) {
			return &comp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockComparisonRepository) GetByUserID(userID uint) ([]*models.Comparison, error) {
	result := make([]*models.Comparison, 0)
	for i := range m.comparisons {
		if m.comparisons[i].UserID == userID {
			result = append(result, &m.comparisons[i])
		}
	}
	return result, nil
}

func (m *MockComparisonRepository) GetUserComparisonsForBook(userID, bookID uint) ([]*models.Comparison, error) {
	result := make([]*models.Comparison, 0)
	for i := range m.comparisons {
		if m.comparisons[i].UserID == userID &&
		   (m.comparisons[i].BookAID == bookID || m.comparisons[i].BookBID == bookID) {
			result = append(result, &m.comparisons[i])
		}
	}
	return result, nil
}

func (m *MockComparisonRepository) GetPendingComparisons(userID uint, limit int) ([]repository.PendingComparison, error) {
	// Return some mock pending comparisons
	return []repository.PendingComparison{
		{
			BookA: models.Book{ID: 1, Title: "Book A", Author: "Author A"},
			BookB: models.Book{ID: 2, Title: "Book B", Author: "Author B"},
		},
	}, nil
}

func (m *MockComparisonRepository) HasUserComparedBooks(userID, bookAID, bookBID uint) (bool, error) {
	_, err := m.GetByUserAndBooks(userID, bookAID, bookBID)
	return err == nil, nil
}

// MockRankingRepository for testing
type MockRankingRepository struct {
	rankings []models.Ranking
}

func NewMockRankingRepository() *MockRankingRepository {
	return &MockRankingRepository{
		rankings: []models.Ranking{
			{UserID: 1, BookID: 1, Score: 1500},
			{UserID: 1, BookID: 2, Score: 1500},
		},
	}
}

func (m *MockRankingRepository) Create(ranking *models.Ranking) error {
	m.rankings = append(m.rankings, *ranking)
	return nil
}

func (m *MockRankingRepository) GetByUserAndBook(userID, bookID uint) (*models.Ranking, error) {
	for i := range m.rankings {
		if m.rankings[i].UserID == userID && m.rankings[i].BookID == bookID {
			return &m.rankings[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockRankingRepository) GetByUserID(userID uint) ([]*models.Ranking, error) {
	result := make([]*models.Ranking, 0)
	for i := range m.rankings {
		if m.rankings[i].UserID == userID {
			result = append(result, &m.rankings[i])
		}
	}
	return result, nil
}

func (m *MockRankingRepository) Update(ranking *models.Ranking) error {
	for i, r := range m.rankings {
		if r.UserID == ranking.UserID && r.BookID == ranking.BookID {
			m.rankings[i] = *ranking
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *MockRankingRepository) GetTopRanked(userID uint, limit int) ([]*models.Ranking, error) {
	return m.GetByUserID(userID)
}

// Additional required methods for RankingRepository interface
func (m *MockRankingRepository) GetAllUserRankings() (map[uint][]*models.Ranking, error) {
	result := make(map[uint][]*models.Ranking)
	for i := range m.rankings {
		userID := m.rankings[i].UserID
		result[userID] = append(result[userID], &m.rankings[i])
	}
	return result, nil
}

func (m *MockRankingRepository) GetUsersByGenre(genre string) ([]uint, error) {
	return []uint{1}, nil
}

func (m *MockRankingRepository) GetAverageRatingByGenre(userID uint) (map[string]float64, error) {
	return map[string]float64{"Fiction": 1500.0}, nil
}

func (m *MockRankingRepository) GetUnratedBooksByUser(userID uint) ([]*models.Book, error) {
	return []*models.Book{}, nil
}

func (m *MockRankingRepository) GetTopRatedBooksByGenre(genre string, limit int) ([]*models.Book, error) {
	return []*models.Book{}, nil
}

func (m *MockRankingRepository) GetUsersWhoRatedBook(bookID uint) ([]uint, error) {
	return []uint{1}, nil
}

// MockBookRepository for testing
type MockBookRepository struct{}

func (m *MockBookRepository) GetByID(id uint) (*models.Book, error) {
	return &models.Book{ID: id, Title: "Mock Book", Author: "Mock Author"}, nil
}

// Implement other required methods (simplified for testing)
func (m *MockBookRepository) Create(book *models.Book) error                                    { return nil }
func (m *MockBookRepository) GetByIDWithMetadata(id uint) (*models.Book, error)                { return m.GetByID(id) }
func (m *MockBookRepository) GetAll(options repository.BookQueryOptions) ([]*models.Book, int64, error) { return []*models.Book{}, 0, nil }
func (m *MockBookRepository) Update(book *models.Book) error                                    { return nil }
func (m *MockBookRepository) Delete(id uint) error                                            { return nil }
func (m *MockBookRepository) SoftDelete(id uint) error                                        { return nil }
func (m *MockBookRepository) Search(query string, limit, offset int) ([]*models.Book, int64, error) { return []*models.Book{}, 0, nil }
func (m *MockBookRepository) GetByGenre(genre string, limit, offset int) ([]*models.Book, error) { return []*models.Book{}, nil }
func (m *MockBookRepository) GetByAuthor(author string, limit, offset int) ([]*models.Book, error) { return []*models.Book{}, nil }
func (m *MockBookRepository) GetAllBooks() ([]*models.Book, error)                            { return []*models.Book{}, nil }
func (m *MockBookRepository) GetBookStats(bookID uint) (*repository.BookStats, error)        { return &repository.BookStats{}, nil }
func (m *MockBookRepository) CreateMetadata(metadata *models.BookMetadata) error             { return nil }
func (m *MockBookRepository) UpdateMetadata(metadata *models.BookMetadata) error             { return nil }
func (m *MockBookRepository) UpsertMetadata(metadata *models.BookMetadata) error             { return nil }
func (m *MockBookRepository) GetMetadataByBookID(bookID uint) ([]*models.BookMetadata, error) { return []*models.BookMetadata{}, nil }
func (m *MockBookRepository) GetBooksByRatingRange(minRating, maxRating int, limit, offset int) ([]*models.Book, error) { return []*models.Book{}, nil }
func (m *MockBookRepository) Count() (int64, error)                                          { return 0, nil }

func setupTestHandler() *Handler {
	comparisonRepo := NewMockComparisonRepository()
	rankingRepo := NewMockRankingRepository()
	bookRepo := &MockBookRepository{}
	eloService := service.NewELOService()

	comparisonService := service.NewComparisonService(comparisonRepo, rankingRepo, bookRepo, eloService)
	return NewHandler(comparisonService)
}

func createTestContext(userID uint) context.Context {
	claims := &auth.Claims{
		UserID:   userID,
		Username: "testuser",
	}
	return context.WithValue(context.Background(), "user", claims)
}

func TestSubmitComparison(t *testing.T) {
	handler := setupTestHandler()

	// Test successful submission with winner_id format (frontend format)
	frontendReq := FrontendComparisonRequest{
		BookAID:  1,
		BookBID:  2,
		WinnerID: 1, // Book A wins
	}

	reqBody, _ := json.Marshal(frontendReq)
	req := httptest.NewRequest(http.MethodPost, "/api/comparisons", bytes.NewBuffer(reqBody))
	req = req.WithContext(createTestContext(1))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.SubmitComparison(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response service.ComparisonResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Comparison == nil {
		t.Error("Expected comparison to be present in response")
	}

	if response.Comparison.Preference != models.PreferenceBookA {
		t.Errorf("Expected preference %s, got %s", models.PreferenceBookA, response.Comparison.Preference)
	}
}

func TestSubmitComparisonInvalidWinner(t *testing.T) {
	handler := setupTestHandler()

	// Test with invalid winner_id (not matching either book)
	frontendReq := FrontendComparisonRequest{
		BookAID:  1,
		BookBID:  2,
		WinnerID: 3, // Invalid - doesn't match either book
	}

	reqBody, _ := json.Marshal(frontendReq)
	req := httptest.NewRequest(http.MethodPost, "/api/comparisons", bytes.NewBuffer(reqBody))
	req = req.WithContext(createTestContext(1))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.SubmitComparison(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetComparisonHistory(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/comparisons", nil)
	req = req.WithContext(createTestContext(1))

	rr := httptest.NewRecorder()
	handler.GetComparisonHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := response["comparisons"]; !exists {
		t.Error("Expected 'comparisons' field in response")
	}

	if _, exists := response["count"]; !exists {
		t.Error("Expected 'count' field in response")
	}
}

func TestGetComparisonHistoryWithLimit(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/comparisons?limit=5", nil)
	req = req.WithContext(createTestContext(1))

	rr := httptest.NewRecorder()
	handler.GetComparisonHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := response["comparisons"]; !exists {
		t.Error("Expected 'comparisons' field in response")
	}
}

func TestGetPendingComparisons(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/comparisons/pending", nil)
	req = req.WithContext(createTestContext(1))

	rr := httptest.NewRecorder()
	handler.GetPendingComparisons(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := response["pending_comparisons"]; !exists {
		t.Error("Expected 'pending_comparisons' field in response")
	}

	if _, exists := response["count"]; !exists {
		t.Error("Expected 'count' field in response")
	}
}

func TestUnauthorizedRequest(t *testing.T) {
	handler := setupTestHandler()

	// Test without auth context
	req := httptest.NewRequest(http.MethodGet, "/api/comparisons", nil)
	// No auth context set

	rr := httptest.NewRecorder()
	handler.GetComparisonHistory(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}