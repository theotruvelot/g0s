package healthcheck

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/theotruvelot/g0s/pkg/client"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

const (
	_defaultTimeout = 5 * time.Second
	_healthEndpoint = "/health"
)

type Service struct {
	client    *client.Client
	interval  time.Duration
	lastCheck bool
}

type Config struct {
	ServerURL string
	Token     string
	Interval  time.Duration
}

func New(cfg Config) *Service {
	return &Service{
		client:    client.NewClient(cfg.ServerURL, cfg.Token, _defaultTimeout),
		interval:  cfg.Interval,
		lastCheck: false,
	}
}

func (s *Service) Start(ctx context.Context) {
	logger.Info("Starting health check service", zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.check(ctx)

	for {
		select {
		case <-ticker.C:
			s.check(ctx)
		case <-ctx.Done():
			logger.Info("Health check service stopped")
			return
		}
	}
}

func (s *Service) IsHealthy() bool {
	return s.lastCheck
}

func (s *Service) check(ctx context.Context) {
	checkCtx, cancel := context.WithTimeout(ctx, _defaultTimeout)
	defer cancel()

	resp, err := s.client.Get(checkCtx, _healthEndpoint)
	if err != nil {
		logger.Error("Health check failed", zap.Error(err))
		s.lastCheck = false
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	s.lastCheck = resp.StatusCode == http.StatusOK

	if !s.lastCheck {
		logger.Warn("Server health check failed", zap.Int("status_code", resp.StatusCode))
	}
}
