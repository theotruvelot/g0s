package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theotruvelot/g0s/internal/server/storage/database"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/server"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

const (
	_defaultHTTPAddr         = ":8080"
	_defaultGRPCAddr         = ":9090"
	_defaultLogLevel         = "info"
	_defaultLogFormat        = "json"
	_defaultVMEndpoint       = "http://localhost:8428"
	_defaultDSN              = "postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable"
	_defaultJWTSecret        = "mongigasecret"
	_defaultJWTRefreshSecret = "mongigasecretrefresh"
	_shutdownTimeout         = 5 * time.Second
)

var (
	httpAddr         string
	grpcAddr         string
	logLevel         string
	logFormat        string
	vmEndpoint       string
	dsn              string
	jwtSecret        string
	jwtRefreshSecret string
)

type serverError struct {
	op  string
	err error
}

func (e *serverError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("server %s: %v", e.op, e.err)
	}
	return fmt.Sprintf("server %s", e.op)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "g0s-server",
		Short: "g0s server",
		Long:  `g0s server`,
		RunE:  runServer,
	}

	rootCmd.Flags().StringVar(&httpAddr, "http-addr", _defaultHTTPAddr, "HTTP server address")
	rootCmd.Flags().StringVar(&grpcAddr, "grpc-addr", _defaultGRPCAddr, "gRPC server address")
	rootCmd.Flags().StringVar(&logLevel, "log-level", _defaultLogLevel, "Log level: debug, info, warn, error")
	rootCmd.Flags().StringVar(&logFormat, "log-format", _defaultLogFormat, "Log format: json or console")
	rootCmd.Flags().StringVar(&vmEndpoint, "vm-endpoint", _defaultVMEndpoint, "VictoriaMetrics endpoint")
	rootCmd.Flags().StringVar(&dsn, "dsn", _defaultDSN, "Database DSN")
	rootCmd.Flags().StringVar(&jwtSecret, "jwt-secret", _defaultJWTSecret, "JWT secret for signing tokens")
	rootCmd.Flags().StringVar(&jwtRefreshSecret, "jwt-refresh-secret", _defaultJWTRefreshSecret, "JWT secret for signing refresh tokens")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer(_ *cobra.Command, _ []string) error {
	logger.InitLogger(logger.Config{
		Level:     logLevel,
		Format:    logFormat,
		Component: "server",
	})
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := server.Config{
		GRPCAddr:         grpcAddr,
		LogLevel:         logLevel,
		LogFormat:        logFormat,
		VMEndpoint:       vmEndpoint,
		JWTSecret:        jwtSecret,
		JWTRefreshSecret: jwtRefreshSecret,
	}

	// Initialize database connection
	_, err := database.Init(dsn)
	if err != nil {
		return &serverError{op: "init database", err: err}
	}

	// Ensure database connection is closed when server shuts down
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		} else {
			logger.Info("Database connection closed successfully")
		}
	}()

	return runServerWithConfig(ctx, cfg)
}

func runServerWithConfig(ctx context.Context, cfg server.Config) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	logger.Info("Starting g0s-server",
		zap.String("grpc_addr", cfg.GRPCAddr),
		zap.String("log_level", cfg.LogLevel),
		zap.String("log_format", cfg.LogFormat))

	srv, err := server.New(cfg)
	if err != nil {
		return &serverError{op: "create", err: err}
	}

	if err := srv.Start(); err != nil {
		return &serverError{op: "start", err: err}
	}

	logger.Info("Server started successfully",
		zap.String("grpc_addr", cfg.GRPCAddr))

	for {
		select {
		case <-ctx.Done():
			return shutdownServer(srv)
		case sig := <-signals:
			logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
			return shutdownServer(srv)
		}
	}
}

func shutdownServer(srv *server.Server) error {
	logger.Info("Initiating graceful shutdown", zap.Duration("timeout", _shutdownTimeout))

	srv.NotifyShutdown()

	ctx, cancel := context.WithTimeout(context.Background(), _shutdownTimeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		return &serverError{op: "shutdown", err: err}
	}

	logger.Info("Server stopped successfully")
	return nil
}
