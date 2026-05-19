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