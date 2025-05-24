package grpc

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Handler contains gRPC service implementations
type Handler struct {
	logger *zap.Logger
	// Add other dependencies here (database, services, etc.)
}

// New creates a new gRPC handler
func New(logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// RegisterServices registers all gRPC services with the server
func (h *Handler) RegisterServices(server *grpc.Server) {
	// TODO: Register your gRPC services here
	// Example:
	// pb.RegisterYourServiceServer(server, h)

	h.logger.Debug("gRPC services registered")
}
