package decode

import "testing"

func TestJoinFieldPath(t *testing.T) {
	if got := joinFieldPath("", "chargingFunctionRecord"); got != "chargingFunctionRecord" {
		t.Fatalf("got %q", got)
	}
	if got := joinFieldPath("chargingFunctionRecord", "recordOpeningTime"); got != "chargingFunctionRecord.recordOpeningTime" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateFieldPath(t *testing.T) {
	if err := validateFieldPath("a.b"); err != nil {
		t.Fatal(err)
	}
	if err := validateFieldPath("recordOpeningTime"); err == nil {
		t.Fatal("expected error for unqualified path")
	}
}
