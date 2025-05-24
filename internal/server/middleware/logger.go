package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	maxBodyLogSize = 1024 * 4 // 4KB max body logging
)

// RequestLogger returns a middleware that logs HTTP requests and responses
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())

			// Log incoming request
			logRequest(r, requestID)

			// Capture request body
			bodyBytes := captureRequestBody(r)

			// Create response writer wrapper
			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     200,
				body:           &bytes.Buffer{},
			}

			// Process request
			next.ServeHTTP(ww, r)

			// Log response
			logResponse(r, ww, time.Since(start), requestID, bodyBytes)
		})
	}
}

// logRequest logs the incoming HTTP request
func logRequest(r *http.Request, requestID string) {
	fields := []zap.Field{
		zap.String("type", "request"),
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}

	// Add query parameters if present
	if r.URL.RawQuery != "" {
		fields = append(fields, zap.String("query", r.URL.RawQuery))
	}

	// Add content type if present
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		fields = append(fields, zap.String("content_type", contentType))
	}

	// Add content length if present
	if r.ContentLength > 0 {
		fields = append(fields, zap.Int64("content_length", r.ContentLength))
	}

	logger.Info("HTTP request started", fields...)
}

// logResponse logs the HTTP response
func logResponse(r *http.Request, ww *responseWriter, duration time.Duration, requestID string, requestBody []byte) {
	fields := []zap.Field{
		zap.String("type", "response"),
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status_code", ww.statusCode),
		zap.Duration("duration", duration),
		zap.Int("response_size", ww.body.Len()),
	}

	// Add request body if present and not too large
	if len(requestBody) > 0 {
		bodyStr := prepareBodyForLogging(requestBody)
		if bodyStr != "" {
			fields = append(fields, zap.String("request_body", bodyStr))
		}
	}

	// Add response body for small responses
	if ww.body.Len() > 0 && ww.body.Len() <= maxBodyLogSize {
		fields = append(fields, zap.String("response_body", ww.body.String()))
	}

	// Choose log level based on status code
	logLevel := getLogLevelForStatus(ww.statusCode)
	switch logLevel {
	case zap.InfoLevel:
		logger.Info("HTTP request completed", fields...)
	case zap.WarnLevel:
		logger.Warn("HTTP request completed with warning", fields...)
	case zap.ErrorLevel:
		logger.Error("HTTP request completed with error", fields...)
	}
}

// captureRequestBody safely captures the request body and restores it
func captureRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Failed to read request body", zap.Error(err))
		return nil
	}

	// Close the original body
	if err := r.Body.Close(); err != nil {
		logger.Warn("Failed to close request body", zap.Error(err))
	}

	// Restore the body for the next handler
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes
}

// prepareBodyForLogging prepares body bytes for logging with size limits
func prepareBodyForLogging(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}

	if len(bodyBytes) > maxBodyLogSize {
		return string(bodyBytes[:maxBodyLogSize]) + "... (truncated)"
	}

	return string(bodyBytes)
}

// getLogLevelForStatus returns appropriate log level based on HTTP status code
func getLogLevelForStatus(statusCode int) zapcore.Level {
	switch {
	case statusCode >= 500:
		return zap.ErrorLevel
	case statusCode >= 400:
		return zap.WarnLevel
	default:
		return zap.InfoLevel
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body for logging
func (rw *responseWriter) Write(data []byte) (int, error) {
	// Capture response body for logging (with size limit)
	if rw.body.Len()+len(data) <= maxBodyLogSize {
		rw.body.Write(data)
	}
	return rw.ResponseWriter.Write(data)
}
