package decode

import "encoding/json"

// ToJSON marshals a decoded value as indented JSON.
func ToJSON(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
