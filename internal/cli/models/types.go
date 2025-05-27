package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/messages"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

// Re-export commonly used types from messages package
type PageType = messages.PageType
type NavigateMsg = messages.NavigateMsg
type ErrorMsg = messages.ErrorMsg
type HealthCheckMsg = messages.HealthCheckMsg
type HealthCheckResult = messages.HealthCheckResult

// Re-export page type constants
const (
	LoadingPage = messages.LoadingPage
	ErrorPage   = messages.ErrorPage
)

// Common interface that all page models should implement
type PageModel interface {
	tea.Model
	GetPageType() PageType
	OnEnter() tea.Cmd // Called when the page becomes active
	OnExit() tea.Cmd  // Called when leaving the page
}

// AppConfig holds configuration passed to models
type AppConfig struct {
	HTTPClient *client.Client
	ServerURL  string
	APIToken   string
	Logger     *zap.Logger
}
