package services

import (
	"context"
	"github.com/theotruvelot/g0s/internal/cli/clients"
	"github.com/theotruvelot/g0s/pkg/proto/health"
)

type HealthCheckService struct {
	Clients *clients.Clients
}

func NewHealthCheckService(clients *clients.Clients) *HealthCheckService {
	return &HealthCheckService{
		Clients: clients,
	}
}

func (h *HealthCheckService) Check(ctx context.Context) (*health.HealthCheckResponse, error) {
	req := &health.HealthCheckRequest{}
	res, err := h.Clients.HealthcheckClient.Check(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
