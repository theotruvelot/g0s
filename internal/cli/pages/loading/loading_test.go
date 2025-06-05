package loading

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

func TestNewModel(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		apiToken  string
	}{
		{
			name:      "valid parameters",
			serverURL: "http://localhost:8080",
			apiToken:  "test-token",
		},
		{
			name:      "https URL",
			serverURL: "https://api.example.com",
			apiToken:  "test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient(tt.serverURL, tt.apiToken, 30*time.Second)

			model := NewModel(httpClient, logger)

			// Check initial state
			if model.state != StateConnecting {
				t.Errorf("NewModel() initial state = %v, want %v", model.state, StateConnecting)
			}

			if model.httpClient == nil {
				t.Errorf("NewModel() httpClient is nil")
			}

			if model.log == nil {
				t.Errorf("NewModel() logger is nil")
			}

			if model.retryCount != 0 {
				t.Errorf("NewModel() retryCount = %d, want 0", model.retryCount)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	logger := zap.NewNop()
	httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
	model := NewModel(httpClient, logger)

	cmd := model.Init()
	if cmd == nil {
		t.Errorf("Init() returned nil command, expected non-nil")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	logger := zap.NewNop()
	httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
	model := NewModel(httpClient, logger)

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(Model)
	if m.width != 80 {
		t.Errorf("Update() width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("Update() height = %d, want 24", m.height)
	}
}

func TestModel_Update_KeyMessages(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		initialState LoadingState
		expectQuit   bool
	}{
		{
			name:         "ctrl+c should quit",
			key:          "ctrl+c",
			initialState: StateConnecting,
			expectQuit:   true,
		},
		{
			name:         "q should quit",
			key:          "q",
			initialState: StateConnecting,
			expectQuit:   true,
		},
		{
			name:         "r should retry in error state",
			key:          "r",
			initialState: StateError,
			expectQuit:   false,
		},
		{
			name:         "r should not affect non-error state",
			key:          "r",
			initialState: StateConnecting,
			expectQuit:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
			model := NewModel(httpClient, logger)
			model.state = tt.initialState
			if tt.initialState == StateError {
				model.error = errors.New("test error")
			}

			// Create key message
			var keyMsg tea.KeyMsg
			switch tt.key {
			case "ctrl+c":
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			case "q":
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
			case "r":
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
			}

			updatedModel, cmd := model.Update(keyMsg)

			if tt.expectQuit {
				if cmd == nil {
					t.Errorf("Update() expected quit command but got nil")
				}
			}

			m := updatedModel.(Model)
			if tt.key == "r" && tt.initialState == StateError {
				// Retry should reset state to connecting
				if m.state != StateConnecting {
					t.Errorf("Update() after retry state = %v, want %v", m.state, StateConnecting)
				}
				if m.error != nil {
					t.Errorf("Update() after retry error should be nil, got %v", m.error)
				}
				if m.retryCount != 0 {
					t.Errorf("Update() after retry retryCount = %d, want 0", m.retryCount)
				}
			}
		})
	}
}

func TestModel_Update_StepMsg(t *testing.T) {
	tests := []struct {
		name        string
		stepState   LoadingState
		expectState LoadingState
	}{
		{
			name:        "step to health check",
			stepState:   StateHealthCheck,
			expectState: StateHealthCheck,
		},
		{
			name:        "step to success",
			stepState:   StateSuccess,
			expectState: StateSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
			model := NewModel(httpClient, logger)

			msg := stepMsg{state: tt.stepState}
			updatedModel, _ := model.Update(msg)

			m := updatedModel.(Model)
			if m.state != tt.expectState {
				t.Errorf("Update() state = %v, want %v", m.state, tt.expectState)
			}
		})
	}
}

func TestModel_Update_HealthCheckResult(t *testing.T) {
	tests := []struct {
		name            string
		result          HealthCheckResult
		initialRetries  int
		expectedState   LoadingState
		expectedRetries int
	}{
		{
			name: "successful health check",
			result: HealthCheckResult{
				Success:   true,
				Status:    "ok",
				Latency:   "10ms",
				Timestamp: "2023-01-01T00:00:00Z",
			},
			initialRetries:  0,
			expectedState:   StateSuccess,
			expectedRetries: 0,
		},
		{
			name: "failed health check - first retry",
			result: HealthCheckResult{
				Success:   false,
				Status:    "error",
				Error:     errors.New("connection failed"),
				Latency:   "0ms",
				Timestamp: "2023-01-01T00:00:00Z",
			},
			initialRetries:  0,
			expectedState:   StateRetrying,
			expectedRetries: 1,
		},
		{
			name: "failed health check - max retries reached",
			result: HealthCheckResult{
				Success:   false,
				Status:    "error",
				Error:     errors.New("connection failed"),
				Latency:   "0ms",
				Timestamp: "2023-01-01T00:00:00Z",
			},
			initialRetries:  maxRetries,
			expectedState:   StateError,
			expectedRetries: maxRetries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
			model := NewModel(httpClient, logger)
			model.retryCount = tt.initialRetries

			updatedModel, _ := model.Update(tt.result)

			m := updatedModel.(Model)
			if m.state != tt.expectedState {
				t.Errorf("Update() state = %v, want %v", m.state, tt.expectedState)
			}

			if m.retryCount != tt.expectedRetries {
				t.Errorf("Update() retryCount = %d, want %d", m.retryCount, tt.expectedRetries)
			}

			if tt.result.Success {
				if m.healthData == nil {
					t.Errorf("Update() healthData should not be nil for successful result")
				}
				if m.error != nil {
					t.Errorf("Update() error should be nil for successful result, got %v", m.error)
				}
			} else {
				if m.error == nil {
					t.Errorf("Update() error should not be nil for failed result")
				}
			}
		})
	}
}

func TestModel_View(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		expectEmpty bool
	}{
		{
			name:        "zero dimensions",
			width:       0,
			height:      0,
			expectEmpty: false, // Should return initializing message
		},
		{
			name:        "normal dimensions",
			width:       80,
			height:      24,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
			model := NewModel(httpClient, logger)
			model.width = tt.width
			model.height = tt.height

			view := model.View()

			if tt.expectEmpty && view != "" {
				t.Errorf("View() = %q, want empty string", view)
			}
			if !tt.expectEmpty && view == "" {
				t.Errorf("View() returned empty string, expected non-empty")
			}
		})
	}
}
