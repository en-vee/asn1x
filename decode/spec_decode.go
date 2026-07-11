package decode

import (
	"fmt"
	"strings"
	"time"

	"github.com/en-vee/asn1x/ber"
)

func decodeWithSpecType(typeName string, content []byte) (any, error) {
	switch normalizeSpecType(typeName) {
	case "UTCTime":
		return decodeTimeValue(content, true)
	case "GeneralizedTime":
		return decodeTimeValue(content, false)
	case "IA5String", "UTF8String", "PrintableString", "VisibleString", "NumericString":
		return string(content), nil
	case "Integer":
		return decodeIntegerValue(content)
	case "OctetString":
		return decodeOctetString(content), nil
	default:
		return nil, fmt.Errorf("unsupported decode spec type %q", typeName)
	}
}

func normalizeSpecType(typeName string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "utctime":
		return "UTCTime"
	case "generalizedtime":
		return "GeneralizedTime"
	case "ia5string":
		return "IA5String"
	case "utf8string":
		return "UTF8String"
	case "printablestring":
		return "PrintableString"
	case "visiblestring":
		return "VisibleString"
	case "numericstring":
		return "NumericString"
	case "integer":
		return "Integer"
	case "octetstring":
		return "OctetString"
	default:
		return typeName
	}
}

func decodeTimeValue(content []byte, preferUTCLayouts bool) (string, error) {
	if len(content) == 0 {
		return "", nil
	}
	if !isTextualOctets(content) {
		return "", fmt.Errorf("time value is not textual")
	}

	s := string(content)
	layouts := timeLayouts(preferUTCLayouts)
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format(time.RFC3339), nil
		}
	}
	return s, nil
}

func timeLayouts(preferUTCLayouts bool) []string {
	generalized := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"20060102150405Z",
		"20060102150405.000Z",
		"20060102150405-0700",
		"20060102150405+0700",
	}
	utc := []string{
		"060102150405Z",
		"0601021504Z",
		"060102150405-0700",
		"060102150405+0700",
	}
	if preferUTCLayouts {
		return append(utc, generalized...)
	}
	return append(generalized, utc...)
}

func decodeIntegerValue(content []byte) (any, error) {
	n, err := ber.DecodeInteger(content)
	if err != nil {
		return nil, err
	}
	return n, nil
}
