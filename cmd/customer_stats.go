package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

var (
	dateRange       string
	datePreset      string
	groupBy         string
	addressLevel    string
	timeGranularity string
)

var (
	statsCmd    *cobra.Command
	statsOnce   sync.Once
)

func init() {
	// Defer stats command registration until Execute() time
	// This ensures all package init functions have run
	cobra.EnableCommandSorting = true
}

func RegisterStatsCommand() {
	statsOnce.Do(func() {
		customerCmd := findCustomerCommand()
		if customerCmd == nil {
			return
		}

		statsCmd = &cobra.Command{
			Use:   "stats",
			Short: "Aggregate customer statistics",
			RunE:  runCustomerStats,
		}

		statsCmd.Flags().StringVar(&dateRange, "date-range", "", "Date range: start,end (YYYY-MM-DD)")
		statsCmd.Flags().StringVar(&datePreset, "date-preset", "", "Date preset: today/week/month")
		statsCmd.Flags().StringVar(&groupBy, "group-by", "", "Group by fields (comma-separated)")
		statsCmd.Flags().StringVar(&addressLevel, "address-level", "", "Address level: province/city/district")
		statsCmd.Flags().StringVar(&timeGranularity, "time-granularity", "", "Time granularity: day/week/month")

		customerCmd.AddCommand(statsCmd)
	})
}

func findCustomerCommand() *cobra.Command {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "customer" {
			return cmd
		}
	}
	return nil
}

func runCustomerStats(cmd *cobra.Command, args []string) error {
	fmt.Println("customer stats command - to be implemented")
	return nil
}