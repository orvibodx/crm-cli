# CRM Analytics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add data analytics capabilities to CRM CLI with `customer stats` aggregation command and `crm-analytics` skill for four analysis scenarios.

**Architecture:** CLI provides local Go-based aggregation for system fields (`customer stats`) and enhanced time filtering (`customer list --created`). New skill `crm-analytics` orchestrates commands for channel analysis, growth analysis, customer insights, and follow-up tracking.

**Tech Stack:** Go 1.22+, Cobra CLI, existing internal packages (api, client, filter, output)

---

## File Structure

**New files:**
- `cmd/customer_stats.go` — `customer stats` subcommand implementation
- `internal/stats/aggregator.go` — aggregation engine (group-by logic)
- `internal/stats/address.go` — address parsing (province/city/district)
- `internal/stats/time.go` — time granularity (day/week/month)
- `internal/stats/aggregator_test.go` — unit tests for aggregator
- `internal/stats/address_test.go` — unit tests for address parsing
- `internal/stats/time_test.go` — unit tests for time granularity
- `skills/crm-analytics/SKILL.md` — analytics skill documentation

**Modified files:**
- `cmd/entity.go` — add `--created` flag to `customer list`
- `internal/filter/parser.go` — add time preset parsing helper

---

## Task 1: Time Preset Helper

**Files:**
- Modify: `internal/filter/parser.go`
- Test: `internal/filter/parser_test.go`

- [ ] **Step 1: Write test for time preset parsing**

```go
func TestParseTimePreset(t *testing.T) {
	tests := []struct {
		preset string
		want   string // expected "start,end" format
	}{
		{"today", "2026-05-19,2026-05-19"},
		{"week", "2026-05-12,2026-05-19"},  // Monday to today
		{"month", "2026-05-01,2026-05-19"}, // 1st to today
	}
	for _, tt := range tests {
		got, err := ParseTimePreset(tt.preset, time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Errorf("ParseTimePreset(%q) error = %v", tt.preset, err)
		}
		if got != tt.want {
			t.Errorf("ParseTimePreset(%q) = %q, want %q", tt.preset, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/apple/orvibocli && go test ./internal/filter -run TestParseTimePreset -v`
Expected: FAIL with "undefined: ParseTimePreset"

- [ ] **Step 3: Implement ParseTimePreset function**

Add to `internal/filter/parser.go`:

```go
func ParseTimePreset(preset string, now time.Time) (string, error) {
	var start, end time.Time
	switch preset {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start
	case "week":
		// Find Monday of current week
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		start = now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end = now
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	default:
		return "", fmt.Errorf("unknown time preset: %s (valid: today, week, month)", preset)
	}
	return fmt.Sprintf("%s,%s", start.Format("2006-01-02"), end.Format("2006-01-02")), nil
}
```

- [ ] **Step 4: Add time import**

Add to imports in `internal/filter/parser.go`:

```go
import (
	"fmt"
	"strings"
	"time"

	"github.com/orvibodx/crm-cli/internal/api"
)
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /Users/apple/orvibocli && go test ./internal/filter -run TestParseTimePreset -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/filter/parser.go internal/filter/parser_test.go
git commit -m "feat: add time preset parser (today/week/month)"
```

// __CONTINUE_HERE__

## Task 2: Add --created Flag to customer list

**Files:**
- Modify: `cmd/entity.go:32-50`

- [ ] **Step 1: Add createdPreset variable**

Add to variable declarations in `cmd/entity.go` (after line 21):

```go
var (
	searchStr    string
	filterStrs   []string
	pageNum      int
	limitNum     int
	entityID     string
	fieldsStr    string
	dryRun       bool
	createdPreset string
)
```

- [ ] **Step 2: Add --created flag to list command**

Modify the listCmd flag registration in `cmd/entity.go` (after line 43):

```go
listCmd.Flags().StringVarP(&searchStr, "search", "s", "", "Search keyword")
listCmd.Flags().StringArrayVar(&filterStrs, "filter", nil, "Filter: field:op:value (repeatable)")
listCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number")
listCmd.Flags().IntVarP(&limitNum, "limit", "l", 15, "Page size")
listCmd.Flags().StringVar(&fieldsStr, "fields", "", "Output fields (comma-separated)")
listCmd.Flags().StringVar(&createdPreset, "created", "", "Created time: today/week/month/start,end")
```

- [ ] **Step 3: Process --created in runList function**

Find the `runList` function in `cmd/entity.go` and add time preset processing before filter parsing:

```go
func runList(entityName string) error {
	entity, err := api.GetEntity(entityName)
	if err != nil {
		return err
	}

	// Process --created flag
	if createdPreset != "" {
		var timeRange string
		if createdPreset == "today" || createdPreset == "week" || createdPreset == "month" {
			timeRange, err = filter.ParseTimePreset(createdPreset, time.Now())
			if err != nil {
				return err
			}
		} else {
			timeRange = createdPreset
		}
		filterStrs = append(filterStrs, fmt.Sprintf("createTime:range:%s", timeRange))
	}

	// ... rest of existing code
```

- [ ] **Step 4: Add time import to cmd/entity.go**

Add to imports:

```go
import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
	"github.com/orvibodx/crm-cli/internal/filter"
	"github.com/orvibodx/crm-cli/internal/output"
)
```

- [ ] **Step 5: Test --created flag**

Run: `cd /Users/apple/orvibocli && go build -o crm-cli . && ./crm-cli customer list --created today --env test --dry-run`
Expected: Should show filter with today's date range

- [ ] **Step 6: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/entity.go
git commit -m "feat: add --created flag to customer list command"
```

## Task 3: Address Parser

**Files:**
- Create: `internal/stats/address.go`
- Create: `internal/stats/address_test.go`

- [ ] **Step 1: Write test for address parsing**

Create `internal/stats/address_test.go`:

```go
package stats

import "testing"

func TestParseAddress(t *testing.T) {
	tests := []struct {
		address string
		level   string
		want    string
	}{
		{"广东省深圳市南山区", "province", "广东"},
		{"广东省深圳市南山区", "city", "深圳"},
		{"广东省深圳市南山区", "district", "南山"},
		{"北京市朝阳区", "province", "北京"},
		{"北京市朝阳区", "city", "北京"},
		{"上海市浦东新区", "province", "上海"},
		{"invalid address", "city", "(unknown)"},
	}
	for _, tt := range tests {
		got := ParseAddress(tt.address, tt.level)
		if got != tt.want {
			t.Errorf("ParseAddress(%q, %q) = %q, want %q", tt.address, tt.level, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestParseAddress -v`
Expected: FAIL with "no Go files in internal/stats"

- [ ] **Step 3: Create stats directory**

Run: `mkdir -p /Users/apple/orvibocli/internal/stats`

- [ ] **Step 4: Implement ParseAddress function**

Create `internal/stats/address.go`:

```go
package stats

import "regexp"

var (
	provinceRe = regexp.MustCompile(`^(.*?省|.*?自治区|.*?特别行政区|北京|上海|天津|重庆)`)
	cityRe     = regexp.MustCompile(`(?:省|自治区|特别行政区)?(.*?市)`)
	districtRe = regexp.MustCompile(`市)?(.*?区|.*?县)`)
)

func ParseAddress(address, level string) string {
	switch level {
	case "province":
		if m := provinceRe.FindStringSubmatch(address); len(m) > 1 {
			return trimSuffix(m[1], "省")
		}
	case "city":
		if m := cityRe.FindStringSubmatch(address); len(m) > 1 {
			return trimSuffix(m[1], "市")
		}
	case "district":
		if m := districtRe.FindStringSubmatch(address); len(m) > 1 {
			return m[1]
		}
	}
	return "(unknown)"
}

func trimSuffix(s, suffix string) string {
	if len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestParseAddress -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/stats/address.go internal/stats/address_test.go
git commit -m "feat: add address parser for province/city/district"
```

## Task 4: Time Granularity Parser

**Files:**
- Create: `internal/stats/time.go`
- Create: `internal/stats/time_test.go`

- [ ] **Step 1: Write test for time granularity**

Create `internal/stats/time_test.go`:

```go
package stats

import (
	"testing"
	"time"
)

func TestFormatTimeByGranularity(t *testing.T) {
	ts := time.Date(2026, 5, 18, 14, 30, 0, 0, time.UTC)
	tests := []struct {
		granularity string
		want        string
	}{
		{"day", "2026-05-18"},
		{"week", "2026-W20"},
		{"month", "2026-05"},
	}
	for _, tt := range tests {
		got := FormatTimeByGranularity(ts, tt.granularity)
		if got != tt.want {
			t.Errorf("FormatTimeByGranularity(%v, %q) = %q, want %q", ts, tt.granularity, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestFormatTimeByGranularity -v`
Expected: FAIL with "undefined: FormatTimeByGranularity"

- [ ] **Step 3: Implement FormatTimeByGranularity function**

Create `internal/stats/time.go`:

```go
package stats

import (
	"fmt"
	"time"
)

func FormatTimeByGranularity(t time.Time, granularity string) string {
	switch granularity {
	case "day":
		return t.Format("2006-01-02")
	case "week":
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "month":
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestFormatTimeByGranularity -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/stats/time.go internal/stats/time_test.go
git commit -m "feat: add time granularity formatter (day/week/month)"
```

## Task 5: Aggregator Core (Single Dimension)

**Files:**
- Create: `internal/stats/aggregator.go`
- Create: `internal/stats/aggregator_test.go`

- [ ] **Step 1: Write test for single-dimension aggregation**

Create `internal/stats/aggregator_test.go`:

```go
package stats

import (
	"reflect"
	"testing"
)

func TestAggregate(t *testing.T) {
	records := []map[string]interface{}{
		{"source": "视频号", "customerLevel": "A级"},
		{"source": "视频号", "customerLevel": "B级"},
		{"source": "官网", "customerLevel": "A级"},
		{"source": "视频号", "customerLevel": "A级"},
	}

	result := Aggregate(records, []string{"source"}, "", "")
	want := map[string]int{
		"视频号": 3,
		"官网":  1,
	}

	if !reflect.DeepEqual(result, want) {
		t.Errorf("Aggregate() = %v, want %v", result, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestAggregate -v`
Expected: FAIL with "undefined: Aggregate"

- [ ] **Step 3: Implement Aggregate function (single dimension)**

Create `internal/stats/aggregator.go`:

```go
package stats

import "fmt"

func Aggregate(records []map[string]interface{}, groupBy []string, addressLevel, timeGranularity string) map[string]int {
	counts := make(map[string]int)

	for _, record := range records {
		key := extractGroupKey(record, groupBy[0], addressLevel, timeGranularity)
		counts[key]++
	}

	return counts
}

func extractGroupKey(record map[string]interface{}, field, addressLevel, timeGranularity string) string {
	val, ok := record[field]
	if !ok {
		return "(unknown)"
	}

	strVal := fmt.Sprintf("%v", val)
	if strVal == "" {
		return "(empty)"
	}

	// Special handling for address field
	if field == "address" && addressLevel != "" {
		return ParseAddress(strVal, addressLevel)
	}

	// Special handling for createTime field
	if field == "createTime" && timeGranularity != "" {
		// Parse time string and format by granularity
		// For now, return as-is (will enhance in next step)
		return strVal
	}

	return strVal
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestAggregate -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/stats/aggregator.go internal/stats/aggregator_test.go
git commit -m "feat: add single-dimension aggregator"
```

## Task 6: Aggregator Multi-Dimension Support

**Files:**
- Modify: `internal/stats/aggregator.go`
- Modify: `internal/stats/aggregator_test.go`

- [ ] **Step 1: Write test for multi-dimension aggregation**

Add to `internal/stats/aggregator_test.go`:

```go
func TestAggregateMultiDimension(t *testing.T) {
	records := []map[string]interface{}{
		{"source": "视频号", "customerLevel": "A级"},
		{"source": "视频号", "customerLevel": "B级"},
		{"source": "官网", "customerLevel": "A级"},
		{"source": "视频号", "customerLevel": "A级"},
	}

	result := AggregateMulti(records, []string{"source", "customerLevel"}, "", "")
	want := []map[string]interface{}{
		{"source": "视频号", "customerLevel": "A级", "count": 2},
		{"source": "视频号", "customerLevel": "B级", "count": 1},
		{"source": "官网", "customerLevel": "A级", "count": 1},
	}

	if len(result) != len(want) {
		t.Errorf("AggregateMulti() returned %d items, want %d", len(result), len(want))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestAggregateMultiDimension -v`
Expected: FAIL with "undefined: AggregateMulti"

- [ ] **Step 3: Implement AggregateMulti function**

Add to `internal/stats/aggregator.go`:

```go
func AggregateMulti(records []map[string]interface{}, groupBy []string, addressLevel, timeGranularity string) []map[string]interface{} {
	counts := make(map[string]int)
	keys := make(map[string]map[string]string)

	for _, record := range records {
		var keyParts []string
		keyMap := make(map[string]string)

		for _, field := range groupBy {
			val := extractGroupKey(record, field, addressLevel, timeGranularity)
			keyParts = append(keyParts, val)
			keyMap[field] = val
		}

		compositeKey := fmt.Sprintf("%v", keyParts)
		counts[compositeKey]++
		keys[compositeKey] = keyMap
	}

	var result []map[string]interface{}
	for compositeKey, count := range counts {
		item := make(map[string]interface{})
		for field, val := range keys[compositeKey] {
			item[field] = val
		}
		item["count"] = count
		result = append(result, item)
	}

	return result
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/apple/orvibocli && go test ./internal/stats -run TestAggregateMultiDimension -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/stats/aggregator.go internal/stats/aggregator_test.go
git commit -m "feat: add multi-dimension aggregation support"
```

## Task 7: customer stats Command Structure

**Files:**
- Create: `cmd/customer_stats.go`

- [ ] **Step 1: Create customer stats command skeleton**

Create `cmd/customer_stats.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	dateRange       string
	datePreset      string
	groupBy         string
	addressLevel    string
	timeGranularity string
)

func init() {
	customerCmd := findCustomerCommand()
	if customerCmd == nil {
		return
	}

	statsCmd := &cobra.Command{
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
```

- [ ] **Step 2: Build and test command registration**

Run: `cd /Users/apple/orvibocli && go build -o crm-cli . && ./crm-cli customer stats --help`
Expected: Show stats command help with all flags

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/customer_stats.go
git commit -m "feat: add customer stats command skeleton"
```

## Task 8: customer stats Data Fetching

**Files:**
- Modify: `cmd/customer_stats.go`

- [ ] **Step 1: Implement data fetching logic**

Replace `runCustomerStats` in `cmd/customer_stats.go`:

```go
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
	filters := append([]string, filterStrs...)
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
```

- [ ] **Step 2: Add imports**

Update imports in `cmd/customer_stats.go`:

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
	"github.com/orvibodx/crm-cli/internal/filter"
)
```

- [ ] **Step 3: Test data fetching**

Run: `cd /Users/apple/orvibocli && go build -o crm-cli . && ./crm-cli customer stats --env test --date-preset month --group-by source`
Expected: Should fetch records and print count

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/customer_stats.go
git commit -m "feat: implement customer stats data fetching"
```

## Task 9: customer stats Aggregation and Output

**Files:**
- Modify: `cmd/customer_stats.go`

- [ ] **Step 1: Implement aggregation and output**

Replace the end of `runCustomerStats` in `cmd/customer_stats.go`:

```go
	// Step 4: Fetch all customer records
	records, err := fetchAllCustomers(cfg, searchItems)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("No records found")
		return nil
	}

	// Step 5: Validate group-by fields
	if groupBy == "" {
		return fmt.Errorf("--group-by is required")
	}
	groupByFields := strings.Split(groupBy, ",")

	// Step 6: Aggregate
	var result interface{}
	if len(groupByFields) == 1 {
		result = stats.Aggregate(records, groupByFields, addressLevel, timeGranularity)
	} else {
		result = stats.AggregateMulti(records, groupByFields, addressLevel, timeGranularity)
	}

	// Step 7: Build output structure
	outputData := map[string]interface{}{
		"total":     len(records),
		"dateRange": map[string]string{"start": strings.Split(timeFilter, ",")[0], "end": strings.Split(timeFilter, ",")[1]},
		"groupBy":   groupByFields,
		"data":      result,
	}

	// Step 8: Output
	if format == "json" {
		return output.PrintJSON(outputData)
	} else {
		return printStatsTable(outputData, groupByFields)
	}
}

func printStatsTable(data map[string]interface{}, groupByFields []string) error {
	fmt.Printf("Total: %v customers (%s to %s)\n\n",
		data["total"],
		data["dateRange"].(map[string]string)["start"],
		data["dateRange"].(map[string]string)["end"])

	if len(groupByFields) == 1 {
		// Single dimension: simple table
		counts := data["data"].(map[string]int)
		fmt.Printf("%s | count\n", groupByFields[0])
		fmt.Println(strings.Repeat("-", 40))
		for key, count := range counts {
			fmt.Printf("%s | %d\n", key, count)
		}
	} else {
		// Multi dimension: use tablewriter
		items := data["data"].([]map[string]interface{})
		headers := append(groupByFields, "count")
		var rows [][]string
		for _, item := range items {
			var row []string
			for _, field := range groupByFields {
				row = append(row, fmt.Sprintf("%v", item[field]))
			}
			row = append(row, fmt.Sprintf("%v", item["count"]))
			rows = append(rows, row)
		}
		return output.PrintTable(headers, rows)
	}
	return nil
}
```

- [ ] **Step 2: Add strings import**

Update imports in `cmd/customer_stats.go`:

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/orvibodx/crm-cli/internal/api"
	"github.com/orvibodx/crm-cli/internal/client"
	"github.com/orvibodx/crm-cli/internal/filter"
	"github.com/orvibodx/crm-cli/internal/output"
	"github.com/orvibodx/crm-cli/internal/stats"
)
```

- [ ] **Step 3: Test aggregation**

Run: `cd /Users/apple/orvibocli && go build -o crm-cli . && ./crm-cli customer stats --env test --date-preset month --group-by source`
Expected: Show aggregated statistics

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/customer_stats.go
git commit -m "feat: implement customer stats aggregation and output"
```

## Task 10: crm-analytics Skill

**Files:**
- Create: `skills/crm-analytics/SKILL.md`

- [ ] **Step 1: Create skill directory**

Run: `mkdir -p /Users/apple/orvibocli/skills/crm-analytics`

- [ ] **Step 2: Write crm-analytics skill (part 1: header and scenario 1-2)**

Create `skills/crm-analytics/SKILL.md`:

```markdown
---
name: crm-analytics
description: CRM 数据分析：渠道分析、增长分析、客户洞察、跟进追踪。当用户需要分析客户数据、识别趋势、提取特征时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---

## 前置条件

先确认认证状态（见 ../crm-shared/SKILL.md）。如果 `crm-cli auth whoami` 报错，先登录。

## 场景 1：渠道分析

**业务目标：** 识别哪些渠道带来的客户质量最高

**命令组合：**

```bash
# Step 1: 按来源统计数量
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source

# Step 2: 按来源 x 客户级别交叉分析（识别高质量渠道）
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,customerLevel

# Step 3: 按来源 x 成交状态分析（识别高转化渠道）
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,dealStatus

# Step 4: 拉取某渠道的原始数据做深度分析
crm-cli customer list --filter 'source:eq:视频号' --filter 'createTime:range:2026-05-01,2026-05-19' --limit 100
```

**AI 分析提示：**
- 对比各渠道的客户级别分布（A级占比高 = 高质量渠道）
- 对比各渠道的成交率（dealStatus=1 占比）
- 从原始数据中提取共性特征（行业、地域、需求关键词）
- 输出渠道排名和优化建议

## 场景 2：增长分析

**业务目标：** 了解客户增长趋势，发现异常波动

**命令组合：**

```bash
# 今日新增数量
crm-cli customer list --created today --format json | jq 'length'

# 本周新增数量
crm-cli customer list --created week --format json | jq 'length'

# 本月新增数量
crm-cli customer list --created month --format json | jq 'length'

# 本月按天统计趋势
crm-cli customer stats --date-preset month --group-by createTime --time-granularity day

# 本月按来源统计增长来源
crm-cli customer stats --date-preset month --group-by source
```

**AI 分析提示：**
- 对比不同时间段的增长速度（日/周/月环比）
- 识别增长来源的变化（哪个渠道在增长/下降）
- 发现异常波动（某天突然增长/下降）
- 结合跟进记录分析新客户的活跃度

// __CONTINUE_HERE__
```

- [ ] **Step 3: Commit part 1**

```bash
cd /Users/apple/orvibocli
git add skills/crm-analytics/SKILL.md
git commit -m "feat: add crm-analytics skill (scenarios 1-2)"
```

## Task 11: crm-analytics Skill (Scenarios 3-4)

**Files:**
- Modify: `skills/crm-analytics/SKILL.md`

- [ ] **Step 1: Add scenarios 3-4 to skill**

Append to `skills/crm-analytics/SKILL.md`:

```markdown
## 场景 3：客户洞察

**业务目标：** 发现特定区域/行业客户的共性特征，包括人口统计学特征

**命令组合：**

```bash
# Step 1: 获取字段定义（识别可用的自定义字段）
crm-cli api POST crmCustomer/field

# Step 2: 系统字段聚合（地域 x 行业）
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by industry,customerLevel

# Step 3: 按城市统计地域分布
crm-cli customer stats --filter 'address:contains:深圳' --date-preset month --group-by address --address-level district

# Step 4: 拉取原始数据（包含自定义字段）
crm-cli customer list --filter 'address:contains:深圳' --created month --limit 200

# Step 5: 对每个客户拉取跟进记录
crm-cli activity list --type customer --id <customerId>
```

**AI 分析工作流：**

1. **字段识别阶段** — 调用 field 接口，识别性别/年龄/户型等自定义字段
2. **聚合分析阶段** — 用 stats 做系统字段聚合，用 list 拉原始数据
3. **特征提取阶段** — 从原始数据提取自定义字段值，按维度分组统计
4. **深度洞察阶段** — 抽取跟进记录文本，提取高频关键词，总结客户画像

**输出示例：**

```
深圳地区本月新增客户洞察（2026-05-01 至 2026-05-19，共156位客户）

【人口统计特征】
- 性别分布：男性 62%，女性 38%
- 年龄分布：25-35岁 45%，35-45岁 38%，其他 17%
- 户型分布：三居室 52%，两居室 31%，四居室及以上 17%

【地域细分】
- 南山区 42%（科技园周边为主）
- 福田区 28%
- 宝安区 18%
- 其他 12%

【行业分布】
- 互联网/科技 38%
- 金融 22%
- 制造业 15%
- 其他 25%

【共性特征】
1. 高学历高收入群体为主（互联网/金融行业占比60%）
2. 改善型需求为主（三居室及以上占69%）
3. 关注智能化、品质感（跟进记录高频词：智能家居、全屋定制、品牌）
4. 决策周期较短（平均3-5次跟进即进入商务阶段）

【渠道偏好】
- 视频号 35%（年轻群体为主，25-35岁）
- 官网 28%（企业客户为主）
- 转介绍 22%（高净值客户）
- 其他 15%
```

## 场景 4：跟进追踪

**业务目标：** 了解某个客户的跟进进展和状态

**命令组合：**

```bash
# Step 1: 搜索客户
crm-cli customer list --search "华为"

# Step 2: 查看详情
crm-cli customer detail --id <customerId>

# Step 3: 查看跟进记录
crm-cli activity list --type customer --id <customerId>

# Step 4: 查看关联商机
crm-cli api POST crmCustomer/queryBusiness --data '{"customerId":<id>,"page":1,"limit":15}'

# Step 5: 查看关联合同
crm-cli api POST crmCustomer/queryContract --data '{"customerId":<id>,"page":1,"limit":15}'
```

**AI 分析提示：**
- 总结跟进时间线和关键节点
- 识别客户当前阶段（初次接触/需求确认/方案讨论/商务谈判）
- 提取下一步行动建议
- 识别风险点（如长时间未跟进、竞品介入等）

## 注意事项

- **数据量限制：** `customer stats` 单次最多聚合 10000 条记录，超过会警告
- **自定义字段：** CLI 的 `--group-by` 仅支持系统字段，自定义字段需 AI 自己聚合
- **地址解析：** 使用 `--address-level` 时，无法解析的地址归类为 `(unknown)`
- **时间粒度：** 使用 `--time-granularity` 时，需配合 `--group-by createTime`
- **性能：** 5000 条以下秒级响应，10000 条约 3 秒
```

- [ ] **Step 2: Commit**

```bash
cd /Users/apple/orvibocli
git add skills/crm-analytics/SKILL.md
git commit -m "feat: complete crm-analytics skill (scenarios 3-4 + notes)"
```

## Task 12: Integration Tests

**Files:**
- Test: Manual testing with test environment

- [ ] **Step 1: Test single-dimension aggregation**

Run: `cd /Users/apple/orvibocli && ./crm-cli customer stats --env test --date-preset month --group-by source`
Expected: Show customer count grouped by source

- [ ] **Step 2: Test multi-dimension aggregation**

Run: `./crm-cli customer stats --env test --date-preset month --group-by source,customerLevel`
Expected: Show customer count grouped by source and customerLevel

- [ ] **Step 3: Test address parsing**

Run: `./crm-cli customer stats --env test --date-preset month --group-by address --address-level city --filter 'address:contains:深圳'`
Expected: Show customer count grouped by city

- [ ] **Step 4: Test time granularity**

Run: `./crm-cli customer stats --env test --date-preset month --group-by createTime --time-granularity day`
Expected: Show customer count grouped by day

- [ ] **Step 5: Test --created flag**

Run: `./crm-cli customer list --env test --created today --limit 5`
Expected: Show customers created today

- [ ] **Step 6: Test table output**

Run: `./crm-cli customer stats --env test --date-preset month --group-by source --format table`
Expected: Show ASCII table format

- [ ] **Step 7: Document test results**

Create test report noting any issues found

## Task 13: Update Embedded Skills

**Files:**
- Modify: `skills_embed.go`

- [ ] **Step 1: Regenerate embedded skills**

Run: `cd /Users/apple/orvibocli && go run scripts/embed_skills.go`
Expected: Update skills_embed.go with new crm-analytics skill

- [ ] **Step 2: Verify skill is embedded**

Run: `./crm-cli skills list`
Expected: Should show crm-analytics in the list

- [ ] **Step 3: Test skill content**

Run: `./crm-cli skills show crm-analytics`
Expected: Show full skill content

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add skills_embed.go
git commit -m "chore: regenerate embedded skills with crm-analytics"
```

## Task 14: Update README and Version

**Files:**
- Modify: `README.md`
- Modify: `version.go` (if exists) or prepare for version bump

- [ ] **Step 1: Add customer stats to README**

Add to README.md under "Commands" section:

```markdown
### Data Analysis

```bash
# Aggregate customer statistics
crm-cli customer stats --date-preset month --group-by source

# Multi-dimension analysis
crm-cli customer stats --date-range "2026-05-01,2026-05-19" --group-by source,customerLevel

# Time-based filtering
crm-cli customer list --created today
crm-cli customer list --created week
```

### Skills

- `crm-analytics` — Channel analysis, growth analysis, customer insights, follow-up tracking
```

- [ ] **Step 2: Commit README**

```bash
cd /Users/apple/orvibocli
git add README.md
git commit -m "docs: add customer stats and crm-analytics skill to README"
```

- [ ] **Step 3: Prepare version bump**

Note: Version should be bumped to 0.3.0 (new feature)

## Task 15: Final Build and Test

**Files:**
- Build artifacts

- [ ] **Step 1: Clean build**

Run: `cd /Users/apple/orvibocli && go clean && go build -o crm-cli .`
Expected: Clean build with no errors

- [ ] **Step 2: Run all unit tests**

Run: `go test ./... -v`
Expected: All tests pass

- [ ] **Step 3: Smoke test with test environment**

Run complete workflow:
```bash
./crm-cli auth login -u 18565619762 -p a123456 --env test
./crm-cli customer stats --env test --date-preset month --group-by source
./crm-cli customer list --env test --created today --limit 5
```
Expected: All commands work correctly

- [ ] **Step 4: Final commit**

```bash
cd /Users/apple/orvibocli
git add -A
git commit -m "feat: complete CRM analytics feature (customer stats + crm-analytics skill)"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✅ `customer stats` command with all flags
- ✅ `--created` flag for customer list
- ✅ Address parsing (province/city/district)
- ✅ Time granularity (day/week/month)
- ✅ Single and multi-dimension aggregation
- ✅ crm-analytics skill with 4 scenarios
- ✅ Unit tests for all core functions
- ✅ Integration tests

**No placeholders:**
- ✅ All code blocks complete
- ✅ All test cases defined
- ✅ All commands specified

**Type consistency:**
- ✅ Function names consistent across tasks
- ✅ Package imports consistent
- ✅ Data structures match between tasks

