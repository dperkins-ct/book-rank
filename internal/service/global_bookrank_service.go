package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"fmt"
	"math"
)

// GlobalBookRankService calculates book ratings across all users
type GlobalBookRankService struct {
	comparisonRepo repository.ComparisonRepository
	rankingRepo    repository.RankingRepository
	bookRepo       repository.BookRepository
}

// NewGlobalBookRankService creates a new global BookRank service
func NewGlobalBookRankService(
	comparisonRepo repository.ComparisonRepository,
	rankingRepo repository.RankingRepository,
	bookRepo repository.BookRepository,
) *GlobalBookRankService {
	return &GlobalBookRankService{
		comparisonRepo: comparisonRepo,
		rankingRepo:    rankingRepo,
		bookRepo:       bookRepo,
	}
}

// GlobalBookStats represents global statistics for a book
type GlobalBookStats struct {
	BookID              uint    `json:"book_id"`
	GlobalBookRankScore float64 `json:"global_bookrank_score"`  // 0-10 score
	TotalComparisons    int     `json:"total_comparisons"`     // How many times compared
	TotalUsers          int     `json:"total_users"`           // How many users have rated it
	WinRate             float64 `json:"win_rate"`              // Percentage of comparisons won
	ConfidenceLevel     string  `json:"confidence_level"`      // Low/Medium/High based on sample size
	AverageELOAcrossUsers float64 `json:"average_elo_across_users"` // Internal calculation
}

// CalculateGlobalBookRank calculates the global BookRank score for a book
func (s *GlobalBookRankService) CalculateGlobalBookRank(bookID uint) (*GlobalBookStats, error) {
	// Get all rankings for this book across all users
	allRankings, err := s.rankingRepo.GetAllByBookID(bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings for book %d: %w", bookID, err)
	}

	if len(allRankings) == 0 {
		return &GlobalBookStats{
			BookID:              bookID,
			GlobalBookRankScore: 5.0, // Default neutral score
			ConfidenceLevel:     "None",
		}, nil
	}

	// Calculate average ELO across all users who have rated this book
	totalELO := 0
	userCount := len(allRankings)

	for _, ranking := range allRankings {
		totalELO += ranking.Score
	}

	averageELO := float64(totalELO) / float64(userCount)

	// Get comparison statistics
	comparisons, err := s.getBookComparisonStats(bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison stats: %w", err)
	}

	// Convert average ELO to 0-10 BookRank scale
	// Use a more aggressive scaling for global scores to utilize full 0-10 range
	globalScore := s.convertAverageELOToGlobalBookRank(averageELO, userCount)

	// Determine confidence level based on sample size
	confidence := s.calculateConfidenceLevel(userCount, comparisons.TotalComparisons)

	return &GlobalBookStats{
		BookID:                bookID,
		GlobalBookRankScore:   globalScore,
		TotalComparisons:      comparisons.TotalComparisons,
		TotalUsers:            userCount,
		WinRate:               comparisons.WinRate,
		ConfidenceLevel:       confidence,
		AverageELOAcrossUsers: averageELO,
	}, nil
}

// BookComparisonStats holds comparison statistics for a book
type BookComparisonStats struct {
	TotalComparisons int
	Wins             int
	Losses           int
	Ties             int
	WinRate          float64
}

// getBookComparisonStats calculates win/loss statistics for a book
func (s *GlobalBookRankService) getBookComparisonStats(bookID uint) (*BookComparisonStats, error) {
	// Get all comparisons involving this book
	// This is a simplified version - you may want to add this method to the comparison repo
	stats := &BookComparisonStats{}

	// For now, we'll estimate based on the number of users who have rated the book
	// In a full implementation, you'd query all comparisons where this book was involved

	return stats, nil
}

// convertAverageELOToGlobalBookRank converts average ELO to 0-10 global BookRank
func (s *GlobalBookRankService) convertAverageELOToGlobalBookRank(averageELO float64, sampleSize int) float64 {
	// Use a more aggressive scaling for global scores
	// Assume global ELO range is wider: 1000-2000
	minGlobalELO := 1000.0
	maxGlobalELO := 2000.0

	// Apply sample size confidence adjustment
	confidenceMultiplier := s.calculateSampleSizeMultiplier(sampleSize)

	// Normalize the ELO
	if averageELO <= minGlobalELO {
		return 0.0 * confidenceMultiplier
	}
	if averageELO >= maxGlobalELO {
		return 10.0 * confidenceMultiplier
	}

	// Linear scaling with confidence adjustment
	ratio := (averageELO - minGlobalELO) / (maxGlobalELO - minGlobalELO)
	baseScore := ratio * 10.0

	// Apply confidence multiplier (lower confidence pulls toward 5.0)
	finalScore := (baseScore * confidenceMultiplier) + (5.0 * (1.0 - confidenceMultiplier))

	// Ensure bounds
	if finalScore < 0 {
		return 0.0
	}
	if finalScore > 10 {
		return 10.0
	}

	return math.Round(finalScore*10) / 10 // Round to 1 decimal place
}

// calculateSampleSizeMultiplier returns a multiplier based on sample size confidence
func (s *GlobalBookRankService) calculateSampleSizeMultiplier(sampleSize int) float64 {
	// Returns 0.0 to 1.0 based on how confident we are in the sample
	if sampleSize >= 50 {
		return 1.0 // High confidence
	}
	if sampleSize >= 20 {
		return 0.8 // Good confidence
	}
	if sampleSize >= 10 {
		return 0.6 // Medium confidence
	}
	if sampleSize >= 5 {
		return 0.4 // Low confidence
	}
	return 0.2 // Very low confidence
}

// calculateConfidenceLevel returns a human-readable confidence level
func (s *GlobalBookRankService) calculateConfidenceLevel(userCount, comparisonCount int) string {
	totalSample := userCount + (comparisonCount / 10) // Weight comparisons less than users

	if totalSample >= 50 {
		return "High"
	}
	if totalSample >= 20 {
		return "Medium"
	}
	if totalSample >= 5 {
		return "Low"
	}
	return "Very Low"
}

// CalculatePersonalRankingPosition calculates where a user ranks their books (0-10 personal scale)
func (s *GlobalBookRankService) CalculatePersonalRankingPosition(userID uint, bookID uint) (float64, error) {
	// Get all of the user's rankings
	userRankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}

	if len(userRankings) <= 1 {
		return 5.0, nil // Default middle score if only one book or none
	}

	// Find the target book's ranking
	var targetRanking *models.Ranking
	for _, ranking := range userRankings {
		if ranking.BookID == bookID {
			targetRanking = ranking
			break
		}
	}

	if targetRanking == nil {
		return 5.0, nil // Not found, return neutral
	}

	// Calculate position within user's personal rankings (0-based)
	position := 0
	for _, ranking := range userRankings {
		if ranking.Score > targetRanking.Score {
			position++
		}
	}

	// Convert position to 0-10 scale
	totalBooks := len(userRankings)
	personalScore := 10.0 - (float64(position)/float64(totalBooks-1))*10.0

	// Ensure bounds
	if personalScore < 0 {
		personalScore = 0.0
	}
	if personalScore > 10 {
		personalScore = 10.0
	}

	return math.Round(personalScore*10) / 10, nil
}

// GetGlobalBookRankLabel returns a label for global BookRank scores
func (s *GlobalBookRankService) GetGlobalBookRankLabel(score float64) string {
	switch {
	case score >= 9.0:
		return "Universally Acclaimed"
	case score >= 8.5:
		return "Highly Regarded"
	case score >= 7.5:
		return "Well Liked"
	case score >= 6.5:
		return "Generally Positive"
	case score >= 5.5:
		return "Mixed Reception"
	case score >= 4.5:
		return "Below Average"
	case score >= 3.5:
		return "Poorly Regarded"
	case score >= 2.0:
		return "Widely Disliked"
	default:
		return "Universally Panned"
	}
}