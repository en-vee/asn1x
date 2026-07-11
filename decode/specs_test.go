package decode_test

import (
	"os"
	"strings"
	"testing"

	"github.com/en-vee/asn1x/decode"
)

func TestLoadFieldSpecs(t *testing.T) {
	f, err := os.Open("testdata/chf-decode-specs.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	specs, err := decode.LoadFieldSpecs(f)
	if err != nil {
		t.Fatal(err)
	}
	if specs["chargingFunctionRecord.recordOpeningTime"] != "TimeStamp" {
		t.Fatalf("recordOpeningTime path = %q", specs["chargingFunctionRecord.recordOpeningTime"])
	}
	if len(specs) < 5 {
		t.Fatalf("specs = %#v", specs)
	}
}

func TestLoadFieldSpecsRequiresQualifiedPath(t *testing.T) {
	_, err := decode.LoadFieldSpecs(strings.NewReader(`
asn1x:
  decodeSpecs:
    - fieldPath: recordOpeningTime
      asn1DataType: GeneralizedTime
`))
	if err == nil {
		t.Fatal("expected error for unqualified fieldPath")
	}
}

func TestDecodeWithFieldSpecs(t *testing.T) {
	mod := loadCHFSchema(t)
	record := firstRecordBytes(t, "../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")

	specsYAML := `
asn1x:
  decodeSpecs:
    - fieldPath: chargingFunctionRecord.recordOpeningTime
      asn1DataType: UTCTime
`
	specs, err := decode.LoadFieldSpecs(strings.NewReader(specsYAML))
	if err != nil {
		t.Fatal(err)
	}

	dec := decode.NewDecoderWithOptions(mod, decode.DecodeOptions{FieldSpecs: specs})
	val, err := dec.DecodeBytes("CHFRecord", record)
	if err != nil {
		t.Fatal(err)
	}

	cfr := val.(map[string]any)["chargingFunctionRecord"].(map[string]any)
	if cfr["recordOpeningTime"] != "2022-04-29T01:11:39Z" {
		t.Fatalf("recordOpeningTime = %#v", cfr["recordOpeningTime"])
	}
}

func TestDecodeWithFieldSpecsNestedPath(t *testing.T) {
	mod := loadCHFSchema(t)
	record := firstRecordBytes(t, "../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")

	specsYAML := `
asn1x:
  decodeSpecs:
    - fieldPath: chargingFunctionRecord.listOfMultipleUnitUsage.usedUnitContainers.pDUContainerInformation.timeOfFirstUsage
      asn1DataType: GeneralizedTime
`
	specs, err := decode.LoadFieldSpecs(strings.NewReader(specsYAML))
	if err != nil {
		t.Fatal(err)
	}

	dec := decode.NewDecoderWithOptions(mod, decode.DecodeOptions{FieldSpecs: specs})
	val, err := dec.DecodeBytes("CHFRecord", record)
	if err != nil {
		t.Fatal(err)
	}

	cfr := val.(map[string]any)["chargingFunctionRecord"].(map[string]any)
	usage := cfr["listOfMultipleUnitUsage"].([]any)[0].(map[string]any)
	container := usage["usedUnitContainers"].([]any)[0].(map[string]any)
	pdu := container["pDUContainerInformation"].(map[string]any)
	if pdu["timeOfFirstUsage"] != "2022-04-29T01:04:42Z" {
		t.Fatalf("timeOfFirstUsage = %#v", pdu["timeOfFirstUsage"])
	}
}
