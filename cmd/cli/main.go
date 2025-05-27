package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theotruvelot/g0s/internal/cli"
	"github.com/theotruvelot/g0s/pkg/logger"
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
		Long:  `g0s CLI`,
		RunE:  runCLI,
	}

	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "", "Server URL to request metrics from")
	rootCmd.Flags().StringVarP(&apiToken, "token", "t", "", "API token for authentication")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level: debug, info, warn, error")

	rootCmd.MarkFlagRequired("server")
	rootCmd.MarkFlagRequired("token")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI(_ *cobra.Command, _ []string) error {
	logger.InitLogger(logger.Config{
		Level:      logLevel,
		Format:     "json",
		OutputPath: "./logs.logs",
		Component:  "cli",
	})
	defer logger.Sync()

	if err := cli.RunWithConfig(serverURL, apiToken, logger.GetLogger()); err != nil {
		return &cliError{op: "running TUI", err: err}
	}

	return nil
}
