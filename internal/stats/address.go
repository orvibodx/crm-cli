package stats

import (
	"regexp"
)

// ParseAddress extracts province, city, or district from a Chinese address string.
// level can be "province", "city", or "district".
// Returns "(unknown)" if the address cannot be parsed.
func ParseAddress(address, level string) string {
	switch level {
	case "province":
		return parseProvince(address)
	case "city":
		return parseCity(address)
	case "district":
		return parseDistrict(address)
	default:
		return "(unknown)"
	}
}

// parseProvince extracts the province from a Chinese address.
func parseProvince(address string) string {
	// Match province patterns: ends with 省, 自治区, 特别行政区, or is one of the direct-controlled municipalities
	pattern := `^(.*?省|.*?自治区|.*?特别行政区|北京|上海|天津|重庆)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(address)
	if len(matches) > 1 {
		return trimSuffix(matches[1], "省", "自治区", "特别行政区")
	}
	return "(unknown)"
}

// parseCity extracts the city from a Chinese address.
func parseCity(address string) string {
	// Match city patterns: text followed by 市
	// First, try to match after province/region suffix
	pattern := `(?:省|自治区|特别行政区)(.+?市)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(address)
	if len(matches) > 1 {
		return trimSuffix(matches[1], "市")
	}

	// If no province/region suffix found, try to match city at the start
	pattern = `^(.+?市)`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(address)
	if len(matches) > 1 {
		return trimSuffix(matches[1], "市")
	}

	return "(unknown)"
}

// parseDistrict extracts the district from a Chinese address.
func parseDistrict(address string) string {
	// Match district patterns: text followed by 区 or 县
	pattern := `市(.*?(?:区|县))`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(address)
	if len(matches) > 1 {
		return trimSuffix(matches[1], "区", "县")
	}
	return "(unknown)"
}

// trimSuffix removes any of the given suffixes from the string.
func trimSuffix(s string, suffixes ...string) string {
	for _, suffix := range suffixes {
		if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
			return s[:len(s)-len(suffix)]
		}
	}
	return s
}
