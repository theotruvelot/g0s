package service

import (
	"context"
	"sync"
	"time"

	"github.com/theotruvelot/g0s/pkg/logger"
	health "github.com/theotruvelot/g0s/pkg/proto/health"
	"go.uber.org/zap"
)

type ClientInfo struct {
	ID          string
	Hostname    string
	IPAddress   string
	ConnectedAt time.Time
}

type HealthCheckService struct {
	clients     map[string]ClientInfo
	clientsLock sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewHealthCheckService() *HealthCheckService {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthCheckService{
		clients: make(map[string]ClientInfo),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *HealthCheckService) Shutdown() {
	s.cancel()
}

func (s *HealthCheckService) NotifyShutdown() {
	logger.Info("Notifying health check clients about server shutdown")
	s.cancel()
}

func (s *HealthCheckService) RegisterClient(id, hostname, ip string) {
	s.clientsLock.Lock()
	s.clients[id] = ClientInfo{
		ID:          id,
		Hostname:    hostname,
		IPAddress:   ip,
		ConnectedAt: time.Now(),
	}
	s.clientsLock.Unlock()
	logger.Debug("Client connected",
		zap.String("client_id", id),
		zap.String("hostname", hostname),
		zap.String("ip_address", ip))
}

func (s *HealthCheckService) UnregisterClient(id string) {
	s.clientsLock.Lock()
	delete(s.clients, id)
	s.clientsLock.Unlock()
	logger.Debug("Client disconnected", zap.String("client_id", id))
}

func (s *HealthCheckService) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthCheckService) Watch(
	ctx context.Context,
	clientID, hostname, ip string,
	sendStatus func(status health.HealthCheckResponse_ServingStatus) error,
) error {
	// Register client
	s.RegisterClient(clientID, hostname, ip)
	defer func() {
		s.UnregisterClient(clientID)
		logger.Info("Health watch stream terminated", zap.String("client_id", clientID))
	}()

	// Send initial health status
	if err := sendStatus(health.HealthCheckResponse_SERVING); err != nil {
		logger.Error("Error sending initial health status", zap.Error(err))
		return err
	}

	// Wait for context cancellation or server shutdown
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		logger.Info("Server is shutting down, notifying client")
		_ = sendStatus(health.HealthCheckResponse_NOT_SERVING)
		return nil
	}
}
