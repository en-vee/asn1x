package ber

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Node is a JSON-friendly wrapped BER TLV element.
// Constructed values are []any of child Nodes (as maps); primitives are
// decoded when the universal tag is known, otherwise hex or text.
type Node map[string]any

// DumpOne reads the first TLV from data and returns a wrapped node tree plus remainder.
func DumpOne(data []byte) (Node, []byte, error) {
	tlv, rest, err := ReadOne(data)
	if err != nil {
		return nil, data, err
	}
	node, err := DumpTLV(tlv)
	if err != nil {
		return nil, data, err
	}
	return node, rest, nil
}

// DumpAll reads consecutive top-level TLVs from data into a slice of wrapped nodes.
func DumpAll(data []byte) ([]Node, error) {
	var out []Node
	for len(data) > 0 {
		node, rest, err := DumpOne(data)
		if err != nil {
			return nil, err
		}
		out = append(out, node)
		data = rest
	}
	return out, nil
}

// DumpTLV converts a parsed TLV into a wrapped JSON-friendly node.
func DumpTLV(tlv TLV) (Node, error) {
	node := Node{
		"tag":         tlv.Tag.Number,
		"class":       className(tlv.Tag.Class),
		"constructed": tlv.Tag.Constructed,
		"length":      len(tlv.Value),
	}
	if tlv.Tag.Constructed {
		children, err := dumpChildren(tlv.Value)
		if err != nil {
			return nil, err
		}
		node["value"] = children
		return node, nil
	}
	node["value"] = dumpPrimitiveValue(tlv.Tag, tlv.Value)
	return node, nil
}

func dumpChildren(content []byte) ([]any, error) {
	var children []any
	r := NewReader(content)
	for r.Remaining() > 0 {
		tlv, err := r.Read()
		if err != nil {
			return nil, err
		}
		node, err := DumpTLV(tlv)
		if err != nil {
			return nil, err
		}
		children = append(children, node)
	}
	if children == nil {
		children = []any{}
	}
	return children, nil
}

func dumpPrimitiveValue(tag Tag, content []byte) any {
	if tag.Class == ClassUniversal {
		switch tag.Number {
		case 1: // BOOLEAN
			v, err := DecodeBoolean(content)
			if err != nil {
				return hex.EncodeToString(content)
			}
			return v
		case 2: // INTEGER
			n, err := DecodeInteger(content)
			if err != nil {
				return hex.EncodeToString(content)
			}
			return n
		case 3: // BIT STRING
			if len(content) == 0 {
				return ""
			}
			return hex.EncodeToString(content[1:])
		case 5: // NULL
			return nil
		case 6: // OBJECT IDENTIFIER
			s, err := decodeObjectIdentifier(content)
			if err != nil {
				return hex.EncodeToString(content)
			}
			return s
		case 4, 12, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 30:
			// OCTET STRING and character/time string types
			return dumpOctets(content)
		}
	}
	return dumpOctets(content)
}

func dumpOctets(content []byte) any {
	if isTextualOctets(content) {
		return string(content)
	}
	return hex.EncodeToString(content)
}

func isTextualOctets(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	if !utf8.Valid(content) {
		return false
	}
	for _, b := range content {
		if b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}

func className(c Class) string {
	switch c {
	case ClassUniversal:
		return "UNIVERSAL"
	case ClassApplication:
		return "APPLICATION"
	case ClassContext:
		return "CONTEXT"
	case ClassPrivate:
		return "PRIVATE"
	default:
		return fmt.Sprintf("CLASS_%d", int(c))
	}
}

func decodeObjectIdentifier(content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("empty object identifier")
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
