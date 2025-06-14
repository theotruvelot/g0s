package metrics

import (
	"bytes"
	"fmt"
	"github.com/theotruvelot/g0s/pkg/logger"
	"net"
	"net/http"
	"sync"
	"time"

	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
)

type MetricStore interface {
	Format(metrics *pb.MetricsPayload, timestamp int64) []string
	Store(data []string) error
}

type Manager struct {
	stores []MetricStore
}

func NewMetricsManager(vmEndpoint string) *Manager {
	return &Manager{
		stores: []MetricStore{
			NewCPUStore(vmEndpoint),
			NewRAMStore(vmEndpoint),
			NewDiskStore(vmEndpoint),
			NewNetworkStore(vmEndpoint),
			NewDockerStore(vmEndpoint),
		},
	}
}

func (m *Manager) StoreAllMetrics(metrics *pb.MetricsPayload) error {
	timestamp := metrics.Timestamp.AsTime().UnixNano() / int64(time.Millisecond)
	var wg sync.WaitGroup
	errors := make(chan error, len(m.stores))

	// Make it parallel
	for _, store := range m.stores {
		wg.Add(1)
		go func(s MetricStore) {
			defer wg.Done()

			lines := s.Format(metrics, timestamp)
			if err := s.Store(lines); err != nil {
				errors <- err
			}
		}(store)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		logger.Error("Failed to store metrics", zap.Error(err))
	}

	return nil
}

// sendWithRetry envoie les données avec retry automatique
func sendWithRetry(endpoint, payload string, metricType string) error {
	const maxRetries = 3
	const baseDelay = 500 * time.Millisecond
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		},
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * baseDelay
			logger.Debug("Retrying request",
				zap.String("metric_type", metricType),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		resp, err := client.Post(endpoint, "text/plain", bytes.NewBufferString(payload))
		if err != nil {
			lastErr = err
			logger.Warn("HTTP request failed",
				zap.String("metric_type", metricType),
				zap.Int("attempt", attempt+1),
				zap.Error(err))
			continue
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			if attempt > 0 {
				logger.Info("Request succeeded after retry",
					zap.String("metric_type", metricType),
					zap.Int("attempts", attempt+1))
			}
			return nil
		}

		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		logger.Warn("Unexpected status code",
			zap.String("metric_type", metricType),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", attempt+1))
	}

	return fmt.Errorf("failed to send %s metrics after %d attempts: %w", metricType, maxRetries, lastErr)
}
