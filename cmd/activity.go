package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
)

var (
	activityType string
	activityID   string
)

func init() {
	activityCmd := &cobra.Command{
		Use:   "activity",
		Short: "Activity / follow-up records",
	}

	activityListCmd := &cobra.Command{
		Use:   "list",
		Short: "List activity records for an entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivityList()
		},
	}
	activityListCmd.Flags().StringVar(&activityType, "type", "", "Entity type: customer, leads, contacts, business, etc. (required)")
	activityListCmd.Flags().StringVar(&activityID, "id", "", "Entity ID (required)")
	activityListCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number")
	activityListCmd.Flags().IntVarP(&limitNum, "limit", "l", 15, "Page size")
	activityListCmd.MarkFlagRequired("type")
	activityListCmd.MarkFlagRequired("id")

	activityCmd.AddCommand(activityListCmd)
	rootCmd.AddCommand(activityCmd)
}

func runActivityList() error {
	ent, err := api.GetEntity(activityType)
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

	body := map[string]interface{}{
		"page":           pageNum,
		"limit":          limitNum,
		"pageType":       1,
		"activityType":   ent.Label,
		"activityTypeId": activityID,
	}

	resp, err := client.DoRequest("POST", "crmActivity/getCrmActivityPageList", body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}
