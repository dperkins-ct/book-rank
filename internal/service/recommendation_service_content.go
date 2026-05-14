package service

import (
	"bookrank/internal/models"
	"context"
	"fmt"
)

// getContentBasedRecommendations implements content-based filtering
func (s *RecommendationService) getContentBasedRecommendations(userID uint, userRankings []*models.Ranking, limit int) ([]*RecommendationResult, error) {
	// Calculate user's genre preferences
	genreAvgs, err := s.rankingRepo.GetAverageRatingByGenre(userID)
	if err != nil {
		return nil, err
	}

	// Get unrated books
	unratedBooks, err := s.rankingRepo.GetUnratedBooksByUser(userID)
	if err != nil {
		return nil, err
	}

	bookScores := make(map[uint]float64)
	bookReasons := make(map[uint][]string)

	// Score books based on user's genre/author preferences
	for _, book := range unratedBooks {
		var score float64
		var reasons []string

		// Genre preference (70% weight)
		if genreAvg, exists := genreAvgs[book.Genre]; exists {
			genreScore := genreAvg / 3000.0 // Normalize to 0-1
			score += genreScore * 0.7
			reasons = append(reasons, fmt.Sprintf("You like %s books (avg rating: %.0f)", book.Genre, genreAvg))
		}

		// Author preference (30% weight)
		authorScore := s.calculateAuthorPreference(userID, book.Author, userRankings)
		if authorScore > 0 {
			score += authorScore * 0.3
			reasons = append(reasons, fmt.Sprintf("You've enjoyed books by %s", book.Author))
		}

		// Recency bonus (newer books get slight boost)
		if book.PublicationDate != nil {
			yearsSince := float64(2024 - book.PublicationDate.Year())
			if yearsSince <= 5 {
				recencyBonus := (5 - yearsSince) / 50 // Small bonus for recent books
				score += recencyBonus
				reasons = append(reasons, "Recent publication")
			}
		}

		if score > 0 {
			bookScores[book.ID] = score
			bookReasons[book.ID] = reasons
		}
	}

	return s.createRecommendationResults(bookScores, bookReasons, "content_based", limit)
}

// calculateAuthorPreference calculates user's preference for a specific author
func (s *RecommendationService) calculateAuthorPreference(userID uint, author string, userRankings []*models.Ranking) float64 {
	var totalScore float64
	var count int

	for _, ranking := range userRankings {
		if ranking.Book.Author == author {
			totalScore += float64(ranking.Score)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return (totalScore / float64(count)) / 3000.0 // Normalize to 0-1
}

// getPopularRecommendations returns popular books as fallback
func (s *RecommendationService) getPopularRecommendations(limit int) ([]*RecommendationResult, error) {
	// Get books with highest average ratings
	books, err := s.bookRepo.GetAllBooks()
	if err != nil {
		return nil, err
	}

	// This is a simplified implementation - in production you'd want
	// to query for books with highest average ratings directly
	bookScores := make(map[uint]float64)
	bookReasons := make(map[uint][]string)

	for i, book := range books {
		if i >= limit {
			break
		}
		// Simple scoring based on order (newest first in most repos)
		score := 1.0 - (float64(i) / float64(len(books)))
		bookScores[book.ID] = score
		bookReasons[book.ID] = []string{"Popular book", "Trending selection"}
	}

	return s.createRecommendationResults(bookScores, bookReasons, "popular", limit)
}

// GetSimilarBooks finds books similar to a given book
func (s *RecommendationService) GetSimilarBooks(bookID uint, limit int) ([]*RecommendationResult, error) {
	// Get the target book
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil {
		return nil, err
	}

	// Get all books
	allBooks, err := s.bookRepo.GetAllBooks()
	if err != nil {
		return nil, err
	}

	// Calculate similarities
	similarities := s.similarity.FindSimilarBooks(book, allBooks, limit, 0.1)

	var results []*RecommendationResult
	for _, sim := range similarities {
		for _, b := range allBooks {
			if b.ID == sim.BookID {
				results = append(results, &RecommendationResult{
					Book:     b,
					Score:    sim.Similarity,
					Reasons:  sim.Reasons,
					Strategy: "similarity",
				})
				break
			}
		}
	}

	return results, nil
}

// Helper functions
func (s *RecommendationService) createRecommendationResults(bookScores map[uint]float64, bookReasons map[uint][]string, strategy string, limit int) ([]*RecommendationResult, error) {
	type scoredBook struct {
		bookID uint
		score  float64
	}

	var scored []scoredBook
	for bookID, score := range bookScores {
		scored = append(scored, scoredBook{bookID: bookID, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	var results []*RecommendationResult
	for i, sb := range scored {
		if i >= limit {
			break
		}

		book, err := s.bookRepo.GetByID(sb.bookID)
		if err != nil {
			continue
		}

		results = append(results, &RecommendationResult{
			Book:     book,
			Score:    sb.score,
			Reasons:  bookReasons[sb.bookID],
			Strategy: strategy,
		})
	}

	return results, nil
}

// deduplicateAndSort removes duplicate recommendations and sorts by score
func (s *RecommendationService) deduplicateAndSort(recommendations []*RecommendationResult, limit int) []*RecommendationResult {
	seen := make(map[uint]bool)
	var unique []*RecommendationResult

	for _, rec := range recommendations {
		if !seen[rec.Book.ID] {
			seen[rec.Book.ID] = true
			unique = append(unique, rec)
		}
	}

	// Sort by score descending
	for i := 0; i < len(unique); i++ {
		for j := i + 1; j < len(unique); j++ {
			if unique[i].Score < unique[j].Score {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	if len(unique) > limit {
		return unique[:limit]
	}

	return unique
}

// InvalidateUserRecommendations removes cached recommendations for a user
func (s *RecommendationService) InvalidateUserRecommendations(userID uint) error {
	if s.cache == nil {
		return nil
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("recommendations:user:%d:*", userID)
	return s.cache.DeleteByPattern(ctx, pattern)
}

// GetRecommendationStats returns statistics about the recommendation engine
func (s *RecommendationService) GetRecommendationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total users
	allRankings, err := s.rankingRepo.GetAllUserRankings()
	if err != nil {
		return nil, err
	}

	stats["total_users"] = len(allRankings)

	// Calculate average rankings per user
	var totalRankings int
	for _, rankings := range allRankings {
		totalRankings += len(rankings)
	}

	if len(allRankings) > 0 {
		stats["avg_rankings_per_user"] = float64(totalRankings) / float64(len(allRankings))
	}

	// Get cache hit ratio if Redis cache is available
	if s.cache != nil {
		stats["cache_info"] = "available"
	}

	return stats, nil
}