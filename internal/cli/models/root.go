package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/clients"
	"github.com/theotruvelot/g0s/internal/cli/config"
	"github.com/theotruvelot/g0s/internal/cli/pages/loading"
	"github.com/theotruvelot/g0s/internal/cli/pages/login"
	"github.com/theotruvelot/g0s/internal/cli/services"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

type PageState int

const (
	PageLogin PageState = iota
	PageLoading
)

type RootModel struct {
	currentPage  PageState
	loginModel   login.Model
	loadingModel loading.Model
	grpcClients  *clients.Clients
	err          error
	width        int
	height       int
}

func NewRootModel(grpcClients *clients.Clients) *RootModel {
	m := &RootModel{
		grpcClients: grpcClients,
		width:       80, // Default width
		height:      24, // Default height
	}

	if config.ConfigExists() {
		logger.Info("Configuration found, going to loading page")
		m.currentPage = PageLoading
		m.loadingModel = loading.NewModel(grpcClients)
		m.loadingModel = m.setModelDimensions(m.loadingModel).(loading.Model)
	} else {
		logger.Info("No configuration found, going to login page")
		m.currentPage = PageLogin
		m.loginModel = login.NewModel(services.NewAuthService(grpcClients))
		m.loginModel = m.setModelDimensions(m.loginModel).(login.Model)
	}

	return m
}

func (m *RootModel) setModelDimensions(model tea.Model) tea.Model {
	sizeMsg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	}

	updatedModel, _ := model.Update(sizeMsg)
	return updatedModel
}

func (m RootModel) Init() tea.Cmd {
	logger.Debug("Initializing application")

	var pageCmd tea.Cmd
	switch m.currentPage {
	case PageLogin:
		pageCmd = m.loginModel.Init()
	case PageLoading:
		pageCmd = m.loadingModel.Init()
	}

	return tea.Batch(
		pageCmd,
		tea.EnterAltScreen, // Ensure we're in alt screen
	)
}

// Update handles messages
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size messages first
	if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowSizeMsg.Width
		m.height = windowSizeMsg.Height
		logger.Debug("Window resized", zap.Int("width", m.width), zap.Int("height", m.height))

		// Propagate the window size to the current page
		switch m.currentPage {
		case PageLogin:
			updatedModel, cmd := m.loginModel.Update(msg)
			if updatedLoginModel, ok := updatedModel.(login.Model); ok {
				m.loginModel = updatedLoginModel
			}
			return m, cmd
		case PageLoading:
			updatedModel, cmd := m.loadingModel.Update(msg)
			m.loadingModel = updatedModel.(loading.Model)
			return m, cmd
		}
	}

	// Handle global quit
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			logger.Info("User requested exit")
			return m, tea.Quit
		}
	}

	switch m.currentPage {
	case PageLogin:
		updatedModel, cmd := m.loginModel.Update(msg)
		if updatedLoginModel, ok := updatedModel.(login.Model); ok {
			m.loginModel = updatedLoginModel

			if m.loginModel.ShouldProceed() {
				logger.Info("Login completed, switching to loading page")

				cfg, err := config.LoadConfig()
				if err != nil {
					logger.Error("Failed to load config after login", zap.Error(err))
					m.err = err
					return m, tea.Quit
				}

				m.grpcClients, err = clients.NewClients(cfg.ServerURL)
				if err != nil {
					logger.Error("Failed to create gRPC clients after login", zap.Error(err))
					m.err = err
					return m, tea.Quit
				}

				m.currentPage = PageLoading
				m.loadingModel = loading.NewModel(m.grpcClients)
				m.loadingModel = m.setModelDimensions(m.loadingModel).(loading.Model)
				return m, m.loadingModel.Init()
			}
		}
		return m, cmd

	case PageLoading:
		updatedModel, cmd := m.loadingModel.Update(msg)
		m.loadingModel = updatedModel.(loading.Model)
		return m, cmd

	default:
		return m, nil
	}
}

// View renders the current view
func (m RootModel) View() string {
	switch m.currentPage {
	case PageLogin:
		return m.loginModel.View()
	case PageLoading:
		return m.loadingModel.View()
	default:
		return "Unknown page"
	}
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
