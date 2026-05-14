package service

import (
	"bookrank/internal/repository"
	"bookrank/pkg/cache"
	"time"
)

// Services holds all business logic services
type Services struct {
	User         *UserService
	Book         *BookService
	Ranking      *RankingService
	Auth         *AuthService
	External     *ExternalAPIService
	Recommendation *RecommendationService
	ELO          *ELOService
	Comparison   *ComparisonService
}

// NewServices creates a new Services instance
func NewServices(repos *repository.Repositories, cache cache.Cache) *Services {
	eloService := NewELOService()
	cacheTTL := time.Hour // 1 hour default cache TTL

	return &Services{
		User:         NewUserService(repos.User),
		Book:         NewBookService(repos.Book),
		Ranking:      NewRankingService(repos.Ranking, eloService),
		Auth:         NewAuthService(repos.User),
		External:     NewExternalAPIService(),
		Recommendation: NewRecommendationService(repos.Ranking, repos.Book, cache, cacheTTL),
		ELO:          eloService,
		Comparison:   NewComparisonService(repos.Comparison, repos.Ranking, repos.Book, eloService),
	}
}