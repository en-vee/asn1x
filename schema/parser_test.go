package schema_test

import (
	"os"
	"strings"
	"testing"

	"github.com/en-vee/asn1x/schema"
)

func TestParseSimpleModule(t *testing.T) {
	f, err := os.Open("testdata/simple.asn")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if mod.ModuleName != "MyModule" {
		t.Fatalf("ModuleName = %q, want MyModule", mod.ModuleName)
	}

	record, ok := mod.Lookup("Record")
	if !ok {
		t.Fatal("Record type not found")
	}
	choice, ok := record.(schema.ChoiceType)
	if !ok {
		t.Fatalf("Record type kind = %T, want ChoiceType", record)
	}
	if len(choice.Components) != 2 {
		t.Fatalf("Record choice arms = %d, want 2", len(choice.Components))
	}

	call := choice.Components[0]
	if call.Name != "call" || call.Tag == nil || call.Tag.Number != 1 || !call.Implicit {
		t.Fatalf("unexpected call component: %+v", call)
	}

	callSeq, ok := call.Type.(schema.SequenceType)
	if !ok {
		t.Fatalf("call type = %T, want SequenceType", call.Type)
	}
	if len(callSeq.Components) != 3 {
		t.Fatalf("call sequence components = %d, want 3", len(callSeq.Components))
	}

	duration := callSeq.Components[1]
	if !duration.Optional {
		t.Fatal("duration component should be OPTIONAL")
	}

	connected := callSeq.Components[2]
	if connected.Default == nil || connected.Default.Kind != schema.DefaultKindBoolean || !connected.Default.Bool {
		t.Fatalf("connected DEFAULT TRUE not parsed: %+v", connected.Default)
	}

	recordList, ok := mod.Lookup("RecordList")
	if !ok {
		t.Fatal("RecordList type not found")
	}
	seqOf, ok := recordList.(schema.SequenceOfType)
	if !ok {
		t.Fatalf("RecordList type = %T, want SequenceOfType", recordList)
	}
	ref, ok := seqOf.Element.(schema.ReferenceType)
	if !ok || ref.Name != "Record" {
		t.Fatalf("RecordList element = %+v, want reference to Record", seqOf.Element)
	}

	status, ok := mod.Lookup("Status")
	if !ok {
		t.Fatal("Status type not found")
	}
	enum, ok := status.(schema.EnumeratedType)
	if !ok {
		t.Fatalf("Status type = %T, want EnumeratedType", status)
	}
	if enum.Values["active"] != 0 || enum.Values["inactive"] != 1 {
		t.Fatalf("unexpected enum values: %+v", enum.Values)
	}
}

func TestParseModuleWithOID(t *testing.T) {
	const src = `ExampleModule{ itu-t (0) identified-organization (4) etsi (0) } DEFINITIONS ::= BEGIN
Flag ::= BOOLEAN
END`

	mod, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if mod.ModuleName != "ExampleModule" {
		t.Fatalf("ModuleName = %q", mod.ModuleName)
	}
	if mod.ModuleOID == "" {
		t.Fatal("expected ModuleOID to be populated")
	}
}

func TestParseRejectsDuplicateAssignments(t *testing.T) {
	const src = `Dup DEFINITIONS ::= BEGIN
A ::= INTEGER
A ::= BOOLEAN
END`

	_, err := schema.Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for duplicate assignment")
	}
}

func TestLexerSkipsComments(t *testing.T) {
	const src = `Commented DEFINITIONS ::= BEGIN
-- this is a comment
X ::= NULL -- trailing comment
END`

	mod, err := schema.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, ok := mod.Lookup("X"); !ok {
		t.Fatal("type X not found")
	}
}

func TestParseCHFChargingDataTypesEXP(t *testing.T) {
	f, err := os.Open("testdata/CHFChargingDataTypes.EXP")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if mod.ModuleName != "CHFChargingDataTypes" {
		t.Fatalf("ModuleName = %q, want CHFChargingDataTypes", mod.ModuleName)
	}
	if mod.ModuleOID == "" {
		t.Fatal("expected ModuleOID to be populated")
	}

	chfRecord, ok := mod.Lookup("CHFRecord")
	if !ok {
		t.Fatal("CHFRecord type not found")
	}
	if _, ok := chfRecord.(schema.ChoiceType); !ok {
		t.Fatalf("CHFRecord type = %T, want ChoiceType", chfRecord)
	}

	pduSessionID, ok := mod.Lookup("PDUSessionId")
	if !ok {
		t.Fatal("PDUSessionId type not found")
	}
	intType, ok := pduSessionID.(schema.IntegerType)
	if !ok {
		t.Fatalf("PDUSessionId = %T, want IntegerType", pduSessionID)
	}
	vr, ok := schema.ValueRangeConstraintFrom(intType.Constraints)
	if !ok || vr.Lower.Value != 0 || vr.Upper.Value != 255 {
		t.Fatalf("PDUSessionId range = %#v", vr)
	}

	amfID, ok := mod.Lookup("AMFID")
	if !ok {
		t.Fatal("AMFID type not found")
	}
	size, ok := schema.SizeConstraintFrom(schema.ConstraintsOf(amfID.(schema.OctetStringType)))
	if !ok || !size.Fixed || size.Lower != 6 {
		t.Fatalf("AMFID size = %#v", size)
	}

	if len(mod.Types) != 49 {
		t.Fatalf("type assignments = %d, want 49", len(mod.Types))
	}
}

func TestParseGPRSChargingDataTypesASN(t *testing.T) {
	f, err := os.Open("testdata/GPRSChargingDataTypes.asn")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if mod.ModuleName != "GPRSChargingDataTypes" {
		t.Fatalf("ModuleName = %q, want GPRSChargingDataTypes", mod.ModuleName)
	}
	if mod.ModuleOID == "" {
		t.Fatal("expected ModuleOID to be populated")
	}

	gprsRecord, ok := mod.Lookup("GPRSRecord")
	if !ok {
		t.Fatal("GPRSRecord type not found")
	}
	if _, ok := gprsRecord.(schema.ChoiceType); !ok {
		t.Fatalf("GPRSRecord type = %T, want ChoiceType", gprsRecord)
	}

	csgID, ok := mod.Lookup("CSGId")
	if !ok {
		t.Fatal("CSGId type not found")
	}
	size, ok := schema.SizeConstraintFrom(schema.ConstraintsOf(csgID.(schema.OctetStringType)))
	if !ok || !size.Fixed || size.Lower != 4 {
		t.Fatalf("CSGId size = %#v", size)
	}

	accessAvail, ok := mod.Lookup("AccessAvailabilityChangeReason")
	if !ok {
		t.Fatal("AccessAvailabilityChangeReason type not found")
	}
	intType, ok := accessAvail.(schema.IntegerType)
	if !ok {
		t.Fatalf("AccessAvailabilityChangeReason = %T, want IntegerType", accessAvail)
	}
	vr, ok := schema.ValueRangeConstraintFrom(intType.Constraints)
	if !ok || vr.Lower.Value != 0 || vr.Upper.Value != 4294967295 {
		t.Fatalf("AccessAvailabilityChangeReason range = %#v", vr)
	}

	if len(mod.Types) != 96 {
		t.Fatalf("type assignments = %d, want 96", len(mod.Types))
	}
}
