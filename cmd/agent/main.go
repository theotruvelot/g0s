package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/agent/collector"
	"github.com/theotruvelot/g0s/internal/agent/healthcheck"
	"github.com/theotruvelot/g0s/internal/agent/model/metric"
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

	collectors := initCollectors()

	for {
		select {
		case <-ctx.Done():
			return nil
		case sig := <-signals:
			logger.Info("Shutting down", zap.String("signal", sig.String()))
			return nil
		case <-ticker.C:
			go func() {
				if err := collectAndSendMetrics(healthService.IsHealthy(), collectors); err != nil {
					logger.Error("Failed to collect metrics", zap.Error(err))
				}
			}()
		}
	}
}

type collectors struct {
	cpu     *collector.CPUCollector
	ram     *collector.RAMCollector
	disk    *collector.DiskCollector
	network *collector.NetworkCollector
	host    *collector.HostCollector
	docker  *collector.DockerCollector
}

func initCollectors() *collectors {
	log := logger.GetLogger()
	dockerCollector, err := collector.NewDockerCollector(log)
	if err != nil {
		log.Error("Failed to initialize Docker collector", zap.Error(err))
	}

	return &collectors{
		cpu:     collector.NewCPUCollector(log),
		ram:     collector.NewRAMCollector(log),
		disk:    collector.NewDiskCollector(log),
		network: collector.NewNetworkCollector(log),
		host:    collector.NewHostCollector(log),
		docker:  dockerCollector,
	}
}

// Structure pour stocker les résultats de collecte
type collectionResult struct {
	cpuMetrics     []metric.CPUMetrics
	ramMetrics     metric.RamMetrics
	diskMetrics    []metric.DiskMetrics
	networkMetrics []metric.NetworkMetrics
	hostMetrics    metric.HostMetrics
	dockerMetrics  []metric.DockerMetrics
	errors         []error
}

func collectAndSendMetrics(isServerHealthy bool, c *collectors) error {
	if !isServerHealthy {
		logger.Warn("Skipping metrics transmission, server unhealthy")
		return nil
	}

	result := &collectionResult{
		errors: make([]error, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Helper function to add errors in a thread-safe way
	addError := func(err error) {
		if err != nil {
			mu.Lock()
			result.errors = append(result.errors, err)
			mu.Unlock()
		}
	}

	// Collect CPU metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		cpuMetrics, err := c.cpu.Collect()
		if err != nil {
			addError(fmt.Errorf("failed to collect CPU metrics: %w", err))
			return
		}
		mu.Lock()
		result.cpuMetrics = cpuMetrics
		mu.Unlock()
	}()

	// Collect RAM metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		ramMetrics, err := c.ram.Collect()
		if err != nil {
			addError(fmt.Errorf("failed to collect RAM metrics: %w", err))
			return
		}
		mu.Lock()
		result.ramMetrics = ramMetrics
		mu.Unlock()
	}()

	// Collect Disk metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		diskMetrics, err := c.disk.Collect()
		if err != nil {
			addError(fmt.Errorf("failed to collect disk metrics: %w", err))
			return
		}
		mu.Lock()
		result.diskMetrics = diskMetrics
		mu.Unlock()
	}()

	// Collect Network metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		networkMetrics, err := c.network.Collect()
		if err != nil {
			addError(fmt.Errorf("failed to collect network metrics: %w", err))
			return
		}
		mu.Lock()
		result.networkMetrics = networkMetrics
		mu.Unlock()
	}()

	// Collect Host metrics in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		hostMetrics, err := c.host.Collect()
		if err != nil {
			addError(fmt.Errorf("failed to collect host metrics: %w", err))
			return
		}
		mu.Lock()
		result.hostMetrics = hostMetrics
		mu.Unlock()
	}()

	// Collect Docker metrics in parallel (optional)
	if c.docker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dockerMetrics, err := c.docker.Collect()
			if err != nil {
				logger.Error("Failed to collect Docker metrics", zap.Error(err))
				// Don't add to errors, continue with other metrics
				return
			}
			mu.Lock()
			result.dockerMetrics = dockerMetrics
			mu.Unlock()
		}()
	}

	// Wait for all collections to be done
	wg.Wait()

	// Check if there are any critical errors
	if len(result.errors) > 0 {
		logger.Error("Errors during metrics collection", zap.Any("errors", result.errors))
		// Return the first error to maintain compatibility
		return result.errors[0]
	}

	metrics := metric.MetricsPayload{
		CPU:       result.cpuMetrics,
		RAM:       result.ramMetrics,
		Disk:      result.diskMetrics,
		Network:   result.networkMetrics,
		Host:      result.hostMetrics,
		Docker:    result.dockerMetrics,
		Timestamp: time.Now(),
	}

	// Format metrics as pretty JSON for logging
	// TODO: remove this
	_, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to format metrics for logging: %w", err)
	}

	//log docker metrics
	logger.Debug("Docker Metrics", zap.Any("metrics", result.dockerMetrics))

	logger.Debug("Sending metrics to server", zap.String("url", serverURL))

	return nil
}
