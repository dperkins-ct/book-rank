package service

import (
	"fmt"
	"math"
)

// BookRankRatingService handles the proprietary BookRank rating system
// This wraps the ELO system but presents ratings on a 0-10 scale
type BookRankRatingService struct {
	eloService *ELOService
	minELO     int // Minimum ELO we expect to see (for scaling)
	maxELO     int // Maximum ELO we expect to see (for scaling)
}

// NewBookRankRatingService creates a new BookRank rating service
func NewBookRankRatingService(eloService *ELOService) *BookRankRatingService {
	return &BookRankRatingService{
		eloService: eloService,
		minELO:     800,  // Conservative minimum for active users
		maxELO:     2200, // Conservative maximum for active users
	}
}

// ConvertELOToBookRank converts an ELO rating to a BookRank score (0.0-10.0)
func (s *BookRankRatingService) ConvertELOToBookRank(eloRating int) float64 {
	// Handle edge cases
	if eloRating <= s.minELO {
		return 0.0
	}
	if eloRating >= s.maxELO {
		return 10.0
	}

	// Linear scaling from minELO-maxELO to 0-10
	ratio := float64(eloRating-s.minELO) / float64(s.maxELO-s.minELO)
	bookRankScore := ratio * 10.0

	// Round to 1 decimal place
	return math.Round(bookRankScore*10) / 10
}

// ConvertELOToBookRankInt converts ELO to BookRank as integer (0-100 for more precision in display)
func (s *BookRankRatingService) ConvertELOToBookRankInt(eloRating int) int {
	bookRankFloat := s.ConvertELOToBookRank(eloRating)
	return int(math.Round(bookRankFloat * 10)) // 0-100 scale for display precision
}

// GetRatingLabel returns a human-readable label for a BookRank score
func (s *BookRankRatingService) GetRatingLabel(bookRankScore float64) string {
	switch {
	case bookRankScore >= 9.0:
		return "Masterpiece"
	case bookRankScore >= 8.0:
		return "Excellent"
	case bookRankScore >= 7.0:
		return "Great"
	case bookRankScore >= 6.0:
		return "Good"
	case bookRankScore >= 5.0:
		return "Average"
	case bookRankScore >= 4.0:
		return "Below Average"
	case bookRankScore >= 3.0:
		return "Poor"
	case bookRankScore >= 2.0:
		return "Bad"
	case bookRankScore >= 1.0:
		return "Awful"
	default:
		return "Unrated"
	}
}

// GetRatingEmoji returns an emoji for a BookRank score
func (s *BookRankRatingService) GetRatingEmoji(bookRankScore float64) string {
	switch {
	case bookRankScore >= 9.0:
		return "🏆"
	case bookRankScore >= 8.0:
		return "⭐"
	case bookRankScore >= 7.0:
		return "🌟"
	case bookRankScore >= 6.0:
		return "👍"
	case bookRankScore >= 5.0:
		return "👌"
	case bookRankScore >= 4.0:
		return "😐"
	case bookRankScore >= 3.0:
		return "👎"
	case bookRankScore >= 2.0:
		return "😞"
	case bookRankScore >= 1.0:
		return "💩"
	default:
		return "❓"
	}
}

// AdjustELORangeBasedOnUserData dynamically adjusts the ELO range based on actual user data
// This should be called periodically or when displaying ratings
func (s *BookRankRatingService) AdjustELORangeBasedOnUserData(userRatings []int) {
	if len(userRatings) == 0 {
		return
	}

	min := userRatings[0]
	max := userRatings[0]

	for _, rating := range userRatings {
		if rating < min {
			min = rating
		}
		if rating > max {
			max = rating
		}
	}

	// Add some padding to ensure we use the full 0-10 range
	padding := int(float64(max-min) * 0.1)
	if padding < 50 { // Minimum padding
		padding = 50
	}

	s.minELO = min - padding
	s.maxELO = max + padding

	// Ensure reasonable bounds
	if s.minELO < 500 {
		s.minELO = 500
	}
	if s.maxELO > 2500 {
		s.maxELO = 2500
	}
}

// FormatBookRankScore formats a BookRank score for display
func (s *BookRankRatingService) FormatBookRankScore(bookRankScore float64) string {
	if bookRankScore < 0.05 {
		return "0.0"
	}
	if bookRankScore >= 9.95 {
		return "10.0"
	}
	return fmt.Sprintf("%.1f", bookRankScore)
}

// GetStartingBookRank returns the starting BookRank equivalent
func (s *BookRankRatingService) GetStartingBookRank() float64 {
	startingELO := s.eloService.GetStartingRating()
	return s.ConvertELOToBookRank(startingELO)
}