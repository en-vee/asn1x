package encode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/schema"
)

// Encoder encodes JSON-friendly Go values to BER using an ASN.1 schema.
type Encoder struct {
	schema     *schema.Schema
	fieldSpecs map[string]string
}

// EncodeJSON unmarshals jsonBytes and encodes the value as rootType.
func (e *Encoder) EncodeJSON(rootType string, jsonBytes []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		return nil, fmt.Errorf("encode: json: %w", err)
	}
	return e.EncodeBytes(rootType, v)
}

// EncodeBytes encodes v as a top-level BER value of rootType.
func (e *Encoder) EncodeBytes(rootType string, v any) ([]byte, error) {
	typ, ok := e.schema.Lookup(rootType)
	if !ok {
		return nil, fmt.Errorf("encode: unknown root type %q", rootType)
	}
	return e.encodeType(typ, v, "")
}

func (e *Encoder) resolve(t schema.Type) schema.Type {
	for {
		ref, ok := t.(schema.ReferenceType)
		if !ok {
			return t
		}
		if strings.Contains(ref.Name, ".&") {
			return t
		}
		next, ok := e.schema.Lookup(ref.Name)
		if !ok {
			return t
		}
		t = next
	}
}

func (e *Encoder) encodeType(t schema.Type, v any, path string) ([]byte, error) {
	t = e.resolve(t)
	if tagged, ok := t.(schema.TaggedType); ok {
		return e.encodeTaggedType(tagged, v, path)
	}
	switch typ := t.(type) {
	case schema.ChoiceType:
		return e.encodeChoice(typ, v, path)
	case schema.SetType:
		content, err := e.encodeComponentList(typ.Components, v, path)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(universalTag(typ), content), nil
	case schema.SequenceType:
		content, err := e.encodeComponentList(typ.Components, v, path)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(universalTag(typ), content), nil
	case schema.SequenceOfType:
		content, err := e.encodeSeqOf(typ.Element, v, path)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(universalTag(typ), content), nil
	case schema.SetOfType:
		content, err := e.encodeSeqOf(typ.Element, v, path)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(universalTag(typ), content), nil
	default:
		content, err := e.encodePrimitiveContent(t, v)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(universalTag(t), content), nil
	}
}

func (e *Encoder) encodeTaggedType(tagged schema.TaggedType, v any, path string) ([]byte, error) {
	innerType := e.resolve(tagged.Type)
	// X.680 requires EXPLICIT tagging for CHOICE even in IMPLICIT TAGS modules.
	explicit := !tagged.Implicit
	if _, ok := e.underlyingType(innerType).(schema.ChoiceType); ok {
		explicit = true
	}
	if !explicit {
		content, err := e.encodeTypeContent(innerType, v, path)
		if err != nil {
			return nil, err
		}
		return ber.EncodeTLV(ber.Tag{
			Class:       ber.Class(tagged.Tag.Class),
			Number:      tagged.Tag.Number,
			Constructed: e.isConstructedType(innerType),
		}, content), nil
	}
	inner, err := e.encodeType(innerType, v, path)
	if err != nil {
		return nil, err
	}
	return ber.EncodeTLV(ber.Tag{
		Class:       ber.Class(tagged.Tag.Class),
		Number:      tagged.Tag.Number,
		Constructed: true,
	}, inner), nil
}

func (e *Encoder) encodeChoice(choice schema.ChoiceType, v any, path string) ([]byte, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("encode: CHOICE at %q must be an object, got %T", pathOrRoot(path), v)
	}
	if len(m) != 1 {
		return nil, fmt.Errorf("encode: CHOICE at %q must have exactly one key, got %d", pathOrRoot(path), len(m))
	}
	var armName string
	var armVal any
	for k, val := range m {
		armName = k
		armVal = val
	}
	comp, ok := findComponent(choice.Components, armName)
	if !ok {
		return nil, fmt.Errorf("encode: unknown CHOICE alternative %q at %q", armName, pathOrRoot(path))
	}
	return e.encodeComponent(comp, armVal, path)
}

func (e *Encoder) encodeComponentList(components []schema.Component, v any, path string) ([]byte, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("encode: SEQUENCE/SET at %q must be an object, got %T", pathOrRoot(path), v)
	}

	var out []byte
	seen := make(map[string]bool, len(m))
	for _, comp := range components {
		val, present := m[comp.Name]
		if !present {
			// Omit absent fields. Many 3GPP schemas leave OPTIONAL off fields that
			// are routinely absent on the wire; encode follows the JSON presence model.
			continue
		}
		seen[comp.Name] = true
		encoded, err := e.encodeComponent(comp, val, path)
		if err != nil {
			return nil, err
		}
		out = append(out, encoded...)
	}
	for name := range m {
		if seen[name] {
			continue
		}
		if strings.HasPrefix(name, "Unknown_") {
			// Extension / unknown fields are not re-encoded in v1.
			continue
		}
		return nil, fmt.Errorf("encode: unknown field %q at %q", name, pathOrRoot(path))
	}
	return out, nil
}

func (e *Encoder) encodeComponent(comp schema.Component, v any, path string) ([]byte, error) {
	fieldPath := joinFieldPath(path, comp.Name)
	typ := e.resolve(comp.Type)

	if specType, ok := e.lookupFieldSpec(fieldPath, comp.Name); ok {
		content, err := encodeWithSpecType(specType, v)
		if err != nil {
			return nil, fmt.Errorf("encode: field %q: %w", fieldPath, err)
		}
		tag := e.expectedTag(comp)
		// Spec overrides produce primitive content; force non-constructed.
		tag.Constructed = false
		if comp.Tag != nil && !comp.Implicit {
			inner := ber.EncodeTLV(universalTag(typ), content)
			return ber.EncodeTLV(ber.Tag{
				Class:       ber.Class(comp.Tag.Class),
				Number:      comp.Tag.Number,
				Constructed: true,
			}, inner), nil
		}
		return ber.EncodeTLV(tag, content), nil
	}

	// Untagged CHOICE: arm tags alone identify the alternative.
	if choice, ok := typ.(schema.ChoiceType); ok && comp.Tag == nil {
		return e.encodeChoice(choice, v, fieldPath)
	}

	if comp.Tag != nil && !comp.Implicit {
		// EXPLICIT: encode the full inner TLV, then wrap with the component tag.
		inner, err := e.encodeType(typ, v, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("encode: field %q: %w", fieldPath, err)
		}
		return ber.EncodeTLV(ber.Tag{
			Class:       ber.Class(comp.Tag.Class),
			Number:      comp.Tag.Number,
			Constructed: true,
		}, inner), nil
	}

	if comp.Tag != nil && comp.Implicit {
		if choice, ok := typ.(schema.ChoiceType); ok {
			// IMPLICIT CHOICE: the chosen arm is encoded under the outer tag.
			// Match decode, which dispatches the outer TLV to the CHOICE type.
			return e.encodeImplicitChoice(choice, comp, v, fieldPath)
		}
		content, err := e.encodeTypeContent(typ, v, fieldPath)
		if err != nil {
			return nil, fmt.Errorf("encode: field %q: %w", fieldPath, err)
		}
		return ber.EncodeTLV(e.expectedTag(comp), content), nil
	}

	// No component tag: encode with universal tag.
	return e.encodeType(typ, v, fieldPath)
}

// encodeImplicitChoice encodes a CHOICE under an IMPLICIT component tag by
// encoding the selected arm's content and applying the outer component tag.
func (e *Encoder) encodeImplicitChoice(choice schema.ChoiceType, outer schema.Component, v any, path string) ([]byte, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("encode: CHOICE at %q must be an object, got %T", path, v)
	}
	if len(m) != 1 {
		return nil, fmt.Errorf("encode: CHOICE at %q must have exactly one key, got %d", path, len(m))
	}
	var armName string
	var armVal any
	for k, val := range m {
		armName = k
		armVal = val
	}
	arm, ok := findComponent(choice.Components, armName)
	if !ok {
		return nil, fmt.Errorf("encode: unknown CHOICE alternative %q at %q", armName, path)
	}
	// Encode arm content as if under the outer IMPLICIT tag.
	tagged := arm
	tagged.Tag = outer.Tag
	tagged.Implicit = true
	return e.encodeComponent(tagged, armVal, path)
}

// encodeTypeContent encodes the value without an outer universal tag for aggregates
// (used for IMPLICIT tagging where the context tag replaces the universal tag).
func (e *Encoder) encodeTypeContent(t schema.Type, v any, path string) ([]byte, error) {
	t = e.resolve(t)
	if tagged, ok := t.(schema.TaggedType); ok {
		// Outer tag already applied by caller; encode the underlying content.
		return e.encodeTypeContent(tagged.Type, v, path)
	}
	switch typ := t.(type) {
	case schema.ChoiceType:
		return e.encodeChoice(typ, v, path)
	case schema.SetType:
		return e.encodeComponentList(typ.Components, v, path)
	case schema.SequenceType:
		return e.encodeComponentList(typ.Components, v, path)
	case schema.SequenceOfType:
		return e.encodeSeqOf(typ.Element, v, path)
	case schema.SetOfType:
		return e.encodeSeqOf(typ.Element, v, path)
	default:
		return e.encodePrimitiveContent(t, v)
	}
}

func (e *Encoder) encodeSeqOf(elem schema.Type, v any, path string) ([]byte, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("encode: SEQUENCE OF/SET OF at %q must be an array, got %T", pathOrRoot(path), v)
	}
	elem = e.resolve(elem)
	var out []byte
	for i, item := range arr {
		encoded, err := e.encodeType(elem, item, path)
		if err != nil {
			return nil, fmt.Errorf("encode: %s[%d]: %w", pathOrRoot(path), i, err)
		}
		out = append(out, encoded...)
	}
	return out, nil
}

func findComponent(components []schema.Component, name string) (schema.Component, bool) {
	for _, comp := range components {
		if comp.Name == name {
			return comp, true
		}
	}
	return schema.Component{}, false
}

func pathOrRoot(path string) string {
	if path == "" {
		return "<root>"
	}
	return path
}
