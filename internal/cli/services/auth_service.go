package services

import (
	"context"
	"fmt"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"

	"github.com/theotruvelot/g0s/internal/cli/clients"
	"github.com/theotruvelot/g0s/pkg/proto/auth"
)

type AuthService struct {
	Clients *clients.Clients
}

func NewAuthService(clients *clients.Clients) *AuthService {
	return &AuthService{
		Clients: clients,
	}
}

func (a *AuthService) Login(ctx context.Context, serverURL, username, token string) (*auth.AuthenticateResponse, error) {
	tempClients, err := clients.NewClients(serverURL)
	if err != nil {
		return nil, fmt.Errorf("could not create temporary gRPC client: %w", err)
	}
	defer func(tempClients *clients.Clients) {
		err := tempClients.Close()
		if err != nil {
			logger.Error("Failed to close temporary gRPC client", zap.Error(err))
		} else {
			logger.Debug("Temporary gRPC client closed successfully")
		}
	}(tempClients)

	req := &auth.AuthenticateRequest{Username: username, Token: token}
	res, err := tempClients.AuthClient.Authenticate(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//TODO REFRESH TOKEN
//func (a *AuthService) RefreshToken(ctx context.Context, refresh string) (*auth.RefreshTokenResponse, error) {
//	req := &auth.RefreshTokenRequest{JwtRefreshToken: refresh}
//	res, err := a.Clients.AuthClient.RefreshToken(ctx, req)
//	if err != nil {
//		return nil, fmt.Errorf("could not refresh token: %w", err)
//	}
//	if res.Status != auth.RefreshTokenResponse_OK {
//		return nil, fmt.Errorf("failed to refresh token: %s", res.Status.String())
//	}
//
//	return res, nil
//}
