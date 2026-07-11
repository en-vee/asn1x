package schema_test

import (
	"strings"
	"testing"

	"github.com/en-vee/asn1x/schema"
)

func TestParseIdentifierComponentName(t *testing.T) {
	const src = `M DEFINITIONS ::= BEGIN
T ::= SEQUENCE {
  identifier INTEGER
}
END`

	_, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestParseInformationObjectReference(t *testing.T) {
	const src = `M DEFINITIONS ::= BEGIN
T ::= SEQUENCE {
  identifier DMI-EXTENSION .&id ( { , ...} ) ,
  information DMI-EXTENSION .&Value ( { , ...} { @.identifier } )
}
END`

	mod, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	typ, ok := mod.Lookup("T")
	if !ok {
		t.Fatal("T not found")
	}
	seq, ok := typ.(schema.SequenceType)
	if !ok {
		t.Fatalf("T = %T", typ)
	}
	if len(seq.Components) != 2 {
		t.Fatalf("components = %d", len(seq.Components))
	}

	idField := seq.Components[0].Type.(schema.ReferenceType)
	if idField.Name != "DMI-EXTENSION.&id" {
		t.Fatalf("id field name = %q", idField.Name)
	}
	if len(idField.Constraints) != 1 {
		t.Fatalf("id constraints = %d", len(idField.Constraints))
	}
	obj, ok := idField.Constraints[0].(schema.ObjectSetConstraint)
	if !ok || len(obj.Braces) != 1 || !obj.Braces[0].Extensible {
		t.Fatalf("unexpected object set constraint: %#v", idField.Constraints[0])
	}

	infoField := seq.Components[1].Type.(schema.ReferenceType)
	if len(infoField.Constraints) != 1 {
		t.Fatalf("info constraints = %d", len(infoField.Constraints))
	}
	infoObj := infoField.Constraints[0].(schema.ObjectSetConstraint)
	if len(infoObj.Braces) != 2 {
		t.Fatalf("info braces = %d", len(infoObj.Braces))
	}
	if infoObj.Braces[1].PropertyRef != "identifier" {
		t.Fatalf("property ref = %q", infoObj.Braces[1].PropertyRef)
	}
}

func TestParseAllBuiltinTypesAndConstraints(t *testing.T) {
	const src = `M DEFINITIONS ::= BEGIN
FixedOctet ::= OCTET STRING ( SIZE( 4 ) )
RangedString ::= IA5String ( SIZE( 1 .. 36 ) )
MultiSize ::= OCTET STRING ( SIZE( 1 .. 20 ) ) ( SIZE( 1 .. 9 ) )
BoundedInt ::= INTEGER ( 0 .. 255 )
NamedInt ::= INTEGER { zero (0), one (1) }
NamedBits ::= BIT STRING { flagA (0), flagB (1) } ( SIZE( 8 ) )
RealVal ::= REAL
ObjectId ::= OBJECT IDENTIFIER
RelativeId ::= RELATIVE-OID
ExternalVal ::= EXTERNAL
EmbeddedVal ::= EMBEDDED PDV
UtcVal ::= UTCTime
GenTimeVal ::= GeneralizedTime
Permitted ::= PrintableString ( FROM("A".."Z") )
END`

	mod, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	fixed, _ := mod.Lookup("FixedOctet")
	size, ok := schema.SizeConstraintFrom(schema.ConstraintsOf(fixed.(schema.OctetStringType)))
	if !ok || !size.Fixed || size.Lower != 4 {
		t.Fatalf("FixedOctet size = %#v", size)
	}

	ranged, _ := mod.Lookup("RangedString")
	size, ok = schema.SizeConstraintFrom(schema.ConstraintsOf(ranged.(schema.StringType)))
	if !ok || size.Fixed || size.Lower != 1 || size.Upper != 36 {
		t.Fatalf("RangedString size = %#v", size)
	}

	multi, _ := mod.Lookup("MultiSize")
	if len(schema.ConstraintsOf(multi.(schema.OctetStringType))) != 2 {
		t.Fatalf("MultiSize constraints = %d", len(schema.ConstraintsOf(multi.(schema.OctetStringType))))
	}

	bounded, _ := mod.Lookup("BoundedInt")
	vr, ok := schema.ValueRangeConstraintFrom(schema.ConstraintsOf(bounded.(schema.IntegerType)))
	if !ok || vr.Lower.Value != 0 || vr.Upper.Value != 255 {
		t.Fatalf("BoundedInt range = %#v", vr)
	}

	bits, _ := mod.Lookup("NamedBits")
	bitType := bits.(schema.BitStringType)
	if bitType.NamedBits["flagA"] != 0 || bitType.NamedBits["flagB"] != 1 {
		t.Fatalf("NamedBits = %#v", bitType.NamedBits)
	}
	size, ok = schema.SizeConstraintFrom(bitType.Constraints)
	if !ok || !size.Fixed || size.Lower != 8 {
		t.Fatalf("NamedBits size = %#v", size)
	}

	if _, ok := mod.Lookup("RealVal"); !ok {
		t.Fatal("RealVal not parsed")
	}
	realType, ok := mod.Lookup("RealVal")
	if !ok {
		t.Fatal("RealVal not found")
	}
	if _, ok := realType.(schema.RealType); !ok {
		t.Fatal("RealVal not parsed as REAL")
	}
	if _, ok := mod.Lookup("ObjectId"); !ok {
		t.Fatal("ObjectId not parsed")
	}
	objectType, ok := mod.Lookup("ObjectId")
	if !ok {
		t.Fatal("ObjectId not found")
	}
	if _, ok := objectType.(schema.ObjectIdentifierType); !ok {
		t.Fatal("ObjectId not parsed as OBJECT IDENTIFIER")
	}
	if _, ok := mod.Lookup("RelativeId"); !ok {
		t.Fatal("RelativeId not parsed")
	}
	relativeType, ok := mod.Lookup("RelativeId")
	if !ok {
		t.Fatal("RelativeId not found")
	}
	if _, ok := relativeType.(schema.RelativeOIDType); !ok {
		t.Fatal("RelativeId not parsed as RELATIVE-OID")
	}
	if _, ok := mod.Lookup("ExternalVal"); !ok {
		t.Fatal("ExternalVal not parsed")
	}
	externalType, ok := mod.Lookup("ExternalVal")
	if !ok {
		t.Fatal("ExternalVal not found")
	}
	if _, ok := externalType.(schema.ExternalType); !ok {
		t.Fatal("ExternalVal not parsed as EXTERNAL")
	}
	if _, ok := mod.Lookup("EmbeddedVal"); !ok {
		t.Fatal("EmbeddedVal not parsed")
	}
	embeddedType, ok := mod.Lookup("EmbeddedVal")
	if !ok {
		t.Fatal("EmbeddedVal not found")
	}
	if _, ok := embeddedType.(schema.EmbeddedPDVType); !ok {
		t.Fatal("EmbeddedVal not parsed as EMBEDDED PDV")
	}
	if _, ok := mod.Lookup("UtcVal"); !ok {
		t.Fatal("UtcVal not parsed")
	}
	utcType, ok := mod.Lookup("UtcVal")
	if !ok {
		t.Fatal("UtcVal not found")
	}
	if _, ok := utcType.(schema.UTCTimeType); !ok {
		t.Fatal("UtcVal not parsed as UTCTime")
	}
	if _, ok := mod.Lookup("GenTimeVal"); !ok {
		t.Fatal("GenTimeVal not parsed")
	}
	genType, ok := mod.Lookup("GenTimeVal")
	if !ok {
		t.Fatal("GenTimeVal not found")
	}
	if _, ok := genType.(schema.GeneralizedTimeType); !ok {
		t.Fatal("GenTimeVal not parsed as GeneralizedTime")
	}

	permitted, _ := mod.Lookup("Permitted")
	from, ok := schema.ConstraintsOf(permitted.(schema.StringType))[0].(schema.PermittedAlphabetConstraint)
	if !ok || from.Lower != "A" || from.Upper != "Z" {
		t.Fatalf("Permitted FROM = %#v", from)
	}
}
