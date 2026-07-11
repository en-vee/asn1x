package decode_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/decode"
	"github.com/en-vee/asn1x/schema"
)

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
		t.Fatal(err)
	}
	_, rest, err := ber.ReadOne(data)
	if err != nil {
		t.Fatal(err)
	}
	return data[:len(data)-len(rest)]
}

func TestDecodeCHFSampleFirstRecord(t *testing.T) {
	mod := loadCHFSchema(t)
	record := firstRecordBytes(t, "../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")

	dec := decode.NewDecoder(mod)
	val, err := dec.DecodeBytes("CHFRecord", record)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}

	root, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("root type = %T, want map[string]any", val)
	}
	cfr, ok := root["chargingFunctionRecord"].(map[string]any)
	if !ok {
		t.Fatalf("chargingFunctionRecord = %#v", root)
	}
	if cfr["recordType"] != "chargingFunctionRecord" {
		t.Fatalf("recordType = %#v, want chargingFunctionRecord", cfr["recordType"])
	}
	if cfr["recordingNetworkFunctionID"] != "cd2c0738-631f-4037-a90b-1f1617eee07e" {
		t.Fatalf("recordingNetworkFunctionID = %#v", cfr["recordingNetworkFunctionID"])
	}

	sub, ok := cfr["subscriberIdentifier"].(map[string]any)
	if !ok {
		t.Fatalf("subscriberIdentifier = %#v", cfr["subscriberIdentifier"])
	}
	if sub["subscriptionIDType"] != "eND-USER-IMSI" {
		t.Fatalf("subscriptionIDType = %#v", sub["subscriptionIDType"])
	}
	if sub["subscriptionIDData"] != "505038309039675" {
		t.Fatalf("subscriptionIDData = %#v", sub["subscriptionIDData"])
	}

	nf, ok := cfr["nFunctionConsumerInformation"].(map[string]any)
	if !ok {
		t.Fatalf("nFunctionConsumerInformation = %#v", cfr["nFunctionConsumerInformation"])
	}
	if nf["networkFunctionality"] != "sMF" {
		t.Fatalf("networkFunctionality = %#v", nf["networkFunctionality"])
	}
	if cfr["recordOpeningTime"] != "2022-04-29T01:11:39Z" {
		t.Fatalf("recordOpeningTime = %#v", cfr["recordOpeningTime"])
	}

	jsonBytes, err := decode.ToJSON(val)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}
	if !json.Valid(jsonBytes) {
		t.Fatalf("invalid JSON output")
	}
	t.Logf("first record JSON (%d bytes)", len(jsonBytes))
}

func TestDecodeCHFSampleMultipleRecords(t *testing.T) {
	mod := loadCHFSchema(t)
	data, err := os.ReadFile("../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")
	if err != nil {
		t.Fatal(err)
	}

	dec := decode.NewDecoder(mod)
	for i := 0; i < 3; i++ {
		if len(data) == 0 {
			t.Fatalf("expected record %d, got EOF", i)
		}
		_, data, err = dec.DecodeNext("CHFRecord", data)
		if err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	if len(data) == 0 {
		t.Fatal("expected remaining bytes after 3 records")
	}
}
