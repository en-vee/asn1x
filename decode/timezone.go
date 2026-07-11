package decode

import (
	"fmt"
	"regexp"
	"strings"
)

var textTimeZonePattern = regexp.MustCompile(`^([+-])(\d{2}):(\d{2})`)

// decodeMSTimeZone decodes a 3GPP MS Time Zone (OCTET STRING SIZE(2)).
// Octet 1 is the offset from UTC in 15-minute units; octet 2 is DST adjustment.
//
// The timezone octet uses GSM semi-octet (nibble-swapped) BCD encoding per
// 3GPP TS 29.060 / GSM 03.40 — the same algorithm Wireshark uses for
// 3GPP-MS-TimeZone. For example, wire bytes 04 00 (hex "0400") decode to +10:00.
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
	return formatOffsetFromQuarterHours(units), nil
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
		return m[1] + m[2] + ":" + m[3], nil
	}
	return s, nil
}
