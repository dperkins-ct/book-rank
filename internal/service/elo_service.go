package service

import (
	"bookrank/internal/models"
	"errors"
	"math"
)

// ELOService handles ELO rating calculations
type ELOService struct {
	kFactor     float64
	startingELO int
}

// NewELOService creates a new ELO service
func NewELOService() *ELOService {
	return &ELOService{
		kFactor:     32.0, // Moderate sensitivity to rating changes
		startingELO: 1500, // Standard starting ELO rating
	}
}

// ComparisonResult represents the outcome of a comparison
type ComparisonResult struct {
	BookAScore float64 // 1.0 if Book A wins, 0.5 for tie, 0.0 if Book B wins
	BookBScore float64 // 1.0 if Book B wins, 0.5 for tie, 0.0 if Book A wins
}

// ELOUpdate represents the rating changes for both books
type ELOUpdate struct {
	BookANewRating int
	BookBNewRating int
	BookAChange    int
	BookBChange    int
}

// CalculateExpectedScore calculates the expected score for a player with given rating
// against an opponent with the given rating using the ELO formula
func (s *ELOService) CalculateExpectedScore(playerRating, opponentRating int) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, float64(opponentRating-playerRating)/400.0))
}

// CalculateNewRating calculates the new ELO rating after a game
func (s *ELOService) CalculateNewRating(currentRating int, expectedScore, actualScore float64) int {
	change := s.kFactor * (actualScore - expectedScore)
	newRating := float64(currentRating) + change

	// Ensure rating doesn't go below 0
	if newRating < 0 {
		newRating = 0
	}

	// Cap at reasonable maximum (3000)
	if newRating > 3000 {
		newRating = 3000
	}

	return int(math.Round(newRating))
}

// ConvertPreferenceToScores converts comparison preference to numeric scores
func (s *ELOService) ConvertPreferenceToScores(preference models.ComparisonPreference) (ComparisonResult, error) {
	switch preference {
	case models.PreferenceBookA:
		return ComparisonResult{BookAScore: 1.0, BookBScore: 0.0}, nil
	case models.PreferenceBookB:
		return ComparisonResult{BookAScore: 0.0, BookBScore: 1.0}, nil
	case models.PreferenceTie:
		return ComparisonResult{BookAScore: 0.5, BookBScore: 0.5}, nil
	default:
		return ComparisonResult{}, errors.New("invalid comparison preference")
	}
}

// CalculateELOUpdate calculates the new ratings for both books after a comparison
func (s *ELOService) CalculateELOUpdate(bookARating, bookBRating int, preference models.ComparisonPreference) (*ELOUpdate, error) {
	// Convert preference to scores
	result, err := s.ConvertPreferenceToScores(preference)
	if err != nil {
		return nil, err
	}

	// Calculate expected scores
	expectedScoreA := s.CalculateExpectedScore(bookARating, bookBRating)
	expectedScoreB := s.CalculateExpectedScore(bookBRating, bookARating)

	// Calculate new ratings
	newRatingA := s.CalculateNewRating(bookARating, expectedScoreA, result.BookAScore)
	newRatingB := s.CalculateNewRating(bookBRating, expectedScoreB, result.BookBScore)

	return &ELOUpdate{
		BookANewRating: newRatingA,
		BookBNewRating: newRatingB,
		BookAChange:    newRatingA - bookARating,
		BookBChange:    newRatingB - bookBRating,
	}, nil
}

// GetStartingRating returns the default starting ELO rating
func (s *ELOService) GetStartingRating() int {
	return s.startingELO
}

// GetKFactor returns the current K-factor
func (s *ELOService) GetKFactor() float64 {
	return s.kFactor
}

// SetKFactor allows changing the K-factor (useful for testing or different user types)
func (s *ELOService) SetKFactor(k float64) {
	if k > 0 {
		s.kFactor = k
	}
}

// CalculateRatingDifferenceProbability calculates the probability that the higher-rated
// book will win based on rating difference
func (s *ELOService) CalculateRatingDifferenceProbability(higherRating, lowerRating int) float64 {
	if higherRating < lowerRating {
		higherRating, lowerRating = lowerRating, higherRating
	}
	return s.CalculateExpectedScore(higherRating, lowerRating)
}

// IsSignificantRatingChange determines if a rating change is significant
// This can be used for notifications or highlighting important changes
func (s *ELOService) IsSignificantRatingChange(change int) bool {
	return math.Abs(float64(change)) >= 25 // Threshold for "significant" change
}

// EstimateRequiredComparisons estimates how many comparisons are needed
// to reach a stable rating for a book (rough heuristic)
func (s *ELOService) EstimateRequiredComparisons(currentComparisons int) int {
	// Heuristic: Most ratings stabilize after 10-15 comparisons
	stabilizationPoint := 15
	if currentComparisons >= stabilizationPoint {
		return 0
	}
	return stabilizationPoint - currentComparisons
}