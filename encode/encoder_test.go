package encode_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/decode"
	"github.com/en-vee/asn1x/encode"
	"github.com/en-vee/asn1x/schema"
)

func loadSimpleSchema(t *testing.T) *schema.Schema {
	t.Helper()
	f, err := os.Open("../schema/testdata/simple.asn")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("schema.Parse() error = %v", err)
	}
	return mod
}

func loadCHFSchema(t *testing.T) *schema.Schema {
	t.Helper()
	f, err := os.Open("../schema/testdata/CHFChargingDataTypes.EXP")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	mod, err := schema.Parse(f)
	if err != nil {
		t.Fatalf("schema.Parse() error = %v", err)
	}
	return mod
}

func firstRecordBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("sample file not available: %v", err)
	}
	_, rest, err := ber.ReadOne(data)
	if err != nil {
		t.Fatal(err)
	}
	return data[:len(data)-len(rest)]
}

func TestEncodeSimpleRecordRoundTrip(t *testing.T) {
	mod := loadSimpleSchema(t)
	enc := encode.NewEncoder(mod)
	dec := decode.NewDecoder(mod)

	val := map[string]any{
		"call": map[string]any{
			"recordType": "moCallRecord",
			"duration":   int64(42),
			"connected":  true,
		},
	}
	berBytes, err := enc.EncodeBytes("Record", val)
	if err != nil {
		t.Fatalf("EncodeBytes() error = %v", err)
	}
	got, err := dec.DecodeBytes("Record", berBytes)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	root := got.(map[string]any)
	call := root["call"].(map[string]any)
	if call["recordType"] != "moCallRecord" {
		t.Fatalf("recordType = %#v", call["recordType"])
	}
	if call["duration"] != int64(42) {
		t.Fatalf("duration = %#v", call["duration"])
	}
	if call["connected"] != true {
		t.Fatalf("connected = %#v", call["connected"])
	}
}

func TestEncodeSimpleRecordOmitOptional(t *testing.T) {
	mod := loadSimpleSchema(t)
	enc := encode.NewEncoder(mod)
	dec := decode.NewDecoder(mod)

	val := map[string]any{
		"sms": map[string]any{
			"messageId": "hi",
		},
	}
	berBytes, err := enc.EncodeBytes("Record", val)
	if err != nil {
		t.Fatalf("EncodeBytes() error = %v", err)
	}
	got, err := dec.DecodeBytes("Record", berBytes)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	root := got.(map[string]any)
	sms := root["sms"].(map[string]any)
	if sms["messageId"] != "hi" {
		t.Fatalf("messageId = %#v", sms["messageId"])
	}
}

func TestEncodeSimpleRecordList(t *testing.T) {
	mod := loadSimpleSchema(t)
	enc := encode.NewEncoder(mod)
	dec := decode.NewDecoder(mod)

	val := []any{
		map[string]any{
			"call": map[string]any{
				"recordType": "mtCallRecord",
				"connected":  false,
			},
		},
		map[string]any{
			"sms": map[string]any{
				"messageId": "abcd",
			},
		},
	}
	berBytes, err := enc.EncodeBytes("RecordList", val)
	if err != nil {
		t.Fatalf("EncodeBytes() error = %v", err)
	}
	got, err := dec.DecodeBytes("RecordList", berBytes)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	arr := got.([]any)
	if len(arr) != 2 {
		t.Fatalf("len = %d", len(arr))
	}
}

func TestEncodeJSONSimple(t *testing.T) {
	mod := loadSimpleSchema(t)
	enc := encode.NewEncoder(mod)
	dec := decode.NewDecoder(mod)

	jsonBytes := []byte(`{"call":{"recordType":"moCallRecord","duration":7,"connected":true}}`)
	berBytes, err := enc.EncodeJSON("Record", jsonBytes)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	got, err := dec.DecodeBytes("Record", berBytes)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	call := got.(map[string]any)["call"].(map[string]any)
	if call["duration"] != int64(7) {
		t.Fatalf("duration = %#v", call["duration"])
	}
}

func TestEncodeCHFSampleRoundTrip(t *testing.T) {
	mod := loadCHFSchema(t)
	record := firstRecordBytes(t, "../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")

	specFile, err := os.Open("../decode/testdata/chf-decode-specs.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer specFile.Close()
	specs, err := decode.LoadFieldSpecs(specFile)
	if err != nil {
		t.Fatal(err)
	}

	dec := decode.NewDecoderWithOptions(mod, decode.DecodeOptions{FieldSpecs: specs})
	enc := encode.NewEncoderWithOptions(mod, encode.EncodeOptions{FieldSpecs: specs})

	val, err := dec.DecodeBytes("CHFRecord", record)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}

	encoded, err := enc.EncodeBytes("CHFRecord", val)
	if err != nil {
		t.Fatalf("EncodeBytes() error = %v", err)
	}

	val2, err := dec.DecodeBytes("CHFRecord", encoded)
	if err != nil {
		t.Fatalf("DecodeBytes(re-encoded) error = %v", err)
	}

	j1, err := json.Marshal(normalizeJSON(val))
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(normalizeJSON(val2))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(j1, j2) {
		t.Fatalf("round-trip JSON mismatch:\nfirst:  %s\nsecond: %s", j1, j2)
	}

	// Spot-check key fields from the original decode test.
	cfr := val2.(map[string]any)["chargingFunctionRecord"].(map[string]any)
	if cfr["recordType"] != "chargingFunctionRecord" {
		t.Fatalf("recordType = %#v", cfr["recordType"])
	}
	if cfr["recordOpeningTime"] != "2022-04-29T01:11:39Z" {
		t.Fatalf("recordOpeningTime = %#v", cfr["recordOpeningTime"])
	}
}

func TestEncodeCHFSampleWithoutSpecs(t *testing.T) {
	mod := loadCHFSchema(t)
	record := firstRecordBytes(t, "../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")

	dec := decode.NewDecoder(mod)
	enc := encode.NewEncoder(mod)

	val, err := dec.DecodeBytes("CHFRecord", record)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	encoded, err := enc.EncodeBytes("CHFRecord", val)
	if err != nil {
		t.Fatalf("EncodeBytes() error = %v", err)
	}
	val2, err := dec.DecodeBytes("CHFRecord", encoded)
	if err != nil {
		t.Fatalf("DecodeBytes(re-encoded) error = %v", err)
	}
	if !reflect.DeepEqual(normalizeJSON(val), normalizeJSON(val2)) {
		t.Fatal("round-trip without specs produced different values")
	}
}

// normalizeJSON converts values to the shapes produced by encoding/json so
// map[string]string and int64 compare equal to their JSON counterparts.
func normalizeJSON(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}
