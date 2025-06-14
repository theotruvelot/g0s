package service

import (
	"errors"
	"github.com/theotruvelot/g0s/internal/server/auth"
	"github.com/theotruvelot/g0s/internal/server/storage/database"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	UserRepo   *database.UserRepository
	JWTService *auth.JWTService
}

func NewAuthService(userRepo database.UserRepository, jwtService auth.JWTService) *AuthService {
	return &AuthService{
		UserRepo:   &userRepo,
		JWTService: &jwtService,
	}
}

func (a *AuthService) Authenticate(username, token string) (auth.Token, error) {
	user, err := a.UserRepo.GetUserByUsername(username)
	if err != nil {
		logger.Error("Error Authentication", zap.Error(err))
		return auth.Token{}, err
	}

	if user == nil || user.Token != token {
		logger.Info("Invalid credentials", zap.String("username", username))
		return auth.Token{}, ErrInvalidCredentials
	}

	return a.JWTService.GenerateJWT(user.Username)
}
