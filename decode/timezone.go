package decode

import (
	"fmt"
	"regexp"
	"strings"
)

var textTimeZonePattern = regexp.MustCompile(`^([+-]\d{2}:\d{2})(?:\+(\d))?$`)

// decodeMSTimeZone decodes a 3GPP MS Time Zone (OCTET STRING SIZE(2)).
// Octet 1 is the offset from UTC in 15-minute units; octet 2 is DST adjustment.
//
// The timezone octet uses GSM semi-octet (nibble-swapped) BCD encoding per
// 3GPP TS 29.060 / GSM 03.40 — the same algorithm Wireshark uses for
// 3GPP-MS-TimeZone. Output format is offset plus DST, e.g. 04 01 → +10:00+1.
func decodeMSTimeZone(content []byte) (string, error) {
	if len(content) == 0 {
		return "", nil
	}
	if isTextualOctets(content) {
		return normalizeTextTimeZone(string(content))
	}
	if len(content) != 2 {
		return "", fmt.Errorf("MSTimeZone length %d, want 2", len(content))
	}

	units := gsmTimeZoneUnits(content[0])
	dst, err := dstAdjustmentHours(content[1])
	if err != nil {
		return "", err
	}
	return formatMSTimeZone(formatOffsetFromQuarterHours(units), dst), nil
}

// gsmTimeZoneUnits decodes the first octet of an MS Time Zone value.
// See Wireshark's 3GPP-MS-TimeZone dissector and GSM 03.40 TP-SCTS timezone.
func gsmTimeZoneUnits(b byte) int {
	units := int(b>>4) + int(b&0x07)*10
	if b&0x08 != 0 {
		units = -units
	}
	return units
}

func dstAdjustmentHours(b byte) (int, error) {
	switch b & 0x03 {
	case 0:
		return 0, nil
	case 1:
		return 1, nil
	case 2:
		return 2, nil
	default:
		return 0, fmt.Errorf("MSTimeZone reserved daylight saving value %02x", b&0x03)
	}
}

func formatMSTimeZone(offset string, dst int) string {
	return offset + fmt.Sprintf("+%d", dst)
}

func formatOffsetFromQuarterHours(units int) string {
	sign := "+"
	if units < 0 {
		sign = "-"
		units = -units
	}
	hours := units / 4
	mins := (units % 4) * 15
	return fmt.Sprintf("%s%02d:%02d", sign, hours, mins)
}

func normalizeTextTimeZone(s string) (string, error) {
	s = strings.TrimSpace(s)
	if m := textTimeZonePattern.FindStringSubmatch(s); m != nil {
		if m[2] != "" {
			return m[1] + "+" + m[2], nil
		}
		return m[1] + "+0", nil
	}
	return s, nil
}
