package service

import (
	"bookrank/internal/models"
	"math"
	"sort"
	"strings"
)

// SimilarityCalculator handles user and book similarity calculations
type SimilarityCalculator struct{}

// NewSimilarityCalculator creates a new SimilarityCalculator
func NewSimilarityCalculator() *SimilarityCalculator {
	return &SimilarityCalculator{}
}

// UserSimilarity represents similarity between users
type UserSimilarity struct {
	UserID     uint    `json:"user_id"`
	Similarity float64 `json:"similarity"`
	SharedBooks int    `json:"shared_books"`
}

// BookSimilarity represents similarity between books
type BookSimilarity struct {
	BookID     uint    `json:"book_id"`
	Similarity float64 `json:"similarity"`
	Reasons    []string `json:"reasons"`
}

// CalculateUserSimilarity calculates similarity between users based on their rankings
// Uses Pearson correlation coefficient
func (sc *SimilarityCalculator) CalculateUserSimilarity(user1Rankings, user2Rankings []*models.Ranking, minSharedBooks int) (*UserSimilarity, error) {
	// Create maps for quick lookup
	user1Scores := make(map[uint]int)
	user2Scores := make(map[uint]int)

	for _, ranking := range user1Rankings {
		user1Scores[ranking.BookID] = ranking.Score
	}

	for _, ranking := range user2Rankings {
		user2Scores[ranking.BookID] = ranking.Score
	}

	// Find shared books
	var sharedBooks []uint
	for bookID := range user1Scores {
		if _, exists := user2Scores[bookID]; exists {
			sharedBooks = append(sharedBooks, bookID)
		}
	}

	// Need minimum shared books for meaningful similarity
	if len(sharedBooks) < minSharedBooks {
		return nil, nil
	}

	// Calculate Pearson correlation coefficient
	var sum1, sum2, sum1Sq, sum2Sq, sumProducts float64
	n := float64(len(sharedBooks))

	for _, bookID := range sharedBooks {
		score1 := float64(user1Scores[bookID])
		score2 := float64(user2Scores[bookID])

		sum1 += score1
		sum2 += score2
		sum1Sq += score1 * score1
		sum2Sq += score2 * score2
		sumProducts += score1 * score2
	}

	// Calculate correlation coefficient
	numerator := sumProducts - (sum1*sum2)/n
	denominator := math.Sqrt((sum1Sq - (sum1*sum1)/n) * (sum2Sq - (sum2*sum2)/n))

	if denominator == 0 {
		return nil, nil // No variation in ratings
	}

	correlation := numerator / denominator

	var userID uint
	if len(user2Rankings) > 0 {
		userID = user2Rankings[0].UserID
	}

	return &UserSimilarity{
		UserID:      userID,
		Similarity:  correlation,
		SharedBooks: len(sharedBooks),
	}, nil
}

// CalculateBookSimilarity calculates similarity between books based on various factors
func (sc *SimilarityCalculator) CalculateBookSimilarity(book1, book2 *models.Book) *BookSimilarity {
	var similarity float64
	var reasons []string

	// Genre similarity (40% weight)
	if book1.Genre == book2.Genre {
		similarity += 0.4
		reasons = append(reasons, "Same genre: "+book1.Genre)
	}

	// Author similarity (30% weight)
	if book1.Author == book2.Author {
		similarity += 0.3
		reasons = append(reasons, "Same author: "+book1.Author)
	}

	// Publication date proximity (20% weight)
	if book1.PublicationDate != nil && book2.PublicationDate != nil {
		yearDiff := math.Abs(float64(book1.PublicationDate.Year() - book2.PublicationDate.Year()))
		if yearDiff <= 5 {
			dateSimilarity := 0.2 * (1 - yearDiff/20) // Max 20 years for any similarity
			if dateSimilarity > 0 {
				similarity += dateSimilarity
				reasons = append(reasons, "Similar publication period")
			}
		}
	}

	// Title similarity (10% weight) - simple word overlap
	titleSimilarity := sc.calculateTitleSimilarity(book1.Title, book2.Title)
	if titleSimilarity > 0.3 {
		similarity += 0.1 * titleSimilarity
		reasons = append(reasons, "Similar titles")
	}

	return &BookSimilarity{
		BookID:     book2.ID,
		Similarity: similarity,
		Reasons:    reasons,
	}
}

// calculateTitleSimilarity calculates similarity between two titles based on word overlap
func (sc *SimilarityCalculator) calculateTitleSimilarity(title1, title2 string) float64 {
	words1 := sc.normalizeAndSplit(title1)
	words2 := sc.normalizeAndSplit(title2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// Count common words
	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	commonWords := 0
	for _, word := range words2 {
		if wordSet1[word] {
			commonWords++
		}
	}

	// Jaccard similarity
	totalWords := len(words1) + len(words2) - commonWords
	if totalWords == 0 {
		return 0
	}

	return float64(commonWords) / float64(totalWords)
}

// normalizeAndSplit normalizes and splits text into words
func (sc *SimilarityCalculator) normalizeAndSplit(text string) []string {
	// Simple normalization - convert to lowercase and split by spaces
	// In a production system, you might want more sophisticated text processing
	words := []string{}

	// Convert to lowercase and remove common words
	text = strings.ToLower(text)
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true,
	}

	for _, word := range strings.Fields(text) {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) > 2 && !commonWords[word] {
			words = append(words, word)
		}
	}

	return words
}

// FindTopSimilarUsers finds the most similar users to a given user
func (sc *SimilarityCalculator) FindTopSimilarUsers(targetUserRankings []*models.Ranking, allUserRankings map[uint][]*models.Ranking, limit int, minSharedBooks int) []*UserSimilarity {
	var similarities []*UserSimilarity

	for userID, rankings := range allUserRankings {
		// Skip self
		if len(targetUserRankings) > 0 && userID == targetUserRankings[0].UserID {
			continue
		}

		similarity, err := sc.CalculateUserSimilarity(targetUserRankings, rankings, minSharedBooks)
		if err == nil && similarity != nil && similarity.Similarity > 0 {
			similarities = append(similarities, similarity)
		}
	}

	// Sort by similarity score descending
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	// Return top N similar users
	if limit > 0 && len(similarities) > limit {
		return similarities[:limit]
	}

	return similarities
}

// FindSimilarBooks finds books similar to a given book
func (sc *SimilarityCalculator) FindSimilarBooks(targetBook *models.Book, allBooks []*models.Book, limit int, minSimilarity float64) []*BookSimilarity {
	var similarities []*BookSimilarity

	for _, book := range allBooks {
		if book.ID == targetBook.ID {
			continue // Skip self
		}

		similarity := sc.CalculateBookSimilarity(targetBook, book)
		if similarity.Similarity >= minSimilarity {
			similarities = append(similarities, similarity)
		}
	}

	// Sort by similarity score descending
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	// Return top N similar books
	if limit > 0 && len(similarities) > limit {
		return similarities[:limit]
	}

	return similarities
}