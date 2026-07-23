package encode

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func encodeWithSpecType(typeName string, v any) ([]byte, error) {
	switch normalizeSpecType(typeName) {
	case "UTCTime", "GeneralizedTime", "TimeStamp":
		return encodeTimeStamp(v)
	case "MSTimeZone", "TimeZone":
		return encodeMSTimeZone(v)
	case "PLMNID", "PLMNIdentifier", "PLMN":
		return encodePLMNID(v)
	case "IA5String", "UTF8String", "PrintableString", "VisibleString", "NumericString":
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case "Integer":
		n, err := asInt64(v)
		if err != nil {
			return nil, err
		}
		return encodeSignedIntegerContent(n), nil
	case "OctetString":
		return encodeOctetStringValue(v)
	default:
		return nil, fmt.Errorf("unsupported encode spec type %q", typeName)
	}
}

func normalizeSpecType(typeName string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "utctime":
		return "UTCTime"
	case "generalizedtime":
		return "GeneralizedTime"
	case "timestamp":
		return "TimeStamp"
	case "mstimezone", "timezone":
		return "MSTimeZone"
	case "plmnid", "plmnidentifier", "plmn":
		return "PLMNID"
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

// encodeTimeStamp encodes an RFC3339 (or similar) timestamp as a 9-byte 3GPP BCD TimeStamp.
func encodeTimeStamp(v any) ([]byte, error) {
	s, err := asString(v)
	if err != nil {
		return nil, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := parseTimeString(s)
	if err != nil {
		return nil, err
	}
	t = t.UTC()
	yy := t.Year() % 100
	out := make([]byte, 9)
	out[0] = toBCDByte(yy)
	out[1] = toBCDByte(int(t.Month()))
	out[2] = toBCDByte(t.Day())
	out[3] = toBCDByte(t.Hour())
	out[4] = toBCDByte(t.Minute())
	out[5] = toBCDByte(t.Second())
	out[6] = '+'
	out[7] = toBCDByte(0)
	out[8] = toBCDByte(0)
	return out, nil
}

func parseTimeString(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"20060102150405Z",
		"20060102150405.000Z",
		"20060102150405-0700",
		"20060102150405+0700",
		"060102150405Z",
		"0601021504Z",
		"060102150405-0700",
		"060102150405+0700",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

func toBCDByte(n int) byte {
	if n < 0 {
		n = -n
	}
	return byte(((n/10)%10)<<4 | (n % 10))
}

// encodeMSTimeZone encodes "+HH:MM+D" into a 2-octet 3GPP MS Time Zone value.
func encodeMSTimeZone(v any) ([]byte, error) {
	s, err := asString(v)
	if err != nil {
		return nil, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	offset := s
	dst := 0
	if i := strings.LastIndex(s, "+"); i > 0 {
		offset = s[:i]
		d, err := strconv.Atoi(s[i+1:])
		if err != nil {
			return nil, fmt.Errorf("MSTimeZone DST: %w", err)
		}
		dst = d
	}
	units, err := parseOffsetQuarterHours(offset)
	if err != nil {
		return nil, err
	}
	if dst < 0 || dst > 2 {
		return nil, fmt.Errorf("MSTimeZone DST hours %d out of range", dst)
	}
	return []byte{encodeGSMTimeZoneUnits(units), byte(dst)}, nil
}

func parseOffsetQuarterHours(offset string) (int, error) {
	if len(offset) < 6 {
		return 0, fmt.Errorf("MSTimeZone offset %q too short", offset)
	}
	sign := 1
	switch offset[0] {
	case '+':
		sign = 1
	case '-':
		sign = -1
	default:
		return 0, fmt.Errorf("MSTimeZone offset %q missing sign", offset)
	}
	parts := strings.Split(offset[1:], ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("MSTimeZone offset %q invalid", offset)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("MSTimeZone hours: %w", err)
	}
	mins, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("MSTimeZone minutes: %w", err)
	}
	if mins%15 != 0 {
		return 0, fmt.Errorf("MSTimeZone minutes %d not a multiple of 15", mins)
	}
	return sign * (hours*4 + mins/15), nil
}

func encodeGSMTimeZoneUnits(units int) byte {
	negative := units < 0
	if negative {
		units = -units
	}
	tens := units / 10
	ones := units % 10
	b := byte((ones << 4) | (tens & 0x07))
	if negative {
		b |= 0x08
	}
	return b
}

// encodePLMNID encodes {"mcc":"...","mnc":"..."} into a 3-octet BCD PLMN identity.
func encodePLMNID(v any) ([]byte, error) {
	m, err := asStringMap(v)
	if err != nil {
		return nil, err
	}
	mcc := strings.TrimSpace(m["mcc"])
	mnc := strings.TrimSpace(m["mnc"])
	if len(mcc) != 3 {
		return nil, fmt.Errorf("PLMNID mcc %q must be 3 digits", mcc)
	}
	if len(mnc) != 2 && len(mnc) != 3 {
		return nil, fmt.Errorf("PLMNID mnc %q must be 2 or 3 digits", mnc)
	}
	for _, d := range mcc + mnc {
		if d < '0' || d > '9' {
			return nil, fmt.Errorf("PLMNID contains non-digit")
		}
	}

	mcc1 := int(mcc[0] - '0')
	mcc2 := int(mcc[1] - '0')
	mcc3 := int(mcc[2] - '0')

	out := make([]byte, 3)
	out[0] = byte((mcc2 << 4) | mcc1)

	if len(mnc) == 2 {
		mnc1 := int(mnc[0] - '0')
		mnc2 := int(mnc[1] - '0')
		out[1] = byte((0x0f << 4) | mcc3)
		out[2] = byte((mnc2 << 4) | mnc1)
	} else {
		// Decode uses mnc = mncDigit3, mncDigit2, mncDigit1 for three-digit MNC.
		mncDigit3 := int(mnc[0] - '0')
		mncDigit2 := int(mnc[1] - '0')
		mncDigit1 := int(mnc[2] - '0')
		out[1] = byte((mncDigit3 << 4) | mcc3)
		out[2] = byte((mncDigit2 << 4) | mncDigit1)
	}
	return out, nil
}

func asStringMap(v any) (map[string]string, error) {
	switch m := v.(type) {
	case map[string]string:
		return m, nil
	case map[string]any:
		out := make(map[string]string, len(m))
		for k, val := range m {
			s, err := asString(val)
			if err != nil {
				return nil, fmt.Errorf("PLMNID field %q: %w", k, err)
			}
			out[k] = s
		}
		return out, nil
	default:
		return nil, fmt.Errorf("PLMNID value must be an object, got %T", v)
	}
}

func encodeOctetStringValue(v any) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return x, nil
	case string:
		if b, ok := decodeHexIfBinary(x); ok {
			return b, nil
		}
		return []byte(x), nil
	default:
		return nil, fmt.Errorf("OCTET STRING value must be string, got %T", v)
	}
}

// decodeHexIfBinary treats s as hex when it matches decode's hex output form:
// even-length lowercase hex that decodes to non-textual bytes. Mixed/upper-case
// strings (e.g. "C54748") are left as raw ASCII to preserve textual OCTET STRINGs.
func decodeHexIfBinary(s string) ([]byte, bool) {
	if len(s) == 0 || len(s)%2 != 0 {
		return nil, false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return nil, false
		}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, false
	}
	if isTextualOctets(b) {
		return nil, false
	}
	return b, true
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
