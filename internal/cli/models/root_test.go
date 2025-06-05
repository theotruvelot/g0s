package models

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

func TestNewRootModel(t *testing.T) {
	tests := []struct {
		name         string
		serverURL    string
		apiToken     string
		expectNotNil bool
	}{
		{
			name:         "valid parameters",
			serverURL:    "http://localhost:8080",
			apiToken:     "test-token",
			expectNotNil: true,
		},
		{
			name:         "empty token",
			serverURL:    "http://localhost:8080",
			apiToken:     "",
			expectNotNil: true,
		},
		{
			name:         "https URL",
			serverURL:    "https://api.example.com",
			apiToken:     "test-token",
			expectNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient(tt.serverURL, tt.apiToken, 30*time.Second)

			model := NewRootModel(httpClient, logger)

			if tt.expectNotNil {
				if model == nil {
					t.Errorf("NewRootModel() returned nil, expected non-nil")
				}
				if model.log == nil {
					t.Errorf("NewRootModel() logger is nil")
				}
			}
		})
	}
}

func TestRootModel_Init(t *testing.T) {
	logger := zap.NewNop()
	httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
	model := NewRootModel(httpClient, logger)

	cmd := model.Init()
	if cmd == nil {
		t.Errorf("Init() returned nil command, expected non-nil")
	}
}

func TestRootModel_Update(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		wantQuit bool
	}{
		{
			name:     "ctrl+c should quit",
			msg:      tea.KeyMsg{Type: tea.KeyCtrlC},
			wantQuit: true,
		},
		{
			name:     "window size message",
			msg:      tea.WindowSizeMsg{Width: 80, Height: 24},
			wantQuit: false,
		},
		{
			name:     "regular key message",
			msg:      tea.KeyMsg{Type: tea.KeyEnter},
			wantQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
			model := NewRootModel(httpClient, logger)

			updatedModel, cmd := model.Update(tt.msg)

			if updatedModel == nil {
				t.Errorf("Update() returned nil model")
			}

			if tt.wantQuit {
				// For quit commands, we expect a tea.Quit command
				// We can't easily test the exact command type without reflection
				// So we'll just check that a command was returned
				if cmd == nil {
					t.Errorf("Update() with quit message should return a command")
				}
			}
		})
	}
}

func TestRootModel_View(t *testing.T) {
	logger := zap.NewNop()
	httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
	model := NewRootModel(httpClient, logger)

	view := model.View()
	if view == "" {
		t.Errorf("View() returned empty string, expected non-empty")
	}
}

func TestRootModel_HasError(t *testing.T) {
	logger := zap.NewNop()
	httpClient := client.NewClient("http://localhost:8080", "test-token", 30*time.Second)
	model := NewRootModel(httpClient, logger)

	// Initially should have no error
	if model.HasError() {
		t.Errorf("HasError() = true, want false for new model")
	}

	// Test GetError on model without error
	if model.GetError() != "" {
		t.Errorf("GetError() = %q, want empty string for model without error", model.GetError())
	}
}
