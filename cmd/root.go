package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	env    string
	format string
)

var rootCmd = &cobra.Command{
	Use:     "crm-cli",
	Short:   "Orvibo CRM command-line tool",
	Version: "dev",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&env, "env", "prod", "Environment: prod, test, local")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json, table")
}

func Execute() {
	// Register any deferred commands after all init functions have run
	registerDeferredCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func registerDeferredCommands() {
	// This will be called after all package init functions have completed
	// allowing customer_stats to find the customer command created by entity.go
	RegisterStatsCommand()
}

func SetVersion(v string) {
	rootCmd.Version = v
}
