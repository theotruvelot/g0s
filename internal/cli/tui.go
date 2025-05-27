package cli

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/models"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

// RunWithConfig initializes and runs the TUI application
func RunWithConfig(serverURL, apiToken string, log *zap.Logger) error {
	log.Info("Starting TUI application",
		zap.String("server", serverURL))

	httpClient := client.NewClient(serverURL, apiToken, 30*time.Second)

	rootModel := models.NewRootModel(httpClient, log)

	program := tea.NewProgram(
		rootModel,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	finalModel, err := program.Run()
	if err != nil {
		log.Error("TUI application failed", zap.Error(err))
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Handle any final cleanup or error reporting
	if rootModel, ok := finalModel.(models.RootModel); ok {
		if rootModel.HasError() {
			return fmt.Errorf("application ended with error: %s", rootModel.GetError())
		}
	}

	log.Info("TUI application ended successfully")
	return nil
}
