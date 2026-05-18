package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

func PrintRawJSON(raw json.RawMessage) error {
	fmt.Println(string(raw))
	return nil
}
