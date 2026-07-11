package decode

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/schema"
)

// Decoder decodes BER data using an ASN.1 schema.
type Decoder struct {
	schema     *schema.Schema
	fieldSpecs map[string]string
}

// NewDecoder returns a decoder for s.
func NewDecoder(s *schema.Schema) *Decoder {
	return NewDecoderWithOptions(s, DecodeOptions{})
}

// Decode reads one top-level BER value of rootType from r.
func (d *Decoder) Decode(rootType string, r io.Reader) (any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return d.DecodeBytes(rootType, data)
}

// DecodeBytes decodes one top-level BER value from data.
func (d *Decoder) DecodeBytes(rootType string, data []byte) (any, error) {
	tlv, _, err := ber.ReadOne(data)
	if err != nil {
		return nil, err
	}
	typ, ok := d.schema.Lookup(rootType)
	if !ok {
		return nil, fmt.Errorf("decode: unknown root type %q", rootType)
	}
	return d.decodeTypeWithTLV(typ, tlv, "")
}

// DecodeNext decodes the next top-level BER value from data and returns the remainder.
func (d *Decoder) DecodeNext(rootType string, data []byte) (any, []byte, error) {
	tlv, rest, err := ber.ReadOne(data)
	if err != nil {
		return nil, data, err
	}
	typ, ok := d.schema.Lookup(rootType)
	if !ok {
		return nil, data, fmt.Errorf("decode: unknown root type %q", rootType)
	}
	val, err := d.decodeTypeWithTLV(typ, tlv, "")
	if err != nil {
		return nil, data, err
	}
	return val, rest, nil
}

func (d *Decoder) resolve(t schema.Type) schema.Type {
	for {
		ref, ok := t.(schema.ReferenceType)
		if !ok {
			return t
		}
		if strings.Contains(ref.Name, ".&") {
			return t
		}
		next, ok := d.schema.Lookup(ref.Name)
		if !ok {
			return t
		}
		t = next
	}
}

func (d *Decoder) decodeTypeWithTLV(t schema.Type, tlv ber.TLV, path string) (any, error) {
	t = d.resolve(t)
	switch typ := t.(type) {
	case schema.ChoiceType:
		return d.decodeChoice(typ, tlv, path)
	case schema.SetType:
		return d.decodeComponentList(typ.Components, typ.Extensible, tlv.Value, path)
	case schema.SequenceType:
		return d.decodeComponentList(typ.Components, typ.Extensible, tlv.Value, path)
	case schema.SequenceOfType:
		return d.decodeSequenceOf(typ, tlv.Value, path)
	case schema.SetOfType:
		return d.decodeSetOf(typ, tlv.Value, path)
	default:
		return d.decodePrimitive(t, tlv.Value)
	}
}

func (d *Decoder) decodeChoice(choice schema.ChoiceType, tlv ber.TLV, path string) (any, error) {
	comp, ok := d.matchChoiceArm(choice.Components, tlv.Tag)
	if !ok {
		return nil, fmt.Errorf("decode: no CHOICE alternative for tag class=%d number=%d", tlv.Tag.Class, tlv.Tag.Number)
	}
	val, err := d.decodeComponent(*comp, tlv, path)
	if err != nil {
		return nil, err
	}
	return map[string]any{comp.Name: val}, nil
}

func (d *Decoder) matchChoiceArm(components []schema.Component, tag ber.Tag) (*schema.Component, bool) {
	for i := range components {
		comp := &components[i]
		if d.tagMatchesComponent(*comp, tag) {
			return comp, true
		}
	}
	return nil, false
}

func (d *Decoder) tagMatchesComponent(comp schema.Component, tag ber.Tag) bool {
	expected := d.expectedTag(comp)
	if tagsEqual(expected, tag) {
		return true
	}
	typ := d.resolve(comp.Type)
	if choice, ok := typ.(schema.ChoiceType); ok && comp.Tag == nil {
		_, ok := d.matchChoiceArm(choice.Components, tag)
		return ok
	}
	return false
}

func (d *Decoder) decodeComponent(comp schema.Component, tlv ber.TLV, path string) (any, error) {
	fieldPath := joinFieldPath(path, comp.Name)
	if specType, ok := d.fieldSpecs[fieldPath]; ok {
		content, err := d.componentContent(comp, tlv)
		if err != nil {
			return nil, fmt.Errorf("decode: field %q: %w", fieldPath, err)
		}
		val, err := decodeWithSpecType(specType, content)
		if err != nil {
			return nil, fmt.Errorf("decode: field %q: %w", fieldPath, err)
		}
		return val, nil
	}

	typ := d.resolve(comp.Type)
	if comp.Tag != nil && !comp.Implicit {
		inner, _, err := ber.ReadOne(tlv.Value)
		if err != nil {
			return nil, fmt.Errorf("decode: field %q: %w", fieldPath, err)
		}
		return d.decodeTypeWithTLV(typ, inner, fieldPath)
	}
	return d.decodeTypeWithTLV(typ, tlv, fieldPath)
}

func (d *Decoder) componentContent(comp schema.Component, tlv ber.TLV) ([]byte, error) {
	if comp.Tag != nil && !comp.Implicit {
		inner, _, err := ber.ReadOne(tlv.Value)
		if err != nil {
			return nil, err
		}
		return inner.Value, nil
	}
	return tlv.Value, nil
}

func (d *Decoder) decodeComponentList(components []schema.Component, extensible bool, content []byte, path string) (map[string]any, error) {
	out := make(map[string]any)
	r := ber.NewReader(content)
	for r.Remaining() > 0 {
		tlv, err := r.Read()
		if err != nil {
			return nil, err
		}
		comp, ok := d.matchComponent(components, tlv.Tag)
		if !ok {
			if extensible {
				continue
			}
			return nil, fmt.Errorf("decode: unknown component tag class=%d number=%d", tlv.Tag.Class, tlv.Tag.Number)
		}
		val, err := d.decodeComponent(*comp, tlv, path)
		if err != nil {
			return nil, fmt.Errorf("decode: component %q: %w", comp.Name, err)
		}
		out[comp.Name] = val
	}
	return out, nil
}

func (d *Decoder) decodeSequenceOf(seqOf schema.SequenceOfType, content []byte, path string) ([]any, error) {
	var out []any
	r := ber.NewReader(content)
	elem := d.resolve(seqOf.Element)
	for r.Remaining() > 0 {
		tlv, err := r.Read()
		if err != nil {
			return nil, err
		}
		val, err := d.decodeTypeWithTLV(elem, tlv, path)
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}

func (d *Decoder) decodeSetOf(setOf schema.SetOfType, content []byte, path string) ([]any, error) {
	var out []any
	r := ber.NewReader(content)
	elem := d.resolve(setOf.Element)
	for r.Remaining() > 0 {
		tlv, err := r.Read()
		if err != nil {
			return nil, err
		}
		val, err := d.decodeTypeWithTLV(elem, tlv, path)
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}

func (d *Decoder) decodePrimitive(t schema.Type, content []byte) (any, error) {
	t = d.resolve(t)
	switch typ := t.(type) {
	case schema.IntegerType:
		n, err := ber.DecodeInteger(content)
		if err != nil {
			return nil, err
		}
		if name := namedNumber(typ.NamedNumbers, int(n)); name != "" {
			return name, nil
		}
		return n, nil
	case schema.EnumeratedType:
		n, err := ber.DecodeInteger(content)
		if err != nil {
			return nil, err
		}
		if name := namedNumber(typ.Values, int(n)); name != "" {
			return name, nil
		}
		return n, nil
	case schema.BooleanType:
		return ber.DecodeBoolean(content)
	case schema.NullType:
		return nil, nil
	case schema.OctetStringType:
		return decodeOctetString(content), nil
	case schema.StringType:
		return string(content), nil
	case schema.BitStringType:
		if len(content) == 0 {
			return "", nil
		}
		unused := content[0]
		if unused > 7 {
			return nil, fmt.Errorf("decode: invalid bit string unused bits %d", unused)
		}
		return hex.EncodeToString(content[1:]), nil
	case schema.ObjectIdentifierType:
		return decodeObjectIdentifier(content)
	case schema.UTCTimeType, schema.GeneralizedTimeType:
		return string(content), nil
	case schema.ReferenceType:
		if strings.Contains(typ.Name, ".&") {
			if len(content) == 0 {
				return nil, nil
			}
			return hex.EncodeToString(content), nil
		}
		return nil, fmt.Errorf("decode: unresolved type reference %q", typ.Name)
	default:
		if len(content) == 0 {
			return nil, nil
		}
		return hex.EncodeToString(content), nil
	}
}

func (d *Decoder) matchComponent(components []schema.Component, tag ber.Tag) (*schema.Component, bool) {
	for i := range components {
		comp := &components[i]
		expected := d.expectedTag(*comp)
		if tagsEqual(expected, tag) {
			return comp, true
		}
	}
	return nil, false
}

func (d *Decoder) expectedTag(comp schema.Component) ber.Tag {
	typ := d.resolve(comp.Type)
	if comp.Tag != nil {
		return ber.Tag{
			Class:       ber.Class(comp.Tag.Class),
			Number:      comp.Tag.Number,
			Constructed: isConstructedType(typ),
		}
	}
	return universalTag(typ)
}

func isConstructedType(t schema.Type) bool {
	switch t.(type) {
	case schema.SetType, schema.SequenceType, schema.ChoiceType, schema.SequenceOfType, schema.SetOfType:
		return true
	case schema.ReferenceType:
		return false
	default:
		return false
	}
}

func universalTag(t schema.Type) ber.Tag {
	switch typ := t.(type) {
	case schema.IntegerType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 2, Constructed: false}
	case schema.EnumeratedType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 10, Constructed: false}
	case schema.BooleanType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 1, Constructed: false}
	case schema.NullType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 5, Constructed: false}
	case schema.OctetStringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 4, Constructed: false}
	case schema.BitStringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 3, Constructed: false}
	case schema.ObjectIdentifierType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 6, Constructed: false}
	case schema.StringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: stringTypeTag(typ.Name), Constructed: false}
	case schema.SequenceType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 16, Constructed: true}
	case schema.SetType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 17, Constructed: true}
	case schema.SequenceOfType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 16, Constructed: true}
	case schema.SetOfType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 17, Constructed: true}
	case schema.ChoiceType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 0, Constructed: true}
	case schema.UTCTimeType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 23, Constructed: false}
	case schema.GeneralizedTimeType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 24, Constructed: false}
	default:
		return ber.Tag{Class: ber.ClassUniversal, Number: 4, Constructed: false}
	}
}

func stringTypeTag(name string) int {
	switch name {
	case "UTF8String":
		return 12
	case "NumericString":
		return 18
	case "PrintableString":
		return 19
	case "IA5String":
		return 22
	case "VisibleString", "ISO646String":
		return 26
	case "GeneralizedTime":
		return 24
	default:
		return 12
	}
}

func tagsEqual(a, b ber.Tag) bool {
	return a.Class == b.Class && a.Number == b.Number && a.Constructed == b.Constructed
}

func namedNumber(values map[string]int, n int) string {
	for name, v := range values {
		if v == n {
			return name
		}
	}
	return ""
}

func decodeObjectIdentifier(content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("decode: empty object identifier")
	}
	first := int(content[0])
	parts := []string{fmt.Sprintf("%d", first/40), fmt.Sprintf("%d", first%40)}
	var v int
	for _, b := range content[1:] {
		v = (v << 7) | int(b&0x7f)
		if b&0x80 == 0 {
			parts = append(parts, fmt.Sprintf("%d", v))
			v = 0
		}
	}
	return strings.Join(parts, "."), nil
}
