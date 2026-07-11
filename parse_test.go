package asn1x_test

import (
	"strings"
	"testing"

	"github.com/en-vee/asn1x"
	asn1schema "github.com/en-vee/asn1x/schema"
)

func TestPublicParse(t *testing.T) {
	const src = `Root DEFINITIONS ::= BEGIN
Value ::= INTEGER
END`

	mod, err := asn1x.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if mod.ModuleName != "Root" {
		t.Fatalf("ModuleName = %q, want Root", mod.ModuleName)
	}
	typ, ok := mod.Lookup("Value")
	if !ok {
		t.Fatal("Value type not found")
	}
	if _, ok := typ.(asn1schema.IntegerType); !ok {
		t.Fatalf("Value type = %T, want IntegerType", typ)
	}
}
