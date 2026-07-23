package ber_test

import (
	"encoding/hex"
	"testing"

	"github.com/en-vee/asn1x/ber"
)

func TestDumpOneInteger(t *testing.T) {
	// INTEGER 200 → 02 02 00 c8
	raw, err := hex.DecodeString("020200c8")
	if err != nil {
		t.Fatal(err)
	}
	node, rest, err := ber.DumpOne(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 0 {
		t.Fatalf("rest = %x", rest)
	}
	if node["tag"] != 2 || node["class"] != "UNIVERSAL" || node["constructed"] != false {
		t.Fatalf("node meta = %#v", node)
	}
	if node["length"] != 2 {
		t.Fatalf("length = %#v", node["length"])
	}
	if node["value"] != int64(200) {
		t.Fatalf("value = %#v", node["value"])
	}
}

func TestDumpOneConstructedWrapped(t *testing.T) {
	// SEQUENCE { INTEGER 7, OCTET STRING "hi" }
	// 30 07  02 01 07  04 02 68 69
	raw, err := hex.DecodeString("300702010704026869")
	if err != nil {
		t.Fatal(err)
	}
	node, _, err := ber.DumpOne(raw)
	if err != nil {
		t.Fatal(err)
	}
	if node["tag"] != 16 || node["constructed"] != true {
		t.Fatalf("node meta = %#v", node)
	}
	children, ok := node["value"].([]any)
	if !ok || len(children) != 2 {
		t.Fatalf("value = %#v", node["value"])
	}
	first := children[0].(ber.Node)
	if first["value"] != int64(7) {
		t.Fatalf("first value = %#v", first["value"])
	}
	second := children[1].(ber.Node)
	if second["value"] != "hi" {
		t.Fatalf("second value = %#v", second["value"])
	}
}

func TestDumpOneContextTag(t *testing.T) {
	// [1] IMPLICIT INTEGER 200 → 81 02 00 c8 (primitive context)
	raw, err := hex.DecodeString("810200c8")
	if err != nil {
		t.Fatal(err)
	}
	node, _, err := ber.DumpOne(raw)
	if err != nil {
		t.Fatal(err)
	}
	if node["tag"] != 1 || node["class"] != "CONTEXT" {
		t.Fatalf("node meta = %#v", node)
	}
	// Non-universal primitives stay as hex/text, not typed integer.
	if node["value"] != "00c8" {
		t.Fatalf("value = %#v", node["value"])
	}
}

func TestDumpAllConcatenated(t *testing.T) {
	raw, err := hex.DecodeString("020107020102")
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := ber.DumpAll(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len = %d", len(nodes))
	}
	if nodes[0]["value"] != int64(7) || nodes[1]["value"] != int64(2) {
		t.Fatalf("nodes = %#v", nodes)
	}
}
