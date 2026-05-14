package service

import (
	"bookrank/internal/models"
	"bookrank/internal/repository"
	"bookrank/pkg/cache"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// RecommendationService handles business logic for recommendation operations
type RecommendationService struct {
	rankingRepo repository.RankingRepository
	bookRepo    repository.BookRepository
	cache       cache.Cache
	similarity  *SimilarityCalculator
	cacheTTL    time.Duration
}

// NewRecommendationService creates a new RecommendationService
func NewRecommendationService(rankingRepo repository.RankingRepository, bookRepo repository.BookRepository, cache cache.Cache, cacheTTL time.Duration) *RecommendationService {
	return &RecommendationService{
		rankingRepo: rankingRepo,
		bookRepo:    bookRepo,
		cache:       cache,
		similarity:  NewSimilarityCalculator(),
		cacheTTL:    cacheTTL,
	}
}

// RecommendationResult represents a book recommendation with score and reasoning
type RecommendationResult struct {
	Book     *models.Book `json:"book"`
	Score    float64      `json:"score"`
	Reasons  []string     `json:"reasons"`
	Strategy string       `json:"strategy"`
}

// GetRecommendations returns personalized book recommendations for a user
func (s *RecommendationService) GetRecommendations(userID uint, limit int) ([]*RecommendationResult, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("recommendations:user:%d:limit:%d", userID, limit)

	// Try cache first
	var cachedResults []*RecommendationResult
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &cachedResults); err == nil {
			return cachedResults, nil
		}
	}

	// Get user's rankings
	userRankings, err := s.rankingRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if len(userRankings) == 0 {
		return s.getPopularRecommendations(limit)
	}

	// Get recommendations using multiple strategies
	recommendations := make([]*RecommendationResult, 0)

	// 1. Collaborative Filtering (60% of results)
	collabLimit := int(math.Ceil(float64(limit) * 0.6))
	collabRecs, err := s.getCollaborativeRecommendations(userID, userRankings, collabLimit)
	if err == nil {
		recommendations = append(recommendations, collabRecs...)
	}

	// 2. Content-Based Filtering (30% of results)
	contentLimit := int(math.Ceil(float64(limit) * 0.3))
	contentRecs, err := s.getContentBasedRecommendations(userID, userRankings, contentLimit)
	if err == nil {
		recommendations = append(recommendations, contentRecs...)
	}

	// 3. Popular books (10% of results)
	popularLimit := limit - len(recommendations)
	if popularLimit > 0 {
		popularRecs, err := s.getPopularRecommendations(popularLimit)
		if err == nil {
			recommendations = append(recommendations, popularRecs...)
		}
	}

	// Remove duplicates and sort by score
	recommendations = s.deduplicateAndSort(recommendations, limit)

	// Cache results
	if s.cache != nil {
		s.cache.Set(ctx, cacheKey, recommendations, s.cacheTTL)
	}

	return recommendations, nil
}

// getCollaborativeRecommendations implements collaborative filtering
func (s *RecommendationService) getCollaborativeRecommendations(userID uint, userRankings []*models.Ranking, limit int) ([]*RecommendationResult, error) {
	// Get all user rankings
	allUserRankings, err := s.rankingRepo.GetAllUserRankings()
	if err != nil {
		return nil, err
	}

	// Find similar users
	similarUsers := s.similarity.FindTopSimilarUsers(userRankings, allUserRankings, 20, 3)
	if len(similarUsers) == 0 {
		return nil, errors.New("no similar users found")
	}

	// Get books user hasn't rated
	unratedBooks, err := s.rankingRepo.GetUnratedBooksByUser(userID)
	if err != nil {
		return nil, err
	}

	bookScores := make(map[uint]float64)
	bookReasons := make(map[uint][]string)

	// Calculate weighted scores for unrated books
	for _, book := range unratedBooks {
		var totalWeight float64
		var weightedSum float64

		for _, simUser := range similarUsers {
			if userRankings := allUserRankings[simUser.UserID]; userRankings != nil {
				for _, ranking := range userRankings {
					if ranking.BookID == book.ID {
						weight := simUser.Similarity * math.Max(0.1, float64(simUser.SharedBooks)/10.0)
						weightedSum += float64(ranking.Score) * weight
						totalWeight += weight
						break
					}
				}
			}
		}

		if totalWeight > 0 {
			score := weightedSum / totalWeight
			bookScores[book.ID] = score / 3000.0 // Normalize to 0-1
			bookReasons[book.ID] = []string{
				fmt.Sprintf("Recommended by %d similar users", len(similarUsers)),
				"Based on collaborative filtering",
			}
		}
	}

	return s.createRecommendationResults(bookScores, bookReasons, "collaborative_filtering", limit)
}