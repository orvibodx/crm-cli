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

// __CONTINUE_HERE__
