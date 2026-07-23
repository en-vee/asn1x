package cli

import (
	"reflect"
	"testing"
)

func TestParseEncodeRecordsSingleObject(t *testing.T) {
	records, err := parseEncodeRecords([]byte(`{"call":{"recordType":"moCallRecord"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}
}

func TestParseEncodeRecordsArray(t *testing.T) {
	records, err := parseEncodeRecords([]byte(`[
		{"call":{"recordType":"moCallRecord"}},
		{"sms":{"messageId":"hi"}}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("len = %d, want 2", len(records))
	}
}

func TestParseEncodeRecordsJSONLines(t *testing.T) {
	records, err := parseEncodeRecords([]byte("{\"a\":1}\n{\"b\":2}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("len = %d, want 2", len(records))
	}
	if !reflect.DeepEqual(records[0], map[string]any{"a": float64(1)}) {
		t.Fatalf("record[0] = %#v", records[0])
	}
}

func TestParseEncodeRecordsConcatenatedObjects(t *testing.T) {
	records, err := parseEncodeRecords([]byte(`{"a":1}{"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("len = %d, want 2", len(records))
	}
}

func TestParseEncodeRecordsRejectsNonObjectArrayElement(t *testing.T) {
	_, err := parseEncodeRecords([]byte(`[1, 2]`))
	if err == nil {
		t.Fatal("expected error")
	}
}
