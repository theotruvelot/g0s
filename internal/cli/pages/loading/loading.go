package loading

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theotruvelot/g0s/internal/cli/clients"
	"github.com/theotruvelot/g0s/internal/cli/styles"
	"github.com/theotruvelot/g0s/pkg/logger"
	"github.com/theotruvelot/g0s/pkg/proto/health"
	"go.uber.org/zap"
)

// Constants
const (
	maxRetries       = 3
	retryDelay       = 2 * time.Second
	progressBarWidth = 60
)

// LoadingState represents the current state
type LoadingState int

const (
	StateConnecting LoadingState = iota
	StateHealthCheck
	StateSuccess
	StateError
	StateRetrying
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Success   bool
	Status    health.HealthCheckResponse_ServingStatus
	Latency   string
	Error     error
	Timestamp string
}

// Model represents the loading page
type Model struct {
	grpcClients *clients.Clients

	spinner  spinner.Model
	progress progress.Model

	state      LoadingState
	error      error
	healthData *HealthCheckResult
	retryCount int

	width  int
	height int
}

// stepMsg represents a state transition
type stepMsg struct {
	state LoadingState
}

// NewModel creates a new loading model
func NewModel(grpcClients *clients.Clients) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))

	p := progress.New(
		progress.WithScaledGradient(styles.Primary, styles.Warning),
		progress.WithWidth(progressBarWidth),
	)

	return Model{
		grpcClients: grpcClients,
		spinner:     s,
		progress:    p,
		state:       StateConnecting,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return stepMsg{state: StateHealthCheck}
		}),
	)
}

// performHealthCheck performs the health check
func (m Model) performHealthCheck() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("Performing health check")

		start := time.Now()
		timestamp := time.Now().Format(time.RFC3339)

		// Perform the health check
		res, err := m.grpcClients.HealthcheckClient.Check(context.Background(), &health.HealthCheckRequest{})
		if err != nil {
			logger.Error("Health check failed", zap.Error(err))
			return HealthCheckResult{
				Success:   false,
				Status:    res.Status,
				Error:     err,
				Latency:   time.Since(start).String(),
				Timestamp: timestamp,
			}
		}

		return HealthCheckResult{
			Success:   true,
			Status:    res.Status,
			Latency:   time.Since(start).String(),
			Timestamp: timestamp,
		}
	}

}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		newWidth := msg.Width - 20
		if newWidth < 20 {
			newWidth = 20
		}
		if newWidth > progressBarWidth {
			newWidth = progressBarWidth
		}
		m.progress.Width = newWidth

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if m.state == StateError {
				logger.Info("Retrying connection")
				m.state = StateConnecting
				m.error = nil
				m.retryCount = 0
				cmds = append(cmds, m.progress.SetPercent(0.3))
				cmds = append(cmds, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
					return stepMsg{state: StateHealthCheck}
				}))
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case stepMsg:
		m.state = msg.state
		switch m.state {
		case StateConnecting:
			cmds = append(cmds, m.progress.SetPercent(0.3))
		case StateHealthCheck:
			cmds = append(cmds, m.progress.SetPercent(0.7))
			cmds = append(cmds, m.performHealthCheck())
		case StateSuccess:
			cmds = append(cmds, m.progress.SetPercent(1.0))
		case StateError:
			// Error state - no additional commands needed
		case StateRetrying:
			// Retrying state - no additional commands needed
		}

	case HealthCheckResult:
		if msg.Success {
			m.healthData = &msg
			m.error = nil
			m.state = StateSuccess
			cmds = append(cmds, m.progress.SetPercent(1.0))
		} else {
			m.error = msg.Error
			if m.retryCount < maxRetries {
				m.state = StateRetrying
				m.retryCount++
				cmds = append(cmds, tea.Tick(retryDelay, func(time.Time) tea.Msg {
					return stepMsg{state: StateHealthCheck}
				}))
			} else {
				m.state = StateError
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the loading page
func (m Model) View() string {
	var content strings.Builder

	// Logo
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.Primary)).
		Bold(true).
		Render(`
 _______  _______  _______ 
|       ||  _    ||       |
|    ___|| | |   ||  _____|
|   | __ | | |   || |_____ 
|   ||  || |_|   ||_____  |
|   |_| ||       | _____| |
|_______||_______||_______|
`))

	content.WriteString("\n\n")

	// Title
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.Text)).
		Bold(true).
		Render("g0s System Monitor"))

	content.WriteString("\n\n")

	// Status line
	var statusText string
	switch m.state {
	case StateConnecting:
		statusText = "Connecting to server..."
	case StateHealthCheck:
		statusText = "Checking server health..."
	case StateSuccess:
		statusText = "✅ Connected successfully!"
	case StateError:
		statusText = "❌ Connection failed"
	case StateRetrying:
		statusText = fmt.Sprintf("⚠️  Retrying... (attempt %d/%d)", m.retryCount, maxRetries)
	}

	content.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.spinner.View()+" ",
		lipgloss.NewStyle().Bold(true).Render(statusText),
	))

	content.WriteString("\n\n")

	// Progress bar - always center it
	progressView := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(m.progress.View())
	content.WriteString(progressView)
	content.WriteString("\n\n")

	// Error message or instructions
	if m.state == StateError {
		content.WriteString(styles.ErrorStyle.Render("Error: " + m.error.Error()))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextMuted)).
			Italic(true).
			Render("Press 'r' to retry or 'q' to quit"))
	} else if m.state == StateSuccess {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextMuted)).
			Italic(true).
			Render("Press any key to continue..."))
	} else {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextMuted)).
			Italic(true).
			Render("Please wait..."))
	}

	// Always center content and use full terminal dimensions
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(m.width).
		Height(m.height).
		Render(content.String())
}
