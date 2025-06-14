package middleware

import (
	"context"
	pbhealth "github.com/theotruvelot/g0s/pkg/proto/health"
	pbmetric "github.com/theotruvelot/g0s/pkg/proto/metric"
	"strings"

	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthType defines the type of authentication required
type AuthType int

const (
	// NoAuth means no authentication required
	NoAuth AuthType = iota
	// JWTAuth means JWT authentication required (for CLI)
	JWTAuth
	// MTLSAuth means mTLS authentication required (for agents)
	MTLSAuth
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWTSecret will hold the JWT secret for validation (future use)
	JWTSecret string
	// RequiredMethods maps gRPC method names to required auth types
	RequiredMethods map[string]AuthType
}

// AuthUnaryInterceptor returns a unary interceptor for authentication
func AuthUnaryInterceptor(config AuthConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if this method requires authentication
		authType, exists := config.RequiredMethods[info.FullMethod]
		if !exists {
			authType = NoAuth // Default to no auth if not specified
		}

		logger.Debug("Checking authentication",
			zap.String("method", info.FullMethod),
			zap.Int("auth_type", int(authType)),
		)

		// Perform authentication based on type
		if err := authenticateRequest(ctx, authType, config); err != nil {
			logger.Warn("Authentication failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, err
		}

		// If auth passed or not required, continue with the request
		return handler(ctx, req)
	}
}

// AuthStreamInterceptor returns a stream interceptor for authentication
func AuthStreamInterceptor(config AuthConfig) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Check if this method requires authentication
		authType, exists := config.RequiredMethods[info.FullMethod]
		if !exists {
			authType = NoAuth // Default to no auth if not specified
		}

		logger.Debug("Checking authentication for stream",
			zap.String("method", info.FullMethod),
			zap.Int("auth_type", int(authType)),
		)

		// Perform authentication based on type
		if err := authenticateRequest(stream.Context(), authType, config); err != nil {
			logger.Warn("Stream authentication failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return err
		}

		// If auth passed or not required, continue with the stream
		return handler(srv, stream)
	}
}

// authenticateRequest performs the actual authentication logic
func authenticateRequest(ctx context.Context, authType AuthType, config AuthConfig) error {
	switch authType {
	case NoAuth:
		// No authentication required
		return nil

	case JWTAuth:
		// TODO: Implement JWT authentication for CLI
		return authenticateJWT(ctx, config)

	case MTLSAuth:
		// TODO: Implement mTLS authentication for agents
		return authenticateMTLS(ctx, config)

	default:
		logger.Error("Unknown authentication type", zap.Int("auth_type", int(authType)))
		return status.Error(codes.Internal, "unknown authentication type")
	}
}

// authenticateJWT validates JWT tokens (placeholder for future implementation)
func authenticateJWT(ctx context.Context, config AuthConfig) error {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Look for authorization header
	auth := md.Get("authorization")
	if len(auth) == 0 {
		return status.Error(codes.Unauthenticated, "missing authorization header")
	}

	// Check if it's a Bearer token
	token := auth[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	// TODO: Implement actual JWT validation
	logger.Debug("JWT authentication placeholder - would validate token here",
		zap.String("token_prefix", token[:20]+"..."),
	)

	// For now, just log and pass through
	// In the future, validate the JWT token here
	return nil
}

// authenticateMTLS validates mTLS certificates (placeholder for future implementation)
func authenticateMTLS(ctx context.Context, config AuthConfig) error {
	// TODO: Implement mTLS certificate validation
	// This would typically involve:
	// 1. Extracting client certificate from TLS connection
	// 2. Validating certificate chain
	// 3. Checking certificate against allowed CAs/certificates

	logger.Debug("mTLS authentication placeholder - would validate certificates here")

	// For now, just log and pass through
	// In the future, validate the client certificate here
	return nil
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		RequiredMethods: map[string]AuthType{
			// Health check methods don't require auth
			pbhealth.HealthService_Check_FullMethodName: NoAuth,
			pbhealth.HealthService_Watch_FullMethodName: NoAuth,

			// For now, metric methods don't require auth either
			// We'll update this later when implementing actual auth
			pbmetric.MetricService_StreamMetrics_FullMethodName:    NoAuth,
			pbmetric.MetricService_GetMetrics_FullMethodName:       NoAuth,
			pbmetric.MetricService_GetMetricsStream_FullMethodName: NoAuth,
		},
	}
}
