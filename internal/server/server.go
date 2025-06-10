package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/theotruvelot/g0s/internal/server/storage/metrics"

	"github.com/theotruvelot/g0s/internal/server/grpc"
	"go.uber.org/zap"
	grpclib "google.golang.org/grpc"
)

const shutdownTimeout = 5 * time.Second

// Config holds server configuration
type Config struct {
	GRPCAddr   string
	LogLevel   string
	LogFormat  string
	VMEndpoint string
}

// Server represents the g0s server
type Server struct {
	cfg     Config
	logger  *zap.Logger
	grpc    *grpclib.Server
	store   *metrics.MetricsManager
	handler *grpc.Handler
}

// New creates a new server instance
func New(cfg Config, logger *zap.Logger) (*Server, error) {
	if cfg.VMEndpoint == "" {
		cfg.VMEndpoint = "http://localhost:8428"
	}

	store := metrics.NewMetricsManager(cfg.VMEndpoint, logger)
	handler := grpc.New(logger, store)

	s := &Server{
		cfg:     cfg,
		logger:  logger,
		store:   store,
		handler: handler,
		grpc:    grpclib.NewServer(),
	}

	handler.RegisterServices(s.grpc)

	return s, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start gRPC server
	lis, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := s.grpc.Serve(lis); err != nil {
			s.logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")

	s.grpc.GracefulStop()

	return nil
}

// NotifyShutdown notifies all connected clients that the server is about to shut down
func (s *Server) NotifyShutdown() {
	s.logger.Info("Notifying clients about server shutdown")
	s.handler.NotifyShutdown()
}
