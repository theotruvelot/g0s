package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

const (
	_defaultTimeout = 10 * time.Second
)

type HTTPError struct {
	URL    string
	Method string
	Err    error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP request failed [%s %s]: %v", e.Method, e.URL, e.Err)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	log        *zap.Logger
}

func NewClient(baseURL string, token string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = _defaultTimeout
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		token:   token,
		log:     logger.GetLogger(),
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	if ctx == nil {
		return nil, &HTTPError{
			Method: method,
			URL:    path,
			Err:    fmt.Errorf("context cannot be nil"),
		}
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)

	c.log.Debug("Creating HTTP request",
		zap.String("url", url),
		zap.String("method", method))

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.log.Error("Failed to create HTTP request",
			zap.String("url", url),
			zap.String("method", method),
			zap.Error(err))
		return nil, &HTTPError{
			Method: method,
			URL:    url,
			Err:    fmt.Errorf("error creating request: %w", err),
		}
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.log.Debug("Sending HTTP request",
		zap.String("url", url),
		zap.String("method", method))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("HTTP request failed",
			zap.String("url", url),
			zap.String("method", method),
			zap.Error(err))
		return nil, &HTTPError{
			Method: method,
			URL:    url,
			Err:    err,
		}
	}

	c.log.Debug("HTTP request completed",
		zap.String("url", url),
		zap.String("method", method),
		zap.Int("status", resp.StatusCode))

	return resp, nil
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, body)
}

func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil)
}
