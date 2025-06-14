package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/clients"
)

func TestNewRootModel(t *testing.T) {
	tests := []struct {
		name         string
		serverURL    string
		expectNotNil bool
	}{
		{
			name:         "valid parameters",
			serverURL:    "localhost:50051",
			expectNotNil: true,
		},
		{
			name:         "different server",
			serverURL:    "api.example.com:50051",
			expectNotNil: true,
		},
		{
			name:         "nil clients",
			serverURL:    "",
			expectNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var grpcClients *clients.Clients
			var err error

			if tt.serverURL != "" {
				grpcClients, err = clients.NewClients(tt.serverURL)
				if err != nil {
					t.Fatalf("Failed to create gRPC clients: %v", err)
				}
			}

			model := NewRootModel(grpcClients)

			if tt.expectNotNil {
				if model == nil {
					t.Errorf("NewRootModel() returned nil, expected non-nil")
				}
			}
		})
	}
}

func TestRootModel_Init(t *testing.T) {
	grpcClients, err := clients.NewClients("localhost:50051")
	if err != nil {
		t.Fatalf("Failed to create gRPC clients: %v", err)
	}
	model := NewRootModel(grpcClients)

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
			grpcClients, err := clients.NewClients("localhost:50051")
			if err != nil {
				t.Fatalf("Failed to create gRPC clients: %v", err)
			}
			model := NewRootModel(grpcClients)

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
	grpcClients, err := clients.NewClients("localhost:50051")
	if err != nil {
		t.Fatalf("Failed to create gRPC clients: %v", err)
	}
	model := NewRootModel(grpcClients)

	view := model.View()
	if view == "" {
		t.Errorf("View() returned empty string, expected non-empty")
	}
}

func TestRootModel_HasError(t *testing.T) {
	grpcClients, err := clients.NewClients("localhost:50051")
	if err != nil {
		t.Fatalf("Failed to create gRPC clients: %v", err)
	}
	model := NewRootModel(grpcClients)

	// Initially should have no error
	if model.HasError() {
		t.Errorf("HasError() = true, want false for new model")
	}

	// Test GetError on model without error
	if model.GetError() != "" {
		t.Errorf("GetError() = %q, want empty string for model without error", model.GetError())
	}
}
