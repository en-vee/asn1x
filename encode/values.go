package encode

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func asString(v any) (string, error) {
	switch x := v.(type) {
	case nil:
		return "", nil
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case fmt.Stringer:
		return x.String(), nil
	default:
		return "", fmt.Errorf("expected string, got %T", v)
	}
}

func asBool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	default:
		return false, fmt.Errorf("expected bool, got %T", v)
	}
}

func asInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint:
		if uint64(x) > math.MaxInt64 {
			return 0, fmt.Errorf("integer %d overflows int64", x)
		}
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		if x > math.MaxInt64 {
			return 0, fmt.Errorf("integer %d overflows int64", x)
		}
		return int64(x), nil
	case float32:
		return floatToInt64(float64(x))
	case float64:
		return floatToInt64(x)
	case jsonNumber:
		return x.Int64()
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("integer string %q: %w", x, err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", v)
	}
}

func asUint64(v any) (uint64, error) {
	switch x := v.(type) {
	case int:
		if x < 0 {
			return 0, fmt.Errorf("negative integer %d", x)
		}
		return uint64(x), nil
	case int64:
		if x < 0 {
			return 0, fmt.Errorf("negative integer %d", x)
		}
		return uint64(x), nil
	case uint:
		return uint64(x), nil
	case uint32:
		return uint64(x), nil
	case uint64:
		return x, nil
	case float64:
		n, err := floatToInt64(x)
		if err != nil {
			return 0, err
		}
		if n < 0 {
			return 0, fmt.Errorf("negative integer %d", n)
		}
		return uint64(n), nil
	case string:
		n, err := strconv.ParseUint(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("integer string %q: %w", x, err)
		}
		return n, nil
	default:
		n, err := asInt64(v)
		if err != nil {
			return 0, err
		}
		if n < 0 {
			return 0, fmt.Errorf("negative integer %d", n)
		}
		return uint64(n), nil
	}
}

func floatToInt64(f float64) (int64, error) {
	if math.Trunc(f) != f {
		return 0, fmt.Errorf("non-integer number %v", f)
	}
	if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
		return 0, fmt.Errorf("integer %v out of range", f)
	}
	return int64(f), nil
}

// jsonNumber mirrors encoding/json.Number without importing encoding/json in helpers.
type jsonNumber interface {
	Int64() (int64, error)
}

func namedNumberValue(values map[string]int, name string) (int, bool) {
	if values == nil {
		return 0, false
	}
	n, ok := values[name]
	return n, ok
}

func encodeObjectIdentifier(s string) ([]byte, error) {
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("OBJECT IDENTIFIER %q needs at least two arcs", s)
	}
	arcs := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("OBJECT IDENTIFIER arc %q invalid", p)
		}
		arcs[i] = n
	}
	if arcs[0] > 2 || (arcs[0] < 2 && arcs[1] >= 40) {
		return nil, fmt.Errorf("OBJECT IDENTIFIER %q has invalid first arcs", s)
	}
	first := arcs[0]*40 + arcs[1]
	var out []byte
	out = append(out, encodeBase128(first)...)
	for _, arc := range arcs[2:] {
		out = append(out, encodeBase128(arc)...)
	}
	return out, nil
}

func encodeBase128(n int) []byte {
	if n < 0x80 {
		return []byte{byte(n)}
	}
	var digits []byte
	for {
		digits = append(digits, byte(n&0x7f))
		n >>= 7
		if n == 0 {
			break
		}
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	for i := 0; i < len(digits)-1; i++ {
		digits[i] |= 0x80
	}
	return digits
}

func encodeBitStringContent(v any) ([]byte, error) {
	s, err := asString(v)
	if err != nil {
		return nil, err
	}
	if s == "" {
		return []byte{0x00}, nil
	}
	payload, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("BIT STRING hex: %w", err)
	}
	// Unused bits = 0 (payload is whole bytes from decode).
	out := make([]byte, 1+len(payload))
	out[0] = 0
	copy(out[1:], payload)
	return out, nil
}
