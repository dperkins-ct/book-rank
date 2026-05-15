package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// ComparisonService handles book comparison logic and ELO updates
type ComparisonService struct {
	comparisonRepo repository.ComparisonRepository
	rankingRepo    repository.RankingRepository
	bookRepo       repository.BookRepository
	eloService     *ELOService
}

// NewComparisonService creates a new comparison service
func NewComparisonService(
	comparisonRepo repository.ComparisonRepository,
	rankingRepo repository.RankingRepository,
	bookRepo repository.BookRepository,
	eloService *ELOService,
) *ComparisonService {
	return &ComparisonService{
		comparisonRepo: comparisonRepo,
		rankingRepo:    rankingRepo,
		bookRepo:       bookRepo,
		eloService:     eloService,
	}
}

// ComparisonRequest represents a user's comparison submission
type ComparisonRequest struct {
	UserID     uint                         `json:"user_id" validate:"required"`
	BookAID    uint                         `json:"book_a_id" validate:"required"`
	BookBID    uint                         `json:"book_b_id" validate:"required"`
	Preference models.ComparisonPreference `json:"preference" validate:"required,oneof=book_a book_b tie"`
}

// ComparisonResponse represents the response after processing a comparison
type ComparisonResponse struct {
	Comparison     *models.Comparison `json:"comparison"`
	ELOUpdate      *ELOUpdate         `json:"elo_update"`
	BookARanking   *models.Ranking    `json:"book_a_ranking"`
	BookBRanking   *models.Ranking    `json:"book_b_ranking"`
	NextComparison *repository.PendingComparison `json:"next_comparison,omitempty"`
}

// SubmitComparison processes a pairwise comparison and updates ELO ratings
func (s *ComparisonService) SubmitComparison(req *ComparisonRequest) (*ComparisonResponse, error) {
	// Validate request
	if err := s.validateComparisonRequest(req); err != nil {
		return nil, err
	}

	// Check if comparison already exists
	existing, err := s.comparisonRepo.GetByUserAndBooks(req.UserID, req.BookAID, req.BookBID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing comparison: %w", err)
	}
	if existing != nil {
		return nil, errors.New("comparison already exists for these books")
	}

	// Get current rankings for both books
	rankingA, err := s.getOrCreateRanking(req.UserID, req.BookAID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking for book A: %w", err)
	}

	rankingB, err := s.getOrCreateRanking(req.UserID, req.BookBID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking for book B: %w", err)
	}

	// Calculate ELO update
	eloUpdate, err := s.eloService.CalculateELOUpdate(rankingA.Score, rankingB.Score, req.Preference)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ELO update: %w", err)
	}

	// Create comparison record
	comparison := &models.Comparison{
		UserID:     req.UserID,
		BookAID:    req.BookAID,
		BookBID:    req.BookBID,
		Preference: req.Preference,
	}

	if err := s.comparisonRepo.Create(comparison); err != nil {
		return nil, fmt.Errorf("failed to create comparison: %w", err)
	}

	// Update rankings with new ELO ratings
	rankingA.Score = eloUpdate.BookANewRating
	rankingB.Score = eloUpdate.BookBNewRating

	if err := s.rankingRepo.Update(rankingA); err != nil {
		return nil, fmt.Errorf("failed to update ranking A: %w", err)
	}

	if err := s.rankingRepo.Update(rankingB); err != nil {
		return nil, fmt.Errorf("failed to update ranking B: %w", err)
	}

	// Get next comparison suggestion
	nextComparisons, _ := s.comparisonRepo.GetPendingComparisons(req.UserID, 1)
	var nextComparison *repository.PendingComparison
	if len(nextComparisons) > 0 {
		nextComparison = &nextComparisons[0]
	}

	return &ComparisonResponse{
		Comparison:     comparison,
		ELOUpdate:      eloUpdate,
		BookARanking:   rankingA,
		BookBRanking:   rankingB,
		NextComparison: nextComparison,
	}, nil
}

// GetPendingComparisons returns books that need comparison for a user
func (s *ComparisonService) GetPendingComparisons(userID uint, limit int) ([]repository.PendingComparison, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	return s.comparisonRepo.GetPendingComparisons(userID, limit)
}

// GetComparisonHistory returns the comparison history for a user
func (s *ComparisonService) GetComparisonHistory(userID uint) ([]*models.Comparison, error) {
	return s.comparisonRepo.GetByUserID(userID)
}

// GetBookComparisons returns all comparisons involving a specific book for a user
func (s *ComparisonService) GetBookComparisons(userID, bookID uint) ([]*models.Comparison, error) {
	return s.comparisonRepo.GetUserComparisonsForBook(userID, bookID)
}

// RecalculateUserRatings recalculates all ratings for a user based on their comparison history
func (s *ComparisonService) RecalculateUserRatings(userID uint) error {
	// Get all user's rankings
	rankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user rankings: %w", err)
	}

	// Reset all ratings to starting value
	startingRating := s.eloService.GetStartingRating()
	for _, ranking := range rankings {
		ranking.Score = startingRating
		if err := s.rankingRepo.Update(ranking); err != nil {
			return fmt.Errorf("failed to reset ranking: %w", err)
		}
	}

	// Get all comparisons ordered by creation time (oldest first for chronological recalculation)
	comparisons, err := s.comparisonRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user comparisons: %w", err)
	}

	// Reverse order to process oldest first
	for i := len(comparisons) - 1; i >= 0; i-- {
		comparison := comparisons[i]

		// Get current rankings
		rankingA, err := s.rankingRepo.GetByUserAndBook(userID, comparison.BookAID)
		if err != nil {
			continue // Skip if ranking doesn't exist
		}

		rankingB, err := s.rankingRepo.GetByUserAndBook(userID, comparison.BookBID)
		if err != nil {
			continue // Skip if ranking doesn't exist
		}

		// Calculate and apply ELO update
		eloUpdate, err := s.eloService.CalculateELOUpdate(rankingA.Score, rankingB.Score, comparison.Preference)
		if err != nil {
			continue // Skip invalid comparisons
		}

		// Update rankings
		rankingA.Score = eloUpdate.BookANewRating
		rankingB.Score = eloUpdate.BookBNewRating

		s.rankingRepo.Update(rankingA)
		s.rankingRepo.Update(rankingB)
	}

	return nil
}

// GetUserRankings returns all rankings for a user, ordered by score
func (s *ComparisonService) GetUserRankings(userID uint) ([]*models.Ranking, error) {
	return s.rankingRepo.GetByUserID(userID)
}

// IsOnboardingComplete checks if a user has completed the onboarding process
func (s *ComparisonService) IsOnboardingComplete(userID uint) (bool, error) {
	// Check if user has at least 10 books
	rankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}

	if len(rankings) < 10 {
		return false, nil
	}

	// Check if user has made at least some comparisons
	comparisons, err := s.comparisonRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}

	// Require at least 5 comparisons to consider onboarding complete
	return len(comparisons) >= 5, nil
}

// validateComparisonRequest validates the comparison request
func (s *ComparisonService) validateComparisonRequest(req *ComparisonRequest) error {
	if req.UserID == 0 {
		return errors.New("user ID is required")
	}
	if req.BookAID == 0 {
		return errors.New("book A ID is required")
	}
	if req.BookBID == 0 {
		return errors.New("book B ID is required")
	}
	if req.BookAID == req.BookBID {
		return errors.New("cannot compare a book with itself")
	}
	if !isValidPreference(req.Preference) {
		return errors.New("invalid preference value")
	}
	return nil
}

// getOrCreateRanking gets an existing ranking or creates a new one with starting ELO
func (s *ComparisonService) getOrCreateRanking(userID, bookID uint) (*models.Ranking, error) {
	ranking, err := s.rankingRepo.GetByUserAndBook(userID, bookID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if ranking == nil {
		// Create new ranking with starting ELO
		ranking = &models.Ranking{
			UserID: userID,
			BookID: bookID,
			Score:  s.eloService.GetStartingRating(),
		}
		if err := s.rankingRepo.Create(ranking); err != nil {
			return nil, err
		}
	}

	return ranking, nil
}

// GetRandomBookPair returns two random books that haven't been compared by the user
func (s *ComparisonService) GetRandomBookPair(userID uint) (*repository.PendingComparison, error) {
	pending, err := s.comparisonRepo.GetPendingComparisons(userID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending comparisons: %w", err)
	}

	if len(pending) == 0 {
		return nil, nil // Return nil without error when no pairs available
	}

	return &pending[0], nil
}

// GetRandomBookPairWithContext returns two random books that haven't been compared by the user with context support
func (s *ComparisonService) GetRandomBookPairWithContext(ctx context.Context, userID uint) (*repository.PendingComparison, error) {
	// Check if context is cancelled before proceeding
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	pending, err := s.comparisonRepo.GetPendingComparisons(userID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending comparisons: %w", err)
	}

	if len(pending) == 0 {
		return nil, nil // Return nil without error when no pairs available
	}

	return &pending[0], nil
}

// isValidPreference checks if the preference is valid
func isValidPreference(pref models.ComparisonPreference) bool {
	return pref == models.PreferenceBookA || pref == models.PreferenceBookB || pref == models.PreferenceTie
}