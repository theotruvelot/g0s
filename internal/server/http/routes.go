package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	servermiddleware "github.com/theotruvelot/g0s/internal/server/middleware"
)

// RegisterRoutes sets up all HTTP routes with the given handler
func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(servermiddleware.RequestLogger())
	r.Use(middleware.Recoverer)

	// Health check route (public)
	r.Get("/health", h.HandleHealth)

	// API routes group (for testing)
	//TODO: change routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Server status
			r.Get("/status", h.HandleStatus)

			// Agent endpoints
			r.Route("/agent", func(r chi.Router) {
				r.Post("/register", h.HandleAgentRegister)
				r.Post("/metrics", h.HandleMetrics)
			})
		})
	})

	return r
}
