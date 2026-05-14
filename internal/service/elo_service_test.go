package service

import (
	"bookrank/internal/models"
	"math"
	"testing"
)

func TestELOService_CalculateExpectedScore(t *testing.T) {
	eloService := NewELOService()

	tests := []struct {
		name           string
		playerRating   int
		opponentRating int
		expectedScore  float64
		tolerance      float64
	}{
		{
			name:           "Equal ratings",
			playerRating:   1500,
			opponentRating: 1500,
			expectedScore:  0.5,
			tolerance:      0.001,
		},
		{
			name:           "Player 100 points higher",
			playerRating:   1600,
			opponentRating: 1500,
			expectedScore:  0.64,
			tolerance:      0.01,
		},
		{
			name:           "Player 400 points higher",
			playerRating:   1900,
			opponentRating: 1500,
			expectedScore:  0.909,
			tolerance:      0.001,
		},
		{
			name:           "Player 200 points lower",
			playerRating:   1300,
			opponentRating: 1500,
			expectedScore:  0.24,
			tolerance:      0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eloService.CalculateExpectedScore(tt.playerRating, tt.opponentRating)
			if math.Abs(result-tt.expectedScore) > tt.tolerance {
				t.Errorf("Expected %f, got %f", tt.expectedScore, result)
			}
		})
	}
}

func TestELOService_CalculateNewRating(t *testing.T) {
	eloService := NewELOService()

	tests := []struct {
		name          string
		currentRating int
		expectedScore float64
		actualScore   float64
		expectedNew   int
	}{
		{
			name:          "Win when expected to lose",
			currentRating: 1400,
			expectedScore: 0.24,
			actualScore:   1.0,
			expectedNew:   1424,
		},
		{
			name:          "Lose when expected to win",
			currentRating: 1600,
			expectedScore: 0.76,
			actualScore:   0.0,
			expectedNew:   1576,
		},
		{
			name:          "Expected result - win",
			currentRating: 1500,
			expectedScore: 0.64,
			actualScore:   1.0,
			expectedNew:   1512,
		},
		{
			name:          "Draw when expected to win",
			currentRating: 1600,
			expectedScore: 0.64,
			actualScore:   0.5,
			expectedNew:   1596, // Adjusted for rounding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eloService.CalculateNewRating(tt.currentRating, tt.expectedScore, tt.actualScore)
			if result != tt.expectedNew {
				t.Errorf("Expected %d, got %d", tt.expectedNew, result)
			}
		})
	}
}

func TestELOService_ConvertPreferenceToScores(t *testing.T) {
	eloService := NewELOService()

	tests := []struct {
		name       string
		preference models.ComparisonPreference
		expectA    float64
		expectB    float64
		expectErr  bool
	}{
		{
			name:       "Book A preferred",
			preference: models.PreferenceBookA,
			expectA:    1.0,
			expectB:    0.0,
			expectErr:  false,
		},
		{
			name:       "Book B preferred",
			preference: models.PreferenceBookB,
			expectA:    0.0,
			expectB:    1.0,
			expectErr:  false,
		},
		{
			name:       "Tie",
			preference: models.PreferenceTie,
			expectA:    0.5,
			expectB:    0.5,
			expectErr:  false,
		},
		{
			name:       "Invalid preference",
			preference: models.ComparisonPreference("invalid"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eloService.ConvertPreferenceToScores(tt.preference)
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result.BookAScore != tt.expectA {
				t.Errorf("Expected BookA score %f, got %f", tt.expectA, result.BookAScore)
			}
			if result.BookBScore != tt.expectB {
				t.Errorf("Expected BookB score %f, got %f", tt.expectB, result.BookBScore)
			}
		})
	}
}

func TestELOService_CalculateELOUpdate(t *testing.T) {
	eloService := NewELOService()

	tests := []struct {
		name         string
		bookARating  int
		bookBRating  int
		preference   models.ComparisonPreference
		expectANew   int
		expectBNew   int
	}{
		{
			name:         "Equal ratings, A wins",
			bookARating:  1500,
			bookBRating:  1500,
			preference:   models.PreferenceBookA,
			expectANew:   1516,
			expectBNew:   1484,
		},
		{
			name:         "Equal ratings, tie",
			bookARating:  1500,
			bookBRating:  1500,
			preference:   models.PreferenceTie,
			expectANew:   1500,
			expectBNew:   1500,
		},
		{
			name:         "Upset victory - lower rated wins",
			bookARating:  1400,
			bookBRating:  1600,
			preference:   models.PreferenceBookA,
			expectANew:   1424,
			expectBNew:   1576,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update, err := eloService.CalculateELOUpdate(tt.bookARating, tt.bookBRating, tt.preference)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if update.BookANewRating != tt.expectANew {
				t.Errorf("Expected BookA new rating %d, got %d", tt.expectANew, update.BookANewRating)
			}
			if update.BookBNewRating != tt.expectBNew {
				t.Errorf("Expected BookB new rating %d, got %d", tt.expectBNew, update.BookBNewRating)
			}
		})
	}
}

func TestELOService_RatingBoundaries(t *testing.T) {
	eloService := NewELOService()

	// Test minimum rating boundary
	newRating := eloService.CalculateNewRating(100, 0.9, 0.0) // Very low rating loses badly
	if newRating < 0 {
		t.Errorf("Rating should not go below 0, got %d", newRating)
	}

	// Test maximum rating boundary
	newRating = eloService.CalculateNewRating(2950, 0.1, 1.0) // Very high rating wins unexpectedly
	if newRating > 3000 {
		t.Errorf("Rating should not go above 3000, got %d", newRating)
	}
}

func TestELOService_IsSignificantRatingChange(t *testing.T) {
	eloService := NewELOService()

	tests := []struct {
		change      int
		significant bool
	}{
		{24, false},
		{25, true},
		{-25, true},
		{30, true},
		{-30, true},
		{10, false},
		{0, false},
	}

	for _, tt := range tests {
		result := eloService.IsSignificantRatingChange(tt.change)
		if result != tt.significant {
			t.Errorf("Change %d: expected %v, got %v", tt.change, tt.significant, result)
		}
	}
}