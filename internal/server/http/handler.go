package http

import (
	"net/http"

	"go.uber.org/zap"
)

// Handler contains HTTP route handlers
type Handler struct {
	logger *zap.Logger
	// Add other dependencies here (database, services, etc.)
}

// New creates a new HTTP handler
func New(logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// HandleHealth is a basic health check endpoint
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"g0s-server"}`))
}

// HandleStatus provides server status information
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Status endpoint called")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"running","service":"g0s-server"}`))
}

// HandleMetrics handles metrics submission from agents
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Metrics endpoint called")
	// TODO: Implement metrics handling logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}

// HandleAgentRegister handles agent registration
func (h *Handler) HandleAgentRegister(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Agent register endpoint called")
	// TODO: Implement agent registration logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"registered"}`))
}
