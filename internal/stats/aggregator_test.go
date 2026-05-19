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