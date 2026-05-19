package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/orvibodx/crm-cli/internal/api"
)

var opToEnum = map[string]int{
	"eq":          1,
	"ne":          2,
	"contains":    3,
	"notcontains": 4,
	"empty":       5,
	"notempty":    6,
	"gt":          7,
	"gte":         8,
	"lt":          9,
	"lte":         10,
	"prefix":      12,
	"suffix":      13,
	"range":       14,
}

var fieldFormTypes = map[string]string{
	"customerName":  "text",
	"mobile":        "mobile",
	"telephone":     "mobile",
	"email":         "email",
	"website":       "text",
	"address":       "text",
	"dealStatus":    "select",
	"customerLevel": "select",
	"industry":      "select",
	"source":        "select",
	"remark":        "text",
	"createTime":    "datetime",
	"updateTime":    "datetime",
	"ownerUserName": "text",
	"lastContent":   "text",
	"nextTime":      "datetime",
	"followup":      "datetime",
	"name":          "text",
	"leadsName":     "text",
	"contactsName":  "text",
	"businessName":  "text",
	"contractName":  "text",
	"money":         "number",
	"contractMoney": "number",
}

func ParseFilter(raw string) (api.SearchItem, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return api.SearchItem{}, fmt.Errorf("invalid filter format, expected field:op[:value], got: %s", raw)
	}

	fieldName := parts[0]
	op := parts[1]
	var value string
	if len(parts) == 3 {
		value = parts[2]
	}

	enumVal, ok := opToEnum[op]
	if !ok {
		return api.SearchItem{}, fmt.Errorf("unknown operator: %s", op)
	}

	var values []string
	switch op {
	case "empty", "notempty":
		values = []string{}
	case "range":
		values = strings.Split(value, ",")
	default:
		if value != "" {
			values = []string{value}
		}
	}

	formType, ok := fieldFormTypes[fieldName]
	if !ok {
		formType = "text"
	}

	return api.SearchItem{
		Name:     fieldName,
		FormType: formType,
		Type:     enumVal,
		Values:   values,
	}, nil
}

func ParseFilters(filters []string) ([]api.SearchItem, error) {
	var items []api.SearchItem
	for _, f := range filters {
		item, err := ParseFilter(f)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func ParseTimePreset(preset string, now time.Time) (string, error) {
	switch preset {
	case "today":
		dateStr := now.Format("2006-01-02")
		return fmt.Sprintf("%s,%s", dateStr, dateStr), nil

	case "week":
		// Get Monday of the current week
		// time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
		// We want: Sunday treated as day 7, so Monday is always the start
		weekday := now.Weekday()
		var daysBack int
		if weekday == 0 {
			// Sunday: go back 6 days to Monday
			daysBack = 6
		} else {
			// Monday-Saturday: go back (weekday - 1) days
			daysBack = int(weekday) - 1
		}
		monday := now.AddDate(0, 0, -daysBack)
		mondayStr := monday.Format("2006-01-02")
		todayStr := now.Format("2006-01-02")
		return fmt.Sprintf("%s,%s", mondayStr, todayStr), nil

	case "month":
		// Get the 1st of the current month
		firstDay := now.AddDate(0, 0, -(now.Day() - 1))
		firstDayStr := firstDay.Format("2006-01-02")
		todayStr := now.Format("2006-01-02")
		return fmt.Sprintf("%s,%s", firstDayStr, todayStr), nil

	default:
		return "", fmt.Errorf("unknown time preset: %s", preset)
	}
}
