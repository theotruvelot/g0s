package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	"time"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrMalformedToken = errors.New("malformed token")
)

type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Token struct {
	Token        string
	RefreshToken string
}

type JWTService struct {
	secret        string
	refreshSecret string
}

func NewJWTService(secret string, refreshSecret string) *JWTService {
	return &JWTService{
		secret:        secret,
		refreshSecret: refreshSecret,
	}
}

func (j *JWTService) GenerateJWT(username string) (Token, error) {
	claims := &JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "g0s",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		logger.Error("Error signing JWT", zap.Error(err))
		return Token{}, err
	}
	refreshClaims := &JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "g0s",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSignedToken, err := refreshToken.SignedString([]byte(j.refreshSecret))
	if err != nil {
		logger.Error("Error signing refresh JWT", zap.Error(err))
		return Token{}, err
	}
	return Token{
		Token:        signedToken,
		RefreshToken: refreshSignedToken,
	}, nil
}

func (j *JWTService) CheckJWT(tokenString string, isRefresh bool) (*JWTClaims, error) {
	claims := &JWTClaims{}
	secret := j.secret
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrMalformedToken
		}
		if isRefresh {
			secret = j.refreshSecret
		}
		return []byte(secret), nil
	})

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
