package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"bookrank/internal/api/auth"
	"bookrank/internal/api/book"
	"bookrank/internal/api/middleware"
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

// SetupRoutes configures all routes and middleware
func (rt *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Create middleware instances
	authMiddleware := middleware.NewAuthMiddleware(rt.authService, rt.logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(rt.logger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(100, rt.logger) // 100 requests per minute

	// Initialize repositories
	repos := repository.NewRepositories(rt.db)

	// Initialize services
	services := service.NewServices(repos, rt.cache)

	// Create handlers
	authHandler := auth.NewAuthHandler(rt.db, rt.authService, rt.logger)
	bookHandler := book.NewHandlers(services.Book)
	recommendationHandler := recommendation.NewHandler(services.Recommendation)

	// Global middleware (applied to all routes)
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	router.Use(loggingMiddleware.LogRequests)
	router.Use(rateLimitMiddleware.LimitByUser)

	// Root endpoint
	router.HandleFunc("/", rt.rootHandler).Methods("GET")

	// Health endpoint (no auth required)
	router.HandleFunc("/health", rt.healthHandler).Methods("GET")

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

	// Rankings routes (placeholder for future implementation)
	v1Router.HandleFunc("/rankings", rt.notImplementedHandler).Methods("GET", "POST")
	v1Router.HandleFunc("/rankings/{id}", rt.notImplementedHandler).Methods("GET", "PUT", "DELETE")

	// Comparisons routes (placeholder for future implementation)
	v1Router.HandleFunc("/comparisons", rt.notImplementedHandler).Methods("GET", "POST")

	// Recommendations routes
	recommendationsRouter := v1Router.PathPrefix("/recommendations").Subrouter()
	recommendationsRouter.HandleFunc("", recommendationHandler.GetRecommendations).Methods("GET")
	recommendationsRouter.HandleFunc("/genre/{genre}", recommendationHandler.GetRecommendationsByGenre).Methods("GET")
	recommendationsRouter.HandleFunc("/similar/{bookId:[0-9]+}", recommendationHandler.GetSimilarBooks).Methods("GET")
	recommendationsRouter.HandleFunc("/refresh", recommendationHandler.RefreshRecommendations).Methods("POST")
	recommendationsRouter.HandleFunc("/stats", recommendationHandler.GetRecommendationStats).Methods("GET")

	return router
}

// healthHandler handles health check requests
func (rt *Router) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	sqlDB, err := rt.db.DB()
	if err != nil {
		rt.logger.Error("Failed to get database connection", "error", err)
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		rt.logger.Error("Database ping failed", "error", err)
		http.Error(w, "Database ping failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "bookrank"}`))
}

// rootHandler handles root endpoint requests
func (rt *Router) rootHandler(w http.ResponseWriter, r *http.Request) {
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
}

// notImplementedHandler returns a not implemented response
func (rt *Router) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "Not Implemented", "message": "This endpoint is not yet implemented", "code": 501}`))
}