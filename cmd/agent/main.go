package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/theotruvelot/g0s/internal/agent/model"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/agent/collector"
	"github.com/theotruvelot/g0s/internal/agent/converter"
	"github.com/theotruvelot/g0s/internal/agent/healthcheck"
	"github.com/theotruvelot/g0s/pkg/logger"
	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_defaultCollectionInterval = 180
	_defaultHealthInterval     = 30
	_defaultLogLevel           = "info"
	_defaultLogFormat          = "json"
	_defaultGRPCPort           = "9090"

	_minConnectTimeout = 10 * time.Second
	_keepaliveTime     = 60 * time.Second
	_keepaliveTimeout  = 15 * time.Second
	_initialWindowSize = 1 << 18
	_initialConnWindow = 1 << 18
	_maxBackoffDelay   = 60 * time.Second
	_backoffMultiplier = 2.0
)

var (
	grpcAddr            string
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

	rootCmd.Flags().StringVar(&grpcAddr, "grpc-addr", _defaultGRPCPort, "Server gRPC address (required, e.g. localhost:9090)")
	rootCmd.Flags().StringVarP(&apiToken, "token", "t", "", "API token for authentication (required)")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", _defaultCollectionInterval, "Collection interval in seconds")
	rootCmd.Flags().StringVar(&logFormat, "log-format", _defaultLogFormat, "Log format: json or console")
	rootCmd.Flags().StringVar(&logLevel, "log-level", _defaultLogLevel, "Log level: debug, info, warn, error")
	rootCmd.Flags().IntVar(&healthCheckInterval, "health-check-interval", _defaultHealthInterval, "Health check interval in seconds")

	err := rootCmd.MarkFlagRequired("grpc-addr")
	err = rootCmd.MarkFlagRequired("token")

	if err = rootCmd.Execute(); err != nil {
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
	defer func() {
		err := logger.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
		} else {
			fmt.Println("Logger synced successfully")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize collectors
	collectors := initCollectors()
	defer cleanupCollectors(collectors)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()
	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                _keepaliveTime,
			Timeout:             _keepaliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithInitialWindowSize(_initialWindowSize),
		grpc.WithInitialConnWindowSize(_initialConnWindow),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: _backoffMultiplier,
				Jitter:     0.2,
				MaxDelay:   _maxBackoffDelay,
			},
			MinConnectTimeout: _minConnectTimeout,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer conn.Close()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid.New().String() // Fallback to UUID if hostname cannot be retrieved
		logger.Error("Failed to get hostname set hostname to UUID", zap.Error(err), zap.String("hostname", hostname))
	}

	healthService := healthcheck.New(conn, logger.GetLogger(), hostname)
	if err = healthService.Start(ctx, time.Duration(healthCheckInterval)*time.Second); err != nil {
		return fmt.Errorf("failed to start health check service: %w", err)
	}

	metricClient := pb.NewMetricServiceClient(conn)
	if err = runMetricsCollection(ctx, healthService, metricClient, collectors); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("Metrics collection stopped due to shutdown")
			return nil
		}
		return fmt.Errorf("metrics collection error: %w", err)
	}

	logger.Info("Agent shutdown complete")
	return nil
}

func runMetricsCollection(ctx context.Context, healthService *healthcheck.Service, client pb.MetricServiceClient, collectors *collectors) error {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	logger.Info("Starting metrics collection",
		zap.String("grpc_addr", grpcAddr),
		zap.Duration("collection_interval", time.Duration(interval)*time.Second),
		zap.Duration("health_interval", time.Duration(healthCheckInterval)*time.Second))

	var stream pb.MetricService_StreamMetricsClient
	var lastHealthy bool

	for {
		select {
		case <-ctx.Done():
			if stream != nil {
				err := stream.CloseSend()
				if err != nil {
					return err
				}
			}
			return ctx.Err()

		case <-ticker.C:
			isHealthy := healthService.IsHealthy()

			if isHealthy != lastHealthy {
				if isHealthy {
					logger.Info("Server became healthy, resuming metrics collection")
				} else {
					logger.Info("Server became unhealthy, pausing metrics collection")
					if stream != nil {
						err := stream.CloseSend()
						if err != nil {
							return err
						}
						stream = nil
					}
				}
				lastHealthy = isHealthy
			}

			if !isHealthy {
				logger.Debug("Skipping metrics collection, server is unhealthy")
				continue
			}

			if stream == nil {
				newStream, err := connectWithRetry(ctx, client)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return err
					}
					logger.Error("Failed to create metrics stream", zap.Error(err))
					continue
				}
				stream = newStream
				logger.Info("Metrics stream established")
			}

			if err := collectAndSendMetrics(true, collectors, stream); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				logger.Error("Failed to send metrics, closing stream", zap.Error(err))
				err := stream.CloseSend()
				if err != nil {
					return err
				}
				stream = nil
			}
		}
	}
}

func connectWithRetry(ctx context.Context, client pb.MetricServiceClient) (pb.MetricService_StreamMetricsClient, error) {
	var retryCount int
	backoffConfig := backoff.Config{
		BaseDelay:  1.0 * time.Second,
		Multiplier: _backoffMultiplier,
		Jitter:     0.2,
		MaxDelay:   _maxBackoffDelay,
	}

	for {
		stream, err := client.StreamMetrics(ctx)
		if err == nil {
			return stream, nil
		}

		retryCount++
		delay := backoffConfig.BaseDelay * time.Duration(float64(backoffConfig.BaseDelay)*float64(retryCount)*backoffConfig.Multiplier)
		if delay > backoffConfig.MaxDelay {
			delay = backoffConfig.MaxDelay
		}

		// Add jitter to avoid thundering herd
		jitter := time.Duration(float64(delay) * (1 + backoffConfig.Jitter*(2*rand.Float64()-1)))
		if jitter > backoffConfig.MaxDelay {
			jitter = backoffConfig.MaxDelay
		}

		logger.Warn("Failed to create metrics stream, retrying",
			zap.Error(err),
			zap.Duration("backoff", jitter),
			zap.Int("attempt", retryCount))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(jitter):
			continue
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

type collectionResult struct {
	cpuMetrics     []model.CPUMetrics
	ramMetrics     model.RamMetrics
	diskMetrics    []model.DiskMetrics
	networkMetrics []model.NetworkMetrics
	hostMetrics    model.HostMetrics
	dockerMetrics  []model.DockerMetrics
	errors         []error
}

func collectAndSendMetrics(isServerHealthy bool, c *collectors, stream pb.MetricService_StreamMetricsClient) error {
	if !isServerHealthy {
		logger.Warn("Skipping metrics transmission, server unhealthy")
		return nil
	}

	result := &collectionResult{
		errors: make([]error, 0),
	}

	var err error

	result.ramMetrics, err = c.ram.Collect()
	if err != nil {
		logger.Error("Failed to collect RAM metrics", zap.Error(err))
		result.errors = append(result.errors, fmt.Errorf("failed to collect RAM metrics: %w", err))
	}

	result.hostMetrics, err = c.host.Collect()
	if err != nil {
		logger.Error("Failed to collect host metrics", zap.Error(err))
		result.errors = append(result.errors, fmt.Errorf("failed to collect host metrics: %w", err))
	}

	result.networkMetrics, err = c.network.Collect()
	if err != nil {
		logger.Error("Failed to collect network metrics", zap.Error(err))
		result.errors = append(result.errors, fmt.Errorf("failed to collect network metrics: %w", err))
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	addError := func(err error) {
		if err != nil {
			mu.Lock()
			result.errors = append(result.errors, err)
			mu.Unlock()
		}
	}

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

	if c.docker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dockerMetrics, err := c.docker.Collect()
			if err != nil {
				logger.Debug("Failed to collect Docker metrics", zap.Error(err)) // Reduced to debug level
				return
			}
			mu.Lock()
			result.dockerMetrics = dockerMetrics
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(result.errors) > 0 {
		logger.Warn("Some metrics collection failed", zap.Int("error_count", len(result.errors)))
		// Don't return error, continue with partial metrics
	}

	pbMetrics := &pb.MetricsPayload{
		Host:      converter.ConvertHostMetrics(result.hostMetrics),
		Cpu:       converter.ConvertCPUMetrics(result.cpuMetrics),
		Ram:       converter.ConvertRAMMetrics(result.ramMetrics),
		Disk:      converter.ConvertDiskMetrics(result.diskMetrics),
		Network:   converter.ConvertNetworkMetrics(result.networkMetrics),
		Docker:    converter.ConvertDockerMetrics(result.dockerMetrics),
		Timestamp: timestamppb.Now(),
	}

	if err := stream.Send(pbMetrics); err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive acknowledgment: %w", err)
	}

	logger.Debug("Metrics sent successfully",
		zap.String("status", resp.Status),
		zap.String("message", resp.Message),
		zap.Int("cpu_metrics", len(result.cpuMetrics)),
		zap.Int("disk_metrics", len(result.diskMetrics)),
		zap.Int("network_metrics", len(result.networkMetrics)),
		zap.Int("docker_metrics", len(result.dockerMetrics)))

	return nil
}

// cleanupCollectors properly closes and cleans up all collectors
func cleanupCollectors(c *collectors) {
	if c.docker != nil {
		c.docker.Close()
	}
}
