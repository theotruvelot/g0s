package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	grpchandler "github.com/theotruvelot/g0s/internal/server/grpc"
	httphandler "github.com/theotruvelot/g0s/internal/server/http"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	HTTPAddr  string
	GRPCAddr  string
	LogLevel  string
	LogFormat string
}

// Server represents the main server instance
type Server struct {
	cfg    Config
	logger *zap.Logger

	httpServer *http.Server
	grpcServer *grpc.Server

	// Handlers
	httpHandler *httphandler.Handler
	grpcHandler *grpchandler.Handler

	// For graceful shutdown
	wg sync.WaitGroup
}

// New creates a new server instance
func New(cfg Config, logger *zap.Logger) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize handlers
	s.httpHandler = httphandler.New(logger)
	s.grpcHandler = grpchandler.New(logger)

	s.setupHTTPServer()
	s.setupGRPCServer()

	logger.Info("Server configured",
		zap.String("http_addr", cfg.HTTPAddr),
		zap.String("grpc_addr", cfg.GRPCAddr))

	return s, nil
}

func (s *Server) setupHTTPServer() {
	router := s.httpHandler.RegisterRoutes()

	s.httpServer = &http.Server{
		Addr:    s.cfg.HTTPAddr,
		Handler: router,
	}

	s.logger.Debug("HTTP server configured", zap.String("addr", s.cfg.HTTPAddr))
}

func (s *Server) setupGRPCServer() {
	s.grpcServer = grpc.NewServer()

	// Register gRPC services through the handler
	s.grpcHandler.RegisterServices(s.grpcServer)

	s.logger.Debug("gRPC server configured", zap.String("addr", s.cfg.GRPCAddr))
}

func (s *Server) Start() error {
	s.logger.Info("Starting servers",
		zap.String("http_addr", s.cfg.HTTPAddr),
		zap.String("grpc_addr", s.cfg.GRPCAddr))

	// Start HTTP server
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Info("HTTP server starting", zap.String("addr", s.cfg.HTTPAddr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Start gRPC server
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Info("gRPC server starting", zap.String("addr", s.cfg.GRPCAddr))
		lis, err := net.Listen("tcp", s.cfg.GRPCAddr)
		if err != nil {
			s.logger.Error("Failed to listen on gRPC port", zap.Error(err))
			return
		}
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops both servers
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping servers")

	// Stop HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error", zap.Error(err))
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}
	s.logger.Info("HTTP server stopped")

	// Stop gRPC server
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Sync logger before exit
	s.logger.Sync()

	return nil
}
