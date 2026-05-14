package api

import (
	"bookrank/internal/service"
	"bookrank/internal/api/recommendation"
)

// Handlers holds all HTTP handlers for the application
type Handlers struct {
	services *service.Services
	Recommendation *recommendation.Handler
}

// NewHandlers creates a new Handlers instance
func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		services: services,
		Recommendation: recommendation.NewHandler(services.Recommendation),
	}
}