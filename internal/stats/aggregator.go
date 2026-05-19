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

	return strVal
}