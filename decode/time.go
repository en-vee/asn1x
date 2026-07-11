package decode

import (
	"fmt"
	"time"
)

// decodeTimeValue decodes a timestamp from either:
//   - ASCII UTCTime / GeneralizedTime text, or
//   - 3GPP TimeStamp (9-byte BCD: YYMMDDhhmmssShhmm per TS 32.298).
func decodeTimeValue(content []byte, preferUTCLayouts bool) (string, error) {
	if len(content) == 0 {
		return "", nil
	}
	if isTextualOctets(content) {
		return decodeTextualTime(content, preferUTCLayouts)
	}
	if len(content) == 9 {
		return decodeBCDTimeStamp(content)
	}
	return "", fmt.Errorf("time value is not textual and length %d is not a 9-byte TimeStamp", len(content))
}

func decodeTextualTime(content []byte, preferUTCLayouts bool) (string, error) {
	s := string(content)
	layouts := timeLayouts(preferUTCLayouts)
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format(time.RFC3339), nil
		}
	}
	return s, nil
}

// decodeBCDTimeStamp decodes the 3GPP charging TimeStamp compact format.
func decodeBCDTimeStamp(content []byte) (string, error) {
	if len(content) != 9 {
		return "", fmt.Errorf("BCD TimeStamp length %d, want 9", len(content))
	}

	yy, err := bcdByte(content[0])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp year: %w", err)
	}
	mm, err := bcdByte(content[1])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp month: %w", err)
	}
	dd, err := bcdByte(content[2])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp day: %w", err)
	}
	hh, err := bcdByte(content[3])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp hour: %w", err)
	}
	min, err := bcdByte(content[4])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp minute: %w", err)
	}
	ss, err := bcdByte(content[5])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp second: %w", err)
	}

	sign, err := timeStampSign(content[6])
	if err != nil {
		return "", err
	}
	tzH, err := bcdByte(content[7])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp timezone hour: %w", err)
	}
	tzM, err := bcdByte(content[8])
	if err != nil {
		return "", fmt.Errorf("BCD TimeStamp timezone minute: %w", err)
	}

	year := 2000 + yy
	offsetSec := sign * (tzH*3600 + tzM*60)
	loc := time.FixedZone("UTC"+formatUTCOffset(sign, tzH, tzM), offsetSec)

	t := time.Date(year, time.Month(mm), dd, hh, min, ss, 0, loc)
	if t.Month() != time.Month(mm) || t.Day() != dd {
		return "", fmt.Errorf("BCD TimeStamp invalid date %04d-%02d-%02d", year, mm, dd)
	}
	return t.UTC().Format(time.RFC3339), nil
}

func bcdByte(b byte) (int, error) {
	hi := int(b >> 4)
	lo := int(b & 0x0f)
	if hi > 9 || lo > 9 {
		return 0, fmt.Errorf("invalid BCD digit %02x", b)
	}
	return hi*10 + lo, nil
}

func timeStampSign(b byte) (int, error) {
	switch b {
	case '+':
		return 1, nil
	case '0':
		// 3GPP uses ASCII '0' to mean '+'.
		return 1, nil
	case '-':
		return -1, nil
	default:
		return 0, fmt.Errorf("BCD TimeStamp unknown sign byte %q (%02x)", b, b)
	}
}

func formatUTCOffset(sign, hour, minute int) string {
	prefix := "+"
	if sign < 0 {
		prefix = "-"
	}
	return fmt.Sprintf("%s%02d:%02d", prefix, hour, minute)
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
