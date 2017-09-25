package ldebug

import "encoding/json"

// DumpJSON dump an object as indented json
func DumpJSON(o interface{}) string {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
