package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/cli"
	"github.com/theotruvelot/g0s/internal/cli/config"
	"github.com/theotruvelot/g0s/pkg/logger"
	"github.com/theotruvelot/g0s/pkg/utils"
)

type cliError struct {
	op  string
	err error
}

var (
	serverURL string
	apiToken  string
	logLevel  string
)

func (e *cliError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("cli %s: %v", e.op, e.err)
	}
	return e.op
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "g0s-cli",
		Short: "g0s CLI",
		Long:  `g0s CLI - System monitoring interface`,
		RunE:  runCLI,
	}

	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "", "Server URL to request metrics from (optional if config exists)")
	rootCmd.Flags().StringVarP(&apiToken, "token", "t", "", "API token for authentication (optional if config exists)")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level: debug, info, warn, error")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI(_ *cobra.Command, _ []string) error {
	hasCliParams := serverURL != "" && apiToken != ""
	hasConfig := config.ConfigExists()

	if !hasCliParams && !hasConfig {
		fmt.Println("No configuration found. You will be prompted to enter server details.")
	}

	if serverURL != "" {
		if err := utils.ValidateServerURL(serverURL); err != nil {
			return &cliError{op: "validating server URL", err: err}
		}
		serverURL = utils.NormalizeServerURL(serverURL)
	}

	logger.InitLogger(logger.Config{
		Level:      logLevel,
		Format:     "json",
		OutputPath: "./logs.logs",
		Component:  "cli",
	})
	defer logger.Sync()

	if err := cli.RunWithConfig(serverURL, apiToken); err != nil {
		return &cliError{op: "running TUI", err: err}
	}

	return nil
}
