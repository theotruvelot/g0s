package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/server"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

const (
	_defaultHTTPAddr  = ":8080"
	_defaultGRPCAddr  = ":9090"
	_defaultLogLevel  = "info"
	_defaultLogFormat = "json"
	_shutdownTimeout  = 30 * time.Second
)

var (
	httpAddr  string
	grpcAddr  string
	logLevel  string
	logFormat string
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
		HTTPAddr:  httpAddr,
		GRPCAddr:  grpcAddr,
		LogLevel:  logLevel,
		LogFormat: logFormat,
	}

	return runServerWithConfig(ctx, cfg)
}

func runServerWithConfig(ctx context.Context, cfg server.Config) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	logger.Info("Starting g0s-server",
		zap.String("http_addr", cfg.HTTPAddr),
		zap.String("grpc_addr", cfg.GRPCAddr),
		zap.String("log_level", cfg.LogLevel),
		zap.String("log_format", cfg.LogFormat))

	srv, err := server.New(cfg, logger.GetLogger())
	if err != nil {
		return &serverError{op: "create", err: err}
	}

	if err := srv.Start(); err != nil {
		return &serverError{op: "start", err: err}
	}

	logger.Info("Server started successfully",
		zap.String("http_addr", cfg.HTTPAddr),
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

	ctx, cancel := context.WithTimeout(context.Background(), _shutdownTimeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		return &serverError{op: "shutdown", err: err}
	}

	logger.Info("Server stopped successfully")
	return nil
}
