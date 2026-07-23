package asn1x_test

import (
	"os"
	"testing"

	"github.com/en-vee/asn1x"
)

func TestPublicEncodeAPI(t *testing.T) {
	f, err := os.Open("schema/testdata/simple.asn")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	schema, err := asn1x.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	jsonBytes := []byte(`{"call":{"recordType":"moCallRecord","duration":3,"connected":true}}`)
	berBytes, err := asn1x.EncodeJSON(schema, "Record", jsonBytes)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}

	dec := asn1x.NewDecoder(schema)
	val, err := dec.DecodeBytes("Record", berBytes)
	if err != nil {
		t.Fatalf("DecodeBytes() error = %v", err)
	}
	call := val.(map[string]any)["call"].(map[string]any)
	if call["recordType"] != "moCallRecord" {
		t.Fatalf("recordType = %#v", call["recordType"])
	}
}
