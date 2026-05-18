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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SetVersion(v string) {
	rootCmd.Version = v
}
