package grpc

import (
	"context"
	"errors"
	"github.com/theotruvelot/g0s/pkg/logger"

	"github.com/theotruvelot/g0s/internal/server/service"
	pb "github.com/theotruvelot/g0s/pkg/proto/auth"
	"google.golang.org/grpc"
)

type AuthHandler struct {
	AuthService *service.AuthService
	pb.UnimplementedAuthServiceServer
	ctx    context.Context
	cancel context.CancelFunc
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &AuthHandler{
		AuthService: authService,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (h *AuthHandler) RegisterServices(server *grpc.Server) {
	pb.RegisterAuthServiceServer(server, h)
	logger.Debug("Auth gRPC service registered")
}

func (h *AuthHandler) Shutdown() {
	h.cancel()
}

func (h *AuthHandler) NotifyShutdown() {
	h.cancel()
}

func (h *AuthHandler) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	token, err := h.AuthService.Authenticate(req.Username, req.Token)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return &pb.AuthenticateResponse{
				Status: pb.AuthenticateResponse_INVALID_CREDENTIALS,
			}, nil
		}
		return &pb.AuthenticateResponse{
			Status: pb.AuthenticateResponse_ERROR,
		}, err
	}

	return &pb.AuthenticateResponse{
		Status:          pb.AuthenticateResponse_OK,
		JwtToken:        token.Token,
		JwtRefreshToken: token.RefreshToken,
	}, nil
}
