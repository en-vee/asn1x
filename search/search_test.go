package search_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/en-vee/asn1x/search"
)

func TestSearcherGrepRawBERFile(t *testing.T) {
	sample := filepath.Join("..", "sample-asn1-files", "vvsl22183_-_87150.20220429_._1113+1000.asn1")
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample file not available")
	}

	searcher, err := search.NewSearcher(search.Config{
		SchemaPath: filepath.Join("..", "schema", "testdata", "CHFChargingDataTypes.EXP"),
		RootType:   "CHFRecord",
		FileHeader: false,
		CDRHeader:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searcher.Grep(search.GrepOptions{
		Path:     sample,
		JSONPath: "chargingFunctionRecord.recordType==chargingFunctionRecord",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].RecordNum != 1 {
		t.Fatalf("record number = %d, want 1", matches[0].RecordNum)
	}
}

func TestSearcherGrepNoMatches(t *testing.T) {
	sample := filepath.Join("..", "sample-asn1-files", "vvsl22183_-_87150.20220429_._1113+1000.asn1")
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample file not available")
	}

	searcher, err := search.NewSearcher(search.Config{
		SchemaPath: filepath.Join("..", "schema", "testdata", "CHFChargingDataTypes.EXP"),
		RootType:   "CHFRecord",
		FileHeader: false,
		CDRHeader:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searcher.Grep(search.GrepOptions{
		Path:     sample,
		JSONPath: "chargingFunctionRecord.recordType==does-not-exist",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("matches = %d, want 0", len(matches))
	}
}

func TestSearcherGrepIncludeRecords(t *testing.T) {
	sample := filepath.Join("..", "sample-asn1-files", "vvsl22183_-_87150.20220429_._1113+1000.asn1")
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample file not available")
	}

	searcher, err := search.NewSearcher(search.Config{
		SchemaPath: filepath.Join("..", "schema", "testdata", "CHFChargingDataTypes.EXP"),
		RootType:   "CHFRecord",
		FileHeader: false,
		CDRHeader:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	matches, err := searcher.Grep(search.GrepOptions{
		Path:           sample,
		JSONPath:       "chargingFunctionRecord.recordType==chargingFunctionRecord",
		IncludeRecords: true,
		Limit:          1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("matches = %d, want 1", len(matches))
	}
	if matches[0].Record == nil {
		t.Fatal("expected decoded record payload")
	}
}
