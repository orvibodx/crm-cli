package stats

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		level    string
		expected string
	}{
		{
			name:     "full address with province city district",
			address:  "广东省深圳市南山区",
			level:    "province",
			expected: "广东",
		},
		{
			name:     "full address with province city district - city level",
			address:  "广东省深圳市南山区",
			level:    "city",
			expected: "深圳",
		},
		{
			name:     "full address with province city district - district level",
			address:  "广东省深圳市南山区",
			level:    "district",
			expected: "南山",
		},
		{
			name:     "beijing city without province",
			address:  "北京市朝阳区",
			level:    "province",
			expected: "北京",
		},
		{
			name:     "beijing city without province - city level",
			address:  "北京市朝阳区",
			level:    "city",
			expected: "北京",
		},
		{
			name:     "shanghai city without district",
			address:  "上海市浦东新区",
			level:    "province",
			expected: "上海",
		},
		{
			name:     "invalid address",
			address:  "invalid address",
			level:    "city",
			expected: "(unknown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAddress(tt.address, tt.level)
			if result != tt.expected {
				t.Errorf("ParseAddress(%q, %q) = %q, want %q", tt.address, tt.level, result, tt.expected)
			}
		})
	}
}
