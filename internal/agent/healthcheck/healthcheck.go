package healthcheck

import (
	"context"
	"sync/atomic"
	"time"

	health "github.com/theotruvelot/g0s/pkg/proto/health"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service struct {
	healthy  atomic.Bool
	hostname string
	client   health.HealthServiceClient
	logger   *zap.Logger
}

func New(conn *grpc.ClientConn, logger *zap.Logger, hostname string) *Service {
	return &Service{
		hostname: hostname,
		client:   health.NewHealthServiceClient(conn),
		logger:   logger,
	}
}

func (s *Service) Start(ctx context.Context, interval time.Duration) error {
	s.logger.Info("Starting health check service", zap.Duration("interval", interval))

	go s.healthCheckLoop(ctx, interval)

	return nil
}

// Optimized health check loop with exponential backoff
func (s *Service) healthCheckLoop(ctx context.Context, interval time.Duration) {
	backoffDelay := 2 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.watchHealth(ctx) {
				// Success: reset backoff
				backoffDelay = 2 * time.Second
				// Sleep for the configured interval before next check
				select {
				case <-ctx.Done():
					return
				case <-time.After(interval):
				}
			} else {
				// Failed: apply exponential backoff
				s.setHealthy(false)
				s.logger.Debug("Health check failed, backing off",
					zap.Duration("backoff", backoffDelay))

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoffDelay):
				}

				// Exponential backoff with jitter
				backoffDelay *= 2
				if backoffDelay > maxBackoff {
					backoffDelay = maxBackoff
				}
			}
		}
	}
}

// watchHealth returns true if stream was successful, false if it should retry
func (s *Service) watchHealth(ctx context.Context) bool {

	stream, err := s.client.Watch(ctx, &health.HealthCheckRequest{
		Hostname: s.hostname,
	})
	if err != nil {
		s.logger.Debug("Failed to start health check stream", zap.Error(err))
		return false
	}

	s.setHealthy(true)
	s.logger.Info("Health stream opened")

	defer func() {
		s.setHealthy(false)
		s.logger.Info("Health stream closed")
	}()

	for {
		// Keep the stream alive by blocking on Recv()
		_, err := stream.Recv()
		if err != nil {
			s.logger.Info("Stream closed by server", zap.Error(err))
			return false
		}

		// Optional: could log heartbeat received if server sends periodic messages
	}
}

func (s *Service) IsHealthy() bool {
	return s.healthy.Load()
}

func (s *Service) setHealthy(healthy bool) {
	s.healthy.Store(healthy)
}
