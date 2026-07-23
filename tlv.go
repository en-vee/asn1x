package asn1x

import "github.com/en-vee/asn1x/ber"

// TLVNode is a JSON-friendly wrapped BER TLV element (anonymous dump).
type TLVNode = ber.Node

// DumpTLV converts one BER TLV into a wrapped tag/length/value JSON tree.
// No schema is required; constructed values nest as arrays of child nodes.
func DumpTLV(data []byte) (TLVNode, error) {
	node, _, err := ber.DumpOne(data)
	return node, err
}

// DumpTLVNext converts the next BER TLV into a wrapped node and returns the remainder.
func DumpTLVNext(data []byte) (TLVNode, []byte, error) {
	return ber.DumpOne(data)
}

// DumpTLVAll converts consecutive top-level BER TLVs into wrapped nodes.
func DumpTLVAll(data []byte) ([]TLVNode, error) {
	return ber.DumpAll(data)
}
