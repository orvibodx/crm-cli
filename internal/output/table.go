package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func PrintTable(rows []map[string]interface{}, fields string) error {
	if len(rows) == 0 {
		fmt.Println("No results.")
		return nil
	}

	var headers []string
	if fields != "" {
		headers = strings.Split(fields, ",")
	} else {
		for k := range rows[0] {
			headers = append(headers, k)
		}
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAutoWrap(0),
		tablewriter.WithRowAlignment(tw.AlignLeft),
	)

	for _, row := range rows {
		var vals []string
		for _, h := range headers {
			vals = append(vals, fmtVal(row[h]))
		}
		table.Append(vals)
	}

	table.Render()
	return nil
}

func fmtVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%f", val)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
