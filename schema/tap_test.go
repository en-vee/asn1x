package schema_test

import (
	"os"
	"strings"
	"testing"

	"github.com/en-vee/asn1x/schema"
)

func TestParseImplicitTagsModule(t *testing.T) {
	const src = `TapMod DEFINITIONS IMPLICIT TAGS ::=
BEGIN
TransferBatch ::= [APPLICATION 1] SEQUENCE {
    batchControlInfo BatchControlInfo OPTIONAL,
    ...
}
BatchControlInfo ::= [APPLICATION 4] SEQUENCE {
    sender Sender OPTIONAL
}
Sender ::= [APPLICATION 196] OCTET STRING
BasicServiceCode ::= [APPLICATION 426] CHOICE {
    bearerServiceCode BearerServiceCode,
    ...
}
BearerServiceCode ::= [APPLICATION 40] OCTET STRING (SIZE(2))
END
`
	mod, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if mod.TagDefault != schema.TagDefaultImplicit {
		t.Fatalf("TagDefault = %v, want Implicit", mod.TagDefault)
	}

	tb, ok := mod.Lookup("TransferBatch")
	if !ok {
		t.Fatal("TransferBatch missing")
	}
	tagged, ok := tb.(schema.TaggedType)
	if !ok {
		t.Fatalf("TransferBatch type = %T", tb)
	}
	if tagged.Tag.Class != schema.TagClassApplication || tagged.Tag.Number != 1 {
		t.Fatalf("TransferBatch tag = %+v", tagged.Tag)
	}
	if !tagged.Implicit {
		t.Fatal("TransferBatch should be IMPLICIT under IMPLICIT TAGS")
	}
	seq, ok := tagged.Type.(schema.SequenceType)
	if !ok || len(seq.Components) != 1 {
		t.Fatalf("TransferBatch inner = %#v", tagged.Type)
	}
	if seq.Components[0].Tag != nil {
		t.Fatalf("untagged component should have nil Tag, got %+v", seq.Components[0].Tag)
	}

	bsc, ok := mod.Lookup("BasicServiceCode")
	if !ok {
		t.Fatal("BasicServiceCode missing")
	}
	bscTagged := bsc.(schema.TaggedType)
	if bscTagged.Implicit {
		t.Fatal("tagged CHOICE must be EXPLICIT per X.680")
	}
}

func TestParseTAP312Schema(t *testing.T) {
	f, err := os.Open("testdata/TAP3-12.asn")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("Parse TAP3-12.asn error = %v", err)
	}
	if mod.ModuleName != "TAP-0312" {
		t.Fatalf("ModuleName = %q", mod.ModuleName)
	}
	if mod.TagDefault != schema.TagDefaultImplicit {
		t.Fatalf("TagDefault = %v, want Implicit", mod.TagDefault)
	}
	if len(mod.Types) < 100 {
		t.Fatalf("types = %d, want a large TAP module", len(mod.Types))
	}

	for _, name := range []string{
		"DataInterChange",
		"TransferBatch",
		"Notification",
		"CallEventDetail",
		"MobileOriginatedCall",
		"GprsCall",
		"BatchControlInfo",
	} {
		if _, ok := mod.Lookup(name); !ok {
			t.Fatalf("missing type %q", name)
		}
	}

	tb := mod.Types["TransferBatch"].(schema.TaggedType)
	if tb.Tag.Number != 1 || tb.Tag.Class != schema.TagClassApplication || !tb.Implicit {
		t.Fatalf("TransferBatch tagged = %+v implicit=%v", tb.Tag, tb.Implicit)
	}
}
