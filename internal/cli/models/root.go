package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/pages/loading"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

// RootModel is a simple wrapper around the loading page
type RootModel struct {
	loadingModel loading.Model
	log          *zap.Logger
	err          error
}

// NewRootModel creates a new root model
func NewRootModel(httpClient *client.Client, log *zap.Logger) *RootModel {
	return &RootModel{
		loadingModel: loading.NewModel(httpClient, log),
		log:          log,
	}
}

// Init initializes the root model
func (m RootModel) Init() tea.Cmd {
	m.log.Debug("Initializing application")
	return m.loadingModel.Init()
}

// Update handles messages
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global quit
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			m.log.Info("User requested exit")
			return m, tea.Quit
		}
	}

	// Update loading model
	updatedModel, cmd := m.loadingModel.Update(msg)
	m.loadingModel = updatedModel.(loading.Model)

	return m, cmd
}

// View renders the current view
func (m RootModel) View() string {
	return m.loadingModel.View()
}

// HasError returns true if there's an error
func (m RootModel) HasError() bool {
	return m.err != nil
}

// GetError returns the error message
func (m RootModel) GetError() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}
