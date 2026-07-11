package decode

import "strings"

func joinFieldPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "." + name
}

func validateFieldPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errEmptyFieldPath()
	}
	if !strings.Contains(path, ".") {
		return errUnqualifiedFieldPath(path)
	}
	return nil
}
