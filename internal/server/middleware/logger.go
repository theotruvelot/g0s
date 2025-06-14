package middleware

import (
	"context"
	"time"

	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor logs unary gRPC requests and responses
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Get peer info
		peerInfo := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			peerInfo = p.Addr.String()
		}

		// Log incoming request
		logger.Info("gRPC unary request started",
			zap.String("method", info.FullMethod),
			zap.String("peer", peerInfo),
		)

		// Execute the handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Log response
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("peer", peerInfo),
			zap.Duration("duration", duration),
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				zap.String("status", st.Code().String()),
				zap.Error(err),
			)
			logger.Error("gRPC unary request failed", fields...)
		} else {
			fields = append(fields, zap.String("status", "OK"))
			logger.Info("gRPC unary request completed", fields...)
		}

		return resp, err
	}
}

// LoggingStreamInterceptor logs streaming gRPC requests
func LoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Get peer info
		peerInfo := "unknown"
		if p, ok := peer.FromContext(stream.Context()); ok {
			peerInfo = p.Addr.String()
		}

		// Log stream start
		logger.Info("gRPC stream started",
			zap.String("method", info.FullMethod),
			zap.String("peer", peerInfo),
			zap.Bool("client_stream", info.IsClientStream),
			zap.Bool("server_stream", info.IsServerStream),
		)

		// Execute the handler
		err := handler(srv, stream)

		duration := time.Since(start)

		// Log stream end
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("peer", peerInfo),
			zap.Duration("duration", duration),
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				zap.String("status", st.Code().String()),
				zap.Error(err),
			)
			logger.Error("gRPC stream failed", fields...)
		} else {
			fields = append(fields, zap.String("status", "OK"))
			logger.Info("gRPC stream completed", fields...)
		}

		return err
	}
}
