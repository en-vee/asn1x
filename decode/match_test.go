package decode

import "testing"

func TestParsePathFilter(t *testing.T) {
	f, err := ParsePathFilter("chargingFunctionRecord.duration==148")
	if err != nil {
		t.Fatal(err)
	}
	if f.Path != "chargingFunctionRecord.duration" || f.Value != "148" {
		t.Fatalf("filter = %#v", f)
	}

	f, err = ParsePathFilter("a.b==+10:00+1")
	if err != nil {
		t.Fatal(err)
	}
	if f.Value != "+10:00+1" {
		t.Fatalf("value = %q", f.Value)
	}
}

func TestPathFilterMatchNestedArrays(t *testing.T) {
	data := map[string]any{
		"chargingFunctionRecord": map[string]any{
			"listOfMultipleUnitUsage": []any{
				map[string]any{
					"usedUnitContainers": []any{
						map[string]any{
							"pDUContainerInformation": map[string]any{
								"uETimeZone": "+10:00+1",
							},
						},
						map[string]any{
							"pDUContainerInformation": map[string]any{
								"uETimeZone": "+10:00+0",
							},
						},
					},
				},
			},
		},
	}

	filter, err := ParsePathFilter("chargingFunctionRecord.listOfMultipleUnitUsage.usedUnitContainers.pDUContainerInformation.uETimeZone==+10:00+1")
	if err != nil {
		t.Fatal(err)
	}
	if !filter.Match(data) {
		t.Fatal("expected match")
	}

	filter.Value = "+10:00+9"
	if filter.Match(data) {
		t.Fatal("expected no match")
	}
}

func TestPathFilterMatchBoolAndNumber(t *testing.T) {
	data := map[string]any{
		"chargingFunctionRecord": map[string]any{
			"duration":                  148,
			"quotaManagementIndicator": false,
		},
	}

	durationFilter, _ := ParsePathFilter("chargingFunctionRecord.duration==148")
	if !durationFilter.Match(data) {
		t.Fatal("expected duration match")
	}

	boolFilter, _ := ParsePathFilter("chargingFunctionRecord.quotaManagementIndicator==false")
	if !boolFilter.Match(data) {
		t.Fatal("expected bool match")
	}
}
