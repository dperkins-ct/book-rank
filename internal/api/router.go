package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"bookrank/internal/api/auth"
	"bookrank/internal/api/book"
	"bookrank/internal/api/comparison"
	"bookrank/internal/api/middleware"
	"bookrank/internal/api/ranking"
	"bookrank/internal/api/recommendation"
	"bookrank/internal/repository"
	"bookrank/internal/service"
	"bookrank/pkg/cache"
	authService "bookrank/internal/auth"
)

// Router holds the HTTP router and its dependencies
type Router struct {
	db          *gorm.DB
	authService *authService.AuthService
	logger      *slog.Logger
	cache       cache.Cache
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(db *gorm.DB, authService *authService.AuthService, logger *slog.Logger, cache cache.Cache) *Router {
	return &Router{
		db:          db,
		authService: authService,
		logger:      logger,
		cache:       cache,
	}
}

// NewServer creates a new HTTP server following the documented pattern
func NewServer(
	db *gorm.DB,
	authService *authService.AuthService,
	logger *slog.Logger,
	cache cache.Cache,
) http.Handler {
	mux := mux.NewRouter()
	addRoutes(
		mux,
		db,
		authService,
		logger,
		cache,
	)

	// Apply middleware
	var handler http.Handler = mux
	handler = middleware.CORS(middleware.DefaultCORSConfig())(handler)
	handler = middleware.NewLoggingMiddleware(logger).LogRequests(handler)
	handler = middleware.NewRateLimitMiddleware(100, logger).LimitByUser(handler)

	return handler
}

// addRoutes maps the entire API surface in one place following the documented pattern
func addRoutes(
	router *mux.Router,
	db *gorm.DB,
	authService *authService.AuthService,
	logger *slog.Logger,
	cache cache.Cache,
) {
	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	services := service.NewServices(repos, cache)

	// Create handlers
	authHandler := auth.NewAuthHandler(db, authService, logger)
	bookHandler := book.NewHandlers(services.Book)
	comparisonHandler := comparison.NewHandler(services.Comparison)
	recommendationHandler := recommendation.NewHandler(services.Recommendation)
	rankingHandler := ranking.NewHandler(services.Ranking)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Root endpoint
	router.Handle("/", rootHandler(logger)).Methods("GET")

	// Health endpoint (no auth required)
	router.Handle("/health", healthHandler(db, logger)).Methods("GET")

	// Authentication routes (no auth required)
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/refresh", authHandler.Refresh).Methods("POST", "OPTIONS")

	// Protected routes (auth required)
	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(authMiddleware.RequireAuth)
	protectedRouter.HandleFunc("/me", authHandler.Me).Methods("GET")

	// API v1 routes (for future expansion)
	v1Router := protectedRouter.PathPrefix("/v1").Subrouter()

	// Comparisons routes (direct under /api for frontend compatibility)
	comparisonsRouter := protectedRouter.PathPrefix("/comparisons").Subrouter()
	comparisonsRouter.HandleFunc("", comparisonHandler.GetComparisonHistory).Methods("GET")
	comparisonsRouter.HandleFunc("", comparisonHandler.SubmitComparison).Methods("POST")
	comparisonsRouter.HandleFunc("/pending", comparisonHandler.GetPendingComparisons).Methods("GET")
	comparisonsRouter.HandleFunc("/history", comparisonHandler.GetComparisonHistory).Methods("GET")
	comparisonsRouter.HandleFunc("/book/{bookId:[0-9]+}", comparisonHandler.GetBookComparisons).Methods("GET")
	comparisonsRouter.HandleFunc("/recalculate", comparisonHandler.RecalculateRatings).Methods("POST")
	comparisonsRouter.HandleFunc("/onboarding-status", comparisonHandler.CheckOnboardingStatus).Methods("GET")
	comparisonsRouter.HandleFunc("/random-pair", comparisonHandler.GetRandomBookPair).Methods("GET")

	// Books routes
	booksRouter := protectedRouter.PathPrefix("/books").Subrouter()
	booksRouter.HandleFunc("", bookHandler.GetBooks).Methods("GET")
	booksRouter.HandleFunc("", bookHandler.CreateBook).Methods("POST")
	booksRouter.HandleFunc("/search", bookHandler.SearchBooks).Methods("GET")
	booksRouter.HandleFunc("/{id:[0-9]+}", bookHandler.GetBook).Methods("GET")
	booksRouter.HandleFunc("/{id:[0-9]+}", bookHandler.UpdateBook).Methods("PUT")
	booksRouter.HandleFunc("/{id:[0-9]+}", bookHandler.DeleteBook).Methods("DELETE")
	booksRouter.HandleFunc("/{id:[0-9]+}/stats", bookHandler.GetBookStats).Methods("GET")
	booksRouter.HandleFunc("/{id:[0-9]+}/metadata", bookHandler.RefreshMetadata).Methods("POST")

	// Rankings routes
	rankingsRouter := protectedRouter.PathPrefix("/rankings").Subrouter()
	rankingsRouter.HandleFunc("/me", rankingHandler.GetMyRankings).Methods("GET")
	rankingsRouter.HandleFunc("/top", rankingHandler.GetTopRanked).Methods("GET")
	rankingsRouter.HandleFunc("/stats", rankingHandler.GetRankingStats).Methods("GET")
	rankingsRouter.HandleFunc("/position/{bookId:[0-9]+}", rankingHandler.GetRankingPosition).Methods("GET")
	rankingsRouter.HandleFunc("/compare", rankingHandler.CompareRankings).Methods("GET")
	rankingsRouter.HandleFunc("/initialize", rankingHandler.InitializeRanking).Methods("POST")
	rankingsRouter.HandleFunc("/user/{userId:[0-9]+}", rankingHandler.GetUserRankings).Methods("GET")

	// Rankings routes (also under v1 for future compatibility)
	v1RankingsRouter := v1Router.PathPrefix("/rankings").Subrouter()
	v1RankingsRouter.HandleFunc("/me", rankingHandler.GetMyRankings).Methods("GET")
	v1RankingsRouter.HandleFunc("/top", rankingHandler.GetTopRanked).Methods("GET")
	v1RankingsRouter.HandleFunc("/stats", rankingHandler.GetRankingStats).Methods("GET")
	v1RankingsRouter.HandleFunc("/position/{bookId:[0-9]+}", rankingHandler.GetRankingPosition).Methods("GET")
	v1RankingsRouter.HandleFunc("/compare", rankingHandler.CompareRankings).Methods("GET")
	v1RankingsRouter.HandleFunc("/initialize", rankingHandler.InitializeRanking).Methods("POST")
	v1RankingsRouter.HandleFunc("/user/{userId:[0-9]+}", rankingHandler.GetUserRankings).Methods("GET")

	// Comparisons routes (also under v1 for future compatibility)
	v1ComparisonsRouter := v1Router.PathPrefix("/comparisons").Subrouter()
	v1ComparisonsRouter.HandleFunc("", comparisonHandler.GetComparisonHistory).Methods("GET")
	v1ComparisonsRouter.HandleFunc("", comparisonHandler.SubmitComparison).Methods("POST")
	v1ComparisonsRouter.HandleFunc("/pending", comparisonHandler.GetPendingComparisons).Methods("GET")
	v1ComparisonsRouter.HandleFunc("/history", comparisonHandler.GetComparisonHistory).Methods("GET")
	v1ComparisonsRouter.HandleFunc("/book/{bookId:[0-9]+}", comparisonHandler.GetBookComparisons).Methods("GET")
	v1ComparisonsRouter.HandleFunc("/recalculate", comparisonHandler.RecalculateRatings).Methods("POST")
	v1ComparisonsRouter.HandleFunc("/onboarding-status", comparisonHandler.CheckOnboardingStatus).Methods("GET")
	v1ComparisonsRouter.HandleFunc("/random-pair", comparisonHandler.GetRandomBookPair).Methods("GET")

	// Recommendations routes
	recommendationsRouter := v1Router.PathPrefix("/recommendations").Subrouter()
	recommendationsRouter.HandleFunc("", recommendationHandler.GetRecommendations).Methods("GET")
	recommendationsRouter.HandleFunc("/genre/{genre}", recommendationHandler.GetRecommendationsByGenre).Methods("GET")
	recommendationsRouter.HandleFunc("/similar/{bookId:[0-9]+}", recommendationHandler.GetSimilarBooks).Methods("GET")
	recommendationsRouter.HandleFunc("/refresh", recommendationHandler.RefreshRecommendations).Methods("POST")
	recommendationsRouter.HandleFunc("/stats", recommendationHandler.GetRecommendationStats).Methods("GET")
}


// healthHandler handles health check requests following the documented handler pattern
func healthHandler(db *gorm.DB, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get database connection", "error", err)
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			logger.Error("Database ping failed", "error", err)
			http.Error(w, "Database ping failed", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "bookrank"}`))
	})
}

// rootHandler handles root endpoint requests following the documented handler pattern
func rootHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Root endpoint accessed", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"service": "BookRank API",
			"version": "1.0.0",
			"description": "Book ranking and recommendation system",
			"endpoints": {
				"health": "/health",
				"auth": {
					"register": "/auth/register",
					"login": "/auth/login",
					"refresh": "/auth/refresh"
				},
				"api": {
					"books": "/api/v1/books",
					"recommendations": "/api/v1/recommendations",
					"user": "/api/me"
				}
			},
			"documentation": "See API_DOCUMENTATION.md for complete API reference"
		}`))
	})
}

// notImplementedHandler returns a not implemented response following the documented pattern
func notImplementedHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Not implemented endpoint accessed", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error": "Not Implemented", "message": "This endpoint is not yet implemented", "code": 501}`))
	})
}