package decode

import "fmt"

func errEmptyFieldPath() error {
	return fmt.Errorf("decode specs: fieldPath is required")
}

func errUnqualifiedFieldPath(path string) error {
	return fmt.Errorf("decode specs: fieldPath must be qualified (e.g. parent.child), got %q", path)
}
