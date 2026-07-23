package encode

func joinFieldPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "." + name
}

func (e *Encoder) lookupFieldSpec(fieldPath, fieldName string) (string, bool) {
	if spec, ok := e.fieldSpecs[fieldPath]; ok {
		return spec, true
	}
	if spec, ok := e.fieldSpecs["*."+fieldName]; ok {
		return spec, true
	}
	return "", false
}
