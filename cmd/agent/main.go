package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/agent/healthcheck"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

const (
	_defaultCollectionInterval = 60
	_defaultHealthInterval     = 15
	_defaultLogLevel           = "debug"
	_defaultLogFormat          = "json"
)

var (
	serverURL           string
	apiToken            string
	interval            int
	logFormat           string
	logLevel            string
	healthCheckInterval int
)

type agentError struct {
	op  string
	err error
}

func (e *agentError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("agent %s: %v", e.op, e.err)
	}
	return fmt.Sprintf("agent %s", e.op)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "g0s-agent",
		Short: "g0s agent",
		Long:  `g0s agent`,
		RunE:  runAgent,
	}

	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "", "Server URL to send metrics to (required)")
	rootCmd.Flags().StringVarP(&apiToken, "token", "t", "", "API token for authentication (required)")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", _defaultCollectionInterval, "Collection interval in seconds")
	rootCmd.Flags().StringVar(&logFormat, "log-format", _defaultLogFormat, "Log format: json or console")
	rootCmd.Flags().StringVar(&logLevel, "log-level", _defaultLogLevel, "Log level: debug, info, warn, error")
	rootCmd.Flags().IntVar(&healthCheckInterval, "health-check-interval", _defaultHealthInterval, "Health check interval in seconds")

	rootCmd.MarkFlagRequired("server")
	rootCmd.MarkFlagRequired("token")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAgent(_ *cobra.Command, _ []string) error {
	logger.InitLogger(logger.Config{
		Level:     logLevel,
		Format:    logFormat,
		Component: "agent",
	})
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthCheckService := healthcheck.New(healthcheck.Config{
		ServerURL: serverURL,
		Token:     apiToken,
		Interval:  time.Duration(healthCheckInterval) * time.Second,
	})

	go healthCheckService.Start(ctx)

	return runMetricsCollection(ctx, healthCheckService)
}

func runMetricsCollection(ctx context.Context, healthService *healthcheck.Service) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	logger.Info("Agent running",
		zap.String("server", serverURL),
		zap.Duration("collection_interval", time.Duration(interval)*time.Second),
		zap.Duration("health_interval", time.Duration(healthCheckInterval)*time.Second))

	for {
		select {
		case <-ctx.Done():
			return nil
		case sig := <-signals:
			logger.Info("Shutting down", zap.String("signal", sig.String()))
			return nil
		case <-ticker.C:
			if err := collectAndSendMetrics(healthService.IsHealthy()); err != nil {
				return &agentError{op: "collect_metrics", err: err}
			}
		}
	}
}

func collectAndSendMetrics(isServerHealthy bool) error {
	if !isServerHealthy {
		logger.Warn("Skipping metrics transmission, server unhealthy")
		return nil
	}

	logger.Debug("Sending metrics", zap.String("url", serverURL))
	return nil
}
