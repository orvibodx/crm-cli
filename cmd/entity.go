package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
	"github.com/orvibodx/crm-cli/internal/filter"
	"github.com/orvibodx/crm-cli/internal/output"
)

var (
	searchStr     string
	filterStrs    []string
	pageNum       int
	limitNum      int
	entityID      string
	fieldsStr     string
	dryRun        bool
	createdPreset string
)

func init() {
	for _, name := range api.EntityNames() {
		entityName := name
		entityCmd := &cobra.Command{
			Use:   entityName,
			Short: fmt.Sprintf("Operate on %s", entityName),
		}

		listCmd := &cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List %s records", entityName),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runList(entityName)
			},
		}
		listCmd.Flags().StringVarP(&searchStr, "search", "s", "", "Search keyword")
		listCmd.Flags().StringArrayVar(&filterStrs, "filter", nil, "Filter: field:op:value (repeatable)")
		listCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number")
		listCmd.Flags().IntVarP(&limitNum, "limit", "l", 15, "Page size")
		listCmd.Flags().StringVar(&fieldsStr, "fields", "", "Output fields (comma-separated)")
		listCmd.Flags().StringVar(&createdPreset, "created", "", "Filter by creation time: today/week/month or custom range (start,end)")

		detailCmd := &cobra.Command{
			Use:   "detail",
			Short: fmt.Sprintf("Get %s detail", entityName),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runDetail(entityName)
			},
		}
		detailCmd.Flags().StringVar(&entityID, "id", "", "Entity ID (required)")
		detailCmd.MarkFlagRequired("id")

		entityCmd.AddCommand(listCmd)
		entityCmd.AddCommand(detailCmd)
		rootCmd.AddCommand(entityCmd)
	}
}

func runList(entityName string) error {
	ent, err := api.GetEntity(entityName)
	if err != nil {
		return err
	}

	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	// Process --created flag
	if createdPreset != "" {
		timeRange, err := filter.ParseTimePreset(createdPreset, time.Now())
		if err != nil {
			// Not a preset, validate as custom range
			if !strings.Contains(createdPreset, ",") {
				return fmt.Errorf("invalid --created value: %q (expected: today/week/month or start,end)", createdPreset)
			}
			timeRange = createdPreset
		}
		filterStrs = append(filterStrs, fmt.Sprintf("createTime:range:%s", timeRange))
	}

	searchItems, err := filter.ParseFilters(filterStrs)
	if err != nil {
		return err
	}

	body := api.SearchBO{
		Page:       pageNum,
		Limit:      limitNum,
		PageType:   1,
		Search:     searchStr,
		Label:      ent.Label,
		SearchList: searchItems,
	}

	path := ent.APIPrefix + "/queryPageList"

	resp, err := client.DoRequest("POST", path, body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}

func runDetail(entityName string) error {
	ent, err := api.GetEntity(entityName)
	if err != nil {
		return err
	}

	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	path := fmt.Sprintf("%s/queryById/%s", ent.APIPrefix, entityID)

	resp, err := client.DoRequest("POST", path, nil, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}

func resolveEnv() string {
	if env != "" {
		return env
	}
	cfg, _ := client.LoadConfig()
	if cfg.Env != "" {
		return cfg.Env
	}
	return "prod"
}

func printOutput(data json.RawMessage) error {
	if format == "table" {
		var pageResult struct {
			List []map[string]interface{} `json:"list"`
		}
		if err := json.Unmarshal(data, &pageResult); err == nil && len(pageResult.List) > 0 {
			return output.PrintTable(pageResult.List, fieldsStr)
		}
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err == nil {
			rows := []map[string]interface{}{obj}
			return output.PrintTable(rows, fieldsStr)
		}
	}
	return output.PrintRawJSON(data)
}
