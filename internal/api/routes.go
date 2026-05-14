package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(router *mux.Router, handlers *Handlers) {
	// Health check endpoint
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// API v1 routes
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Authentication routes
	auth := v1.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	auth.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	// Protected routes (require authentication)
	protected := v1.PathPrefix("").Subrouter()
	// protected.Use(handlers.AuthMiddleware) // Will be implemented later

	// User routes
	users := protected.PathPrefix("/users").Subrouter()
	users.HandleFunc("/me", handlers.GetCurrentUser).Methods("GET")
	users.HandleFunc("/me", handlers.UpdateCurrentUser).Methods("PUT")

	// Book routes
	books := protected.PathPrefix("/books").Subrouter()
	books.HandleFunc("", handlers.CreateBook).Methods("POST")
	books.HandleFunc("", handlers.GetBooks).Methods("GET")
	books.HandleFunc("/{id:[0-9]+}", handlers.GetBook).Methods("GET")
	books.HandleFunc("/{id:[0-9]+}", handlers.UpdateBook).Methods("PUT")
	books.HandleFunc("/{id:[0-9]+}", handlers.DeleteBook).Methods("DELETE")

	// Ranking routes
	rankings := protected.PathPrefix("/rankings").Subrouter()
	rankings.HandleFunc("", handlers.GetUserRankings).Methods("GET")
	rankings.HandleFunc("/compare", handlers.CompareBooks).Methods("POST")

	// Recommendation routes
	recommendations := protected.PathPrefix("/recommendations").Subrouter()
	recommendations.HandleFunc("", handlers.Recommendation.GetRecommendations).Methods("GET")
	recommendations.HandleFunc("/genre/{genre}", handlers.Recommendation.GetRecommendationsByGenre).Methods("GET")
	recommendations.HandleFunc("/similar/{bookId:[0-9]+}", handlers.Recommendation.GetSimilarBooks).Methods("GET")
	recommendations.HandleFunc("/refresh", handlers.Recommendation.RefreshRecommendations).Methods("POST")
	recommendations.HandleFunc("/stats", handlers.Recommendation.GetRecommendationStats).Methods("GET")
}

// healthCheck is a simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}