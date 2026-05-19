package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
	"github.com/orvibodx/crm-cli/internal/filter"
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
	// Step 1: Parse date range
	var timeFilter string
	if datePreset != "" {
		parsed, err := filter.ParseTimePreset(datePreset, time.Now())
		if err != nil {
			return err
		}
		timeFilter = parsed
	} else if dateRange != "" {
		timeFilter = dateRange
	} else {
		return fmt.Errorf("either --date-range or --date-preset is required")
	}

	// Step 2: Build search filters
	filters := []string{}
	filters = append(filters, fmt.Sprintf("createTime:range:%s", timeFilter))

	searchItems, err := filter.ParseFilters(filters)
	if err != nil {
		return err
	}

	// Step 3: Load config
	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}

	// Step 4: Fetch all customer records
	records, err := fetchAllCustomers(cfg, searchItems)
	if err != nil {
		return err
	}

	fmt.Printf("Fetched %d customer records\n", len(records))
	return nil
}

func fetchAllCustomers(cfg *client.Config, searchItems []api.SearchItem) ([]map[string]interface{}, error) {
	var allRecords []map[string]interface{}
	page := 1
	limit := 1000

	for {
		searchBO := api.SearchBO{
			Page:       page,
			Limit:      limit,
			PageType:   0,
			Label:      2, // customer
			SearchList: searchItems,
		}

		resp, err := client.DoRequest("POST", "crmCustomer/queryPageList", searchBO, cfg.Token, cfg.Env)
		if err != nil {
			return nil, err
		}

		if err := client.CheckResponse(resp); err != nil {
			return nil, err
		}

		var pageData struct {
			List []map[string]interface{} `json:"list"`
		}
		if err := json.Unmarshal(resp.Data, &pageData); err != nil {
			return nil, fmt.Errorf("parse response: %w", err)
		}

		allRecords = append(allRecords, pageData.List...)

		if len(pageData.List) < limit {
			break
		}

		page++

		if len(allRecords) > 10000 {
			fmt.Fprintf(os.Stderr, "Warning: Fetched %d records, performance may degrade. Consider narrowing filters.\n", len(allRecords))
		}
	}

	return allRecords, nil
}