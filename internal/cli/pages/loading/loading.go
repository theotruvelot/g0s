package loading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theotruvelot/g0s/internal/cli/messages"
	"github.com/theotruvelot/g0s/internal/cli/styles"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

// Configuration constants
const (
	healthCheckTimeout = 10 * time.Second
	maxRetries         = 3
	retryDelay         = 2 * time.Second
	progressBarWidth   = 60
)

// Timing constants for each step
var stepTimings = map[LoadingState]time.Duration{
	StateStep1:      800 * time.Millisecond,
	StateStep2:      800 * time.Millisecond,
	StateStep3:      800 * time.Millisecond,
	StateStep4:      600 * time.Millisecond,
	StateStep5:      1 * time.Second,
	StateFinalizing: 800 * time.Millisecond,
}

// Progress percentages for each state
var stateProgress = map[LoadingState]float64{
	StateInitializing: 0.0,
	StateStep1:        0.25,
	StateStep2:        0.50,
	StateStep3:        0.75,
	StateHealthCheck:  0.85,
	StateStep4:        0.90,
	StateStep5:        0.95,
	StateFinalizing:   1.0,
}

// State messages
var stateMessages = map[LoadingState]string{
	StateInitializing: "Initializing application...",
	StateStep1:        "Loading modules...",
	StateStep2:        "Configuring interface...",
	StateStep3:        "Connecting to server...",
	StateHealthCheck:  "Checking connection...",
	StateStep4:        "Verification successful",
	StateStep5:        "Starting application...",
	StateFinalizing:   "Finalizing...",
	StateSuccess:      "Connection successful!",
	StateError:        "Connection failed",
	StateRetrying:     "Retrying connection...",
}

// HealthResponse represents the expected response from the health endpoint
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Version   string            `json:"version,omitempty"`
}

// LoadingState represents the current state of the loading process
type LoadingState int

const (
	StateInitializing LoadingState = iota
	StateStep1                     // 25%
	StateStep2                     // 50%
	StateStep3                     // 75%
	StateHealthCheck               // 85%
	StateStep4                     // 90%
	StateStep5                     // 95%
	StateFinalizing                // 100%
	StateSuccess
	StateError
	StateRetrying
)

// String returns the string representation of the loading state
func (s LoadingState) String() string {
	if msg, exists := stateMessages[s]; exists {
		return msg
	}
	return "Unknown state"
}

// Progress returns the progress percentage for the state
func (s LoadingState) Progress() float64 {
	if progress, exists := stateProgress[s]; exists {
		return progress
	}
	return 0.0
}

// NextState returns the next state in the loading sequence
func (s LoadingState) NextState() LoadingState {
	switch s {
	case StateStep1:
		return StateStep2
	case StateStep2:
		return StateStep3
	case StateStep3:
		return StateHealthCheck
	case StateStep4:
		return StateStep5
	case StateStep5:
		return StateFinalizing
	default:
		return s
	}
}

// Delay returns the delay before transitioning to the next state
func (s LoadingState) Delay() time.Duration {
	if delay, exists := stepTimings[s]; exists {
		return delay
	}
	return 500 * time.Millisecond
}

// Model represents the loading page model
type Model struct {
	// Dependencies
	httpClient *client.Client
	log        *zap.Logger

	// UI Components
	spinner  spinner.Model
	progress progress.Model

	// State
	state      LoadingState
	error      error
	healthData *HealthResponse

	// Retry logic
	retryCount int

	// Dimensions
	width  int
	height int
}

// Custom messages
type (
	stepMsg struct {
		state LoadingState
	}

	healthCheckResultMsg struct {
		success   bool
		status    string
		latency   string
		timestamp string
		error     error
	}
)

// NewModel creates a new loading page model with logger
func NewModel(httpClient *client.Client, log *zap.Logger) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))

	p := progress.New(
		progress.WithScaledGradient(styles.Primary, styles.Warning),
		progress.WithWidth(progressBarWidth),
	)

	return Model{
		httpClient: httpClient,
		log:        log,
		spinner:    s,
		progress:   p,
		state:      StateInitializing,
	}
}

// GetPageType returns the page type
func (m Model) GetPageType() messages.PageType {
	return messages.LoadingPage
}

// OnEnter is called when the page becomes active
func (m Model) OnEnter() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return stepMsg{state: StateStep1}
		}),
	)
}

// OnExit is called when leaving the page
func (m Model) OnExit() tea.Cmd {
	m.log.Debug("Loading page exited")
	return nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.OnEnter()
}

// performHealthCheck creates a command to check server health
func (m Model) performHealthCheck() tea.Cmd {
	return func() tea.Msg {
		m.log.Debug("Starting health check")

		ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
		defer cancel()

		start := time.Now()
		latency := func() string { return time.Since(start).String() }
		timestamp := time.Now().Format(time.RFC3339)

		resp, err := m.httpClient.Get(ctx, "/health")
		if err != nil {
			m.log.Error("Health check request failed", zap.Error(err))
			return healthCheckResultMsg{
				success:   false,
				status:    "error",
				error:     err,
				latency:   latency(),
				timestamp: timestamp,
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("server returned status %d", resp.StatusCode)
			m.log.Error("Health check failed", zap.Int("status", resp.StatusCode))
			return healthCheckResultMsg{
				success:   false,
				status:    fmt.Sprintf("HTTP %d", resp.StatusCode),
				error:     err,
				latency:   latency(),
				timestamp: timestamp,
			}
		}

		var healthResp HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
			m.log.Error("Failed to parse health response", zap.Error(err))
			return healthCheckResultMsg{
				success:   false,
				status:    "parse_error",
				error:     fmt.Errorf("failed to parse response: %w", err),
				latency:   latency(),
				timestamp: timestamp,
			}
		}

		return healthCheckResultMsg{
			success:   true,
			status:    healthResp.Status,
			latency:   latency(),
			timestamp: timestamp,
		}
	}
}

// handleStepMessage handles step progression messages
func (m *Model) handleStepMessage(msg stepMsg) []tea.Cmd {
	var cmds []tea.Cmd

	m.state = msg.state

	// Update progress bar
	if cmd := m.progress.SetPercent(m.state.Progress()); cmd != nil {
		cmds = append(cmds, cmd)
	}
	cmds = append(cmds, m.spinner.Tick)

	// Handle special states
	switch m.state {
	case StateHealthCheck:
		cmds = append(cmds, m.performHealthCheck())
	default:
		// Schedule next step
		nextState := m.state.NextState()
		if nextState != m.state {
			delay := m.state.Delay()
			cmds = append(cmds, tea.Tick(delay, func(t time.Time) tea.Msg {
				return stepMsg{state: nextState}
			}))
		}
	}

	return cmds
}

// handleHealthCheckResult handles health check results
func (m *Model) handleHealthCheckResult(msg healthCheckResultMsg) []tea.Cmd {
	var cmds []tea.Cmd

	if msg.success {
		m.healthData = &HealthResponse{
			Status:    msg.status,
			Timestamp: msg.timestamp,
		}
		// Continue to next step
		cmds = append(cmds, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
			return stepMsg{state: StateStep4}
		}))
	} else {
		m.log.Error("Health check failed", zap.Error(msg.error))
		m.error = msg.error

		if m.retryCount < maxRetries {
			m.state = StateRetrying
			m.retryCount++
			m.log.Info("Scheduling retry",
				zap.Int("attempt", m.retryCount),
				zap.Int("maxRetries", maxRetries))

			cmds = append(cmds,
				m.progress.SetPercent(StateHealthCheck.Progress()),
				tea.Tick(retryDelay, func(time.Time) tea.Msg {
					return stepMsg{state: StateHealthCheck}
				}),
				m.spinner.Tick,
			)
		} else {
			m.state = StateError
			m.log.Error("Max retries exceeded")
		}
	}

	return cmds
}

// resetForRetry resets the model state for a retry
func (m *Model) resetForRetry() []tea.Cmd {
	m.state = StateInitializing
	m.retryCount = 0
	return []tea.Cmd{
		m.progress.SetPercent(0),
		m.spinner.Tick,
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return stepMsg{state: StateStep1}
		}),
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = min(m.width-20, progressBarWidth)
		m.log.Debug("Window size updated",
			zap.Int("width", m.width),
			zap.Int("height", m.height))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == StateError {
				return m, tea.Quit
			}
		case "r":
			if m.state == StateError {
				m.log.Info("User requested retry")
				cmds = append(cmds, m.resetForRetry()...)
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
		cmds = append(cmds, m.handleStepMessage(msg)...)

	case healthCheckResultMsg:
		cmds = append(cmds, m.handleHealthCheckResult(msg)...)
	}

	return m, tea.Batch(cmds...)
}

// renderLogo returns the styled application logo
func (m Model) renderLogo() string {
	return lipgloss.NewStyle().
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
`)
}

// renderTitle returns the styled application title
func (m Model) renderTitle() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.Text)).
		Bold(true).
		Render("g0s System Monitor")
}

// renderLoadingLine returns the loading line with spinner and text
func (m Model) renderLoadingLine() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.spinner.View()+" ",
		lipgloss.NewStyle().Bold(true).Render(m.state.String()),
	)
}

// renderStatusMessage returns the status message based on current state
func (m Model) renderStatusMessage() string {
	switch m.state {
	case StateError:
		content := styles.ErrorStyle.Render("❌ " + m.state.String())
		if m.error != nil {
			content += "\n" + styles.ErrorStyle.Render(m.error.Error())
		}
		content += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextMuted)).
			Italic(true).
			Render("Press 'r' to retry or 'q' to quit")
		return content

	case StateRetrying:
		return styles.WarningStyle.Render(fmt.Sprintf("⚠️  Attempt %d/%d", m.retryCount, maxRetries))

	default:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextMuted)).
			Italic(true).
			Render("Please wait...")
	}
}

// View renders the loading page
func (m Model) View() string {
	if m.width == 0 {
		return "\n  Initializing..."
	}

	var content strings.Builder

	// Build content
	content.WriteString(m.renderLogo())
	content.WriteString("\n\n")
	content.WriteString(m.renderTitle())
	content.WriteString("\n\n")
	content.WriteString(m.renderLoadingLine())
	content.WriteString("\n\n")
	content.WriteString(m.progress.View())
	content.WriteString("\n\n")
	content.WriteString(m.renderStatusMessage())

	// Center everything in the terminal
	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(m.width).
		Height(m.height)

	return containerStyle.Render(content.String())
}
