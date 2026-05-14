package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// RankingService handles business logic for ranking operations
type RankingService struct {
	rankingRepo repository.RankingRepository
	eloService  *ELOService
}

// NewRankingService creates a new RankingService
func NewRankingService(rankingRepo repository.RankingRepository, eloService *ELOService) *RankingService {
	return &RankingService{
		rankingRepo: rankingRepo,
		eloService:  eloService,
	}
}

// RankingStats represents statistical information about a user's rankings
type RankingStats struct {
	TotalBooks      int     `json:"total_books"`
	AverageRating   float64 `json:"average_rating"`
	HighestRating   int     `json:"highest_rating"`
	LowestRating    int     `json:"lowest_rating"`
	RatingSpread    int     `json:"rating_spread"`
	BooksAboveStart int     `json:"books_above_starting_rating"`
	BooksBelowStart int     `json:"books_below_starting_rating"`
}

// GetUserRankings retrieves all rankings for a specific user
func (s *RankingService) GetUserRankings(userID uint) ([]*models.Ranking, error) {
	return s.rankingRepo.GetByUserID(userID)
}

// GetTopRankedBooks returns the highest-ranked books for a user
func (s *RankingService) GetTopRankedBooks(userID uint, limit int) ([]*models.Ranking, error) {
	return s.rankingRepo.GetTopRanked(userID, limit)
}

// GetUserRankingStats calculates statistical information about a user's rankings
func (s *RankingService) GetUserRankingStats(userID uint) (*RankingStats, error) {
	rankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if len(rankings) == 0 {
		return &RankingStats{}, nil
	}

	startingRating := s.eloService.GetStartingRating()
	stats := &RankingStats{
		TotalBooks:    len(rankings),
		HighestRating: rankings[0].Score, // Already sorted by score DESC
		LowestRating:  rankings[len(rankings)-1].Score,
	}

	totalScore := 0
	for _, ranking := range rankings {
		totalScore += ranking.Score
		if ranking.Score > startingRating {
			stats.BooksAboveStart++
		} else if ranking.Score < startingRating {
			stats.BooksBelowStart++
		}
	}

	stats.AverageRating = float64(totalScore) / float64(len(rankings))
	stats.RatingSpread = stats.HighestRating - stats.LowestRating

	return stats, nil
}

// InitializeRanking creates an initial ranking for a book when a user first encounters it
func (s *RankingService) InitializeRanking(userID, bookID uint) error {
	// Check if ranking already exists
	existing, err := s.rankingRepo.GetByUserAndBook(userID, bookID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return errors.New("ranking already exists")
	}

	ranking := &models.Ranking{
		UserID: userID,
		BookID: bookID,
		Score:  s.eloService.GetStartingRating(),
	}
	return s.rankingRepo.Create(ranking)
}

// UpdateRanking updates an existing ranking
func (s *RankingService) UpdateRanking(userID, bookID uint, newScore int) error {
	ranking, err := s.rankingRepo.GetByUserAndBook(userID, bookID)
	if err != nil {
		return err
	}

	ranking.Score = newScore
	return s.rankingRepo.Update(ranking)
}

// GetRanking retrieves a specific ranking
func (s *RankingService) GetRanking(userID, bookID uint) (*models.Ranking, error) {
	return s.rankingRepo.GetByUserAndBook(userID, bookID)
}

// BulkInitializeRankings creates initial rankings for multiple books
func (s *RankingService) BulkInitializeRankings(userID uint, bookIDs []uint) error {
	startingRating := s.eloService.GetStartingRating()

	for _, bookID := range bookIDs {
		// Check if ranking already exists
		existing, err := s.rankingRepo.GetByUserAndBook(userID, bookID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing ranking for book %d: %w", bookID, err)
		}

		// Skip if ranking already exists
		if existing != nil {
			continue
		}

		ranking := &models.Ranking{
			UserID: userID,
			BookID: bookID,
			Score:  startingRating,
		}

		if err := s.rankingRepo.Create(ranking); err != nil {
			return fmt.Errorf("failed to create ranking for book %d: %w", bookID, err)
		}
	}

	return nil
}

// GetRankingPosition returns the position of a book in the user's rankings
func (s *RankingService) GetRankingPosition(userID, bookID uint) (int, error) {
	rankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}

	for i, ranking := range rankings {
		if ranking.BookID == bookID {
			return i + 1, nil // 1-indexed position
		}
	}

	return 0, errors.New("book not found in user's rankings")
}

// GetRankingComparison compares two books' rankings for a user
func (s *RankingService) GetRankingComparison(userID, bookAID, bookBID uint) (*ComparisonAnalysis, error) {
	rankingA, err := s.rankingRepo.GetByUserAndBook(userID, bookAID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking for book A: %w", err)
	}

	rankingB, err := s.rankingRepo.GetByUserAndBook(userID, bookBID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking for book B: %w", err)
	}

	analysis := &ComparisonAnalysis{
		BookA:          *rankingA,
		BookB:          *rankingB,
		RatingDiff:     rankingA.Score - rankingB.Score,
		ExpectedWinProbA: s.eloService.CalculateExpectedScore(rankingA.Score, rankingB.Score),
	}
	analysis.ExpectedWinProbB = 1.0 - analysis.ExpectedWinProbA

	return analysis, nil
}

// ComparisonAnalysis provides detailed analysis of two books' relative ratings
type ComparisonAnalysis struct {
	BookA            models.Ranking `json:"book_a"`
	BookB            models.Ranking `json:"book_b"`
	RatingDiff       int            `json:"rating_difference"`
	ExpectedWinProbA float64        `json:"expected_win_probability_a"`
	ExpectedWinProbB float64        `json:"expected_win_probability_b"`
}