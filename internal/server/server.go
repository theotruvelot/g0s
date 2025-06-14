package server

import (
	"context"
	"fmt"
	"github.com/theotruvelot/g0s/internal/server/auth"
	"github.com/theotruvelot/g0s/internal/server/grpc"
	"github.com/theotruvelot/g0s/internal/server/middleware"
	"github.com/theotruvelot/g0s/internal/server/service"
	"github.com/theotruvelot/g0s/internal/server/storage/database"
	"github.com/theotruvelot/g0s/internal/server/storage/metrics"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	grpclib "google.golang.org/grpc"
	"net"
)

// Config holds server configuration
type Config struct {
	GRPCAddr         string
	LogLevel         string
	LogFormat        string
	VMEndpoint       string
	JWTSecret        string
	JWTRefreshSecret string
}

// Server represents the g0s server
type Server struct {
	cfg         Config
	grpc        *grpclib.Server
	store       *metrics.Manager
	handler     *grpc.Handler
	authService *service.AuthService
}

// New creates a new server instance
func New(cfg Config) (*Server, error) {
	// Initialize dependencies
	store := metrics.NewMetricsManager(cfg.VMEndpoint)

	// Create auth dependencies using the global database connection
	db := database.GetDB()
	userRepo := database.NewUserRepository(db)
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTRefreshSecret)
	authService := service.NewAuthService(*userRepo, *jwtService)

	healthCheckService := service.NewHealthCheckService()

	// Create the main handler orchestrator
	handler := grpc.New(store, authService, healthCheckService)

	// Setup authentication config
	authConfig := middleware.DefaultAuthConfig()

	// Create gRPC server with middlewares
	grpcServer := grpclib.NewServer(
		grpclib.ChainUnaryInterceptor(
			middleware.LoggingUnaryInterceptor(),
			middleware.AuthUnaryInterceptor(authConfig),
		),
		grpclib.ChainStreamInterceptor(
			middleware.LoggingStreamInterceptor(),
			middleware.AuthStreamInterceptor(authConfig),
		),
	)

	s := &Server{
		cfg:         cfg,
		store:       store,
		handler:     handler,
		grpc:        grpcServer,
		authService: authService,
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
			logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping gRPC server")

	s.grpc.GracefulStop()

	return nil
}

// NotifyShutdown notifies all connected clients that the server is about to shut down
func (s *Server) NotifyShutdown() {
	logger.Info("Notifying clients about server shutdown")
	s.handler.NotifyShutdown()
}
