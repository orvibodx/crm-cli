package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orvibo/crm-cli/internal/client"
)

var (
	apiData  string
	apiQuery string
)

func init() {
	apiCmd := &cobra.Command{
		Use:   "api [METHOD] [PATH]",
		Short: "Make raw API calls",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPI(args[0], args[1])
		},
	}
	apiCmd.Flags().StringVar(&apiData, "data", "", "JSON body")
	apiCmd.Flags().StringVar(&apiQuery, "query", "", "URL query string (for @RequestParam endpoints)")
	apiCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print request without sending")

	rootCmd.AddCommand(apiCmd)
}

func runAPI(method, path string) error {
	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	var body interface{}
	if apiData != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(apiData), &parsed); err != nil {
			return fmt.Errorf("invalid JSON in --data: %w", err)
		}
		body = parsed
	}

	if apiQuery != "" {
		path = path + "?" + apiQuery
		body = nil
	}

	if dryRun {
		fmt.Printf("DRY RUN: %s %s\n", method, path)
		if body != nil {
			jsonData, _ := json.MarshalIndent(body, "", "  ")
			fmt.Println(string(jsonData))
		}
		return nil
	}

	resp, err := client.DoRequest(method, path, body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}
