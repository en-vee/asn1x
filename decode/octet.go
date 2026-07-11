package decode

import (
	"encoding/hex"
	"unicode/utf8"
)

// decodeOctetString returns a JSON-friendly value for OCTET STRING content.
// Textual payloads (common for 3GPP time and identifier fields encoded as
// OCTET STRING) are returned as strings; binary payloads stay hex-encoded.
func decodeOctetString(content []byte) any {
	if isTextualOctets(content) {
		return string(content)
	}
	return hex.EncodeToString(content)
}

func isTextualOctets(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	if !utf8.Valid(content) {
		return false
	}
	for _, b := range content {
		if b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}
