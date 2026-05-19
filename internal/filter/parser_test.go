package filter

import (
	"testing"
	"time"
)

func TestParseFilter_Eq(t *testing.T) {
	item, err := ParseFilter("mobile:eq:13800138000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "mobile" {
		t.Errorf("expected name=mobile, got %s", item.Name)
	}
	if item.Type != 1 {
		t.Errorf("expected type=1, got %d", item.Type)
	}
	if item.FormType != "mobile" {
		t.Errorf("expected formType=mobile, got %s", item.FormType)
	}
	if len(item.Values) != 1 || item.Values[0] != "13800138000" {
		t.Errorf("expected values=[13800138000], got %v", item.Values)
	}
}

func TestParseFilter_Contains(t *testing.T) {
	item, err := ParseFilter("customerName:contains:华为")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 3 {
		t.Errorf("expected type=3, got %d", item.Type)
	}
	if item.FormType != "text" {
		t.Errorf("expected formType=text, got %s", item.FormType)
	}
}

func TestParseFilter_Empty(t *testing.T) {
	item, err := ParseFilter("mobile:empty:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 5 {
		t.Errorf("expected type=5, got %d", item.Type)
	}
	if len(item.Values) != 0 {
		t.Errorf("expected empty values, got %v", item.Values)
	}
}

func TestParseFilter_Range(t *testing.T) {
	item, err := ParseFilter("money:range:100,500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 14 {
		t.Errorf("expected type=14, got %d", item.Type)
	}
	if len(item.Values) != 2 || item.Values[0] != "100" || item.Values[1] != "500" {
		t.Errorf("expected values=[100,500], got %v", item.Values)
	}
}

func TestParseFilter_InvalidOp(t *testing.T) {
	_, err := ParseFilter("mobile:badop:value")
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
}

func TestParseFilter_InvalidFormat(t *testing.T) {
	_, err := ParseFilter("nocolon")
	if err == nil {
		t.Fatal("expected error for missing colon")
	}
}

func TestParseFilters(t *testing.T) {
	items, err := ParseFilters([]string{"mobile:eq:123", "customerName:contains:华为"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestParseFilter_DefaultFormType(t *testing.T) {
	item, err := ParseFilter("unknownField:eq:val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.FormType != "text" {
		t.Errorf("expected default formType=text, got %s", item.FormType)
	}
}

func TestParseTimePreset_Today(t *testing.T) {
	// 2026-05-19 is a Tuesday
	now := parseTime(t, "2026-05-19")
	result, err := ParseTimePreset("today", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-19,2026-05-19"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_Week(t *testing.T) {
	// 2026-05-19 is a Tuesday, Monday is 2026-05-18
	now := parseTime(t, "2026-05-19")
	result, err := ParseTimePreset("week", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-18,2026-05-19"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_Week_Monday(t *testing.T) {
	// 2026-05-18 is a Monday
	now := parseTime(t, "2026-05-18")
	result, err := ParseTimePreset("week", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-18,2026-05-18"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_Week_Sunday(t *testing.T) {
	// 2026-05-17 is a Sunday, Monday of that week is 2026-05-11
	now := parseTime(t, "2026-05-17")
	result, err := ParseTimePreset("week", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-11,2026-05-17"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_Month(t *testing.T) {
	// 2026-05-19 is in May, 1st is 2026-05-01
	now := parseTime(t, "2026-05-19")
	result, err := ParseTimePreset("month", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-01,2026-05-19"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_Month_FirstDay(t *testing.T) {
	// 2026-05-01 is the first day of May
	now := parseTime(t, "2026-05-01")
	result, err := ParseTimePreset("month", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2026-05-01,2026-05-01"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestParseTimePreset_InvalidPreset(t *testing.T) {
	now := parseTime(t, "2026-05-19")
	_, err := ParseTimePreset("invalid", now)
	if err == nil {
		t.Fatal("expected error for invalid preset")
	}
}

func parseTime(t *testing.T, dateStr string) time.Time {
	const layout = "2006-01-02"
	tm, err := time.Parse(layout, dateStr)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	return tm
}
