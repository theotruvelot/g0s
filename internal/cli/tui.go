package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/clients"
	"github.com/theotruvelot/g0s/internal/cli/config"
	"github.com/theotruvelot/g0s/internal/cli/models"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
)

// RunWithConfig initializes and runs the TUI application
func RunWithConfig(serverURL, apiToken string) error {
	logger.Info("Starting TUI application")

	var grpcClients *clients.Clients
	var err error

	// If CLI parameters are provided, use them
	if serverURL != "" && apiToken != "" {
		logger.Info("Using CLI parameters", zap.String("server", serverURL))
		grpcClients, err = clients.NewClients(serverURL)
		if err != nil {
			logger.Error("Failed to create gRPC clients", zap.Error(err))
			return fmt.Errorf("failed to create gRPC clients: %w", err)
		}
	} else if config.ConfigExists() {
		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Error("Failed to load config", zap.Error(err))
			return fmt.Errorf("failed to load config: %w", err)
		}
		logger.Info("Using config file", zap.String("server", cfg.ServerURL))
		grpcClients, err = clients.NewClients(cfg.ServerURL)
		if err != nil {
			logger.Error("Failed to create gRPC clients", zap.Error(err))
			return fmt.Errorf("failed to create gRPC clients: %w", err)
		}
	} else {
		logger.Info("No configuration found, will configure after login")
		grpcClients = nil
	}

	rootModel := models.NewRootModel(grpcClients)

	program := tea.NewProgram(
		rootModel,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	finalModel, err := program.Run()
	if err != nil {
		logger.Error("TUI application failed", zap.Error(err))
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Handle any final cleanup or error reporting
	if rootModel, ok := finalModel.(*models.RootModel); ok {
		if rootModel.HasError() {
			return fmt.Errorf("application ended with error: %s", rootModel.GetError())
		}
	}

	logger.Info("TUI application ended successfully")
	return nil
}
