package filter

import (
	"testing"
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
