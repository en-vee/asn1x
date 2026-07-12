package decode

import (
	"fmt"
	"strconv"
	"strings"
)

// PathFilter matches decoded JSON values at a dotted field path.
type PathFilter struct {
	Path  string
	Value string
}

// ParsePathFilter parses "field.path==value". The value may contain '=' characters.
func ParsePathFilter(expr string) (PathFilter, error) {
	expr = strings.TrimSpace(expr)
	path, value, ok := strings.Cut(expr, "==")
	if !ok || strings.TrimSpace(path) == "" {
		return PathFilter{}, fmt.Errorf("path filter %q: expected field.path==value", expr)
	}
	return PathFilter{
		Path:  strings.TrimSpace(path),
		Value: value,
	}, nil
}

// Match reports whether data contains the filter path with the expected value.
// Arrays along the path are searched: a match in any element satisfies the filter.
func (f PathFilter) Match(data any) bool {
	segments := strings.Split(f.Path, ".")
	if len(segments) == 0 || segments[0] == "" {
		return false
	}
	return matchAt(data, segments, f.Value)
}

func matchAt(data any, segments []string, want string) bool {
	switch cur := data.(type) {
	case []any:
		for _, item := range cur {
			if matchAt(item, segments, want) {
				return true
			}
		}
		return false
	case map[string]any:
		val, ok := cur[segments[0]]
		if !ok {
			return false
		}
		if len(segments) == 1 {
			return valuesEqual(val, want)
		}
		return matchAt(val, segments[1:], want)
	default:
		return false
	}
}

func valuesEqual(got any, want string) bool {
	switch v := got.(type) {
	case string:
		return v == want
	case bool:
		switch want {
		case "true":
			return v
		case "false":
			return !v
		}
	case float64:
		if n, err := strconv.ParseFloat(want, 64); err == nil {
			return v == n
		}
	case int:
		if n, err := strconv.ParseInt(want, 10, 64); err == nil {
			return int64(v) == n
		}
	case int64:
		if n, err := strconv.ParseInt(want, 10, 64); err == nil {
			return v == n
		}
	case uint64:
		if n, err := strconv.ParseUint(want, 10, 64); err == nil {
			return v == n
		}
	case map[string]any:
		return fmt.Sprint(v) == want
	}
	return fmt.Sprint(got) == want
}
