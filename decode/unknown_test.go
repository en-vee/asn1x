package decode_test

import (
	"os"
	"testing"

	"github.com/en-vee/asn1x/decode"
)

func TestDecodeUnknownComponentAsHex(t *testing.T) {
	mod := loadCHFSchema(t)
	data, err := os.ReadFile("../sample-asn1-files/iot-1.ber")
	if err != nil {
		t.Fatal(err)
	}

	dec := decode.NewDecoder(mod)
	val, err := dec.DecodeBytes("CHFRecord", data)
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
	list, ok := cfr["listOfMultipleUnitUsage"].([]any)
	if !ok || len(list) == 0 {
		t.Fatalf("listOfMultipleUnitUsage = %#v", cfr["listOfMultipleUnitUsage"])
	}
	muu, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("multiple unit usage = %#v", list[0])
	}
	containers, ok := muu["usedUnitContainers"].([]any)
	if !ok || len(containers) == 0 {
		t.Fatalf("usedUnitContainers = %#v", muu["usedUnitContainers"])
	}
	container, ok := containers[0].(map[string]any)
	if !ok {
		t.Fatalf("used unit container = %#v", containers[0])
	}
	if got := container["Unknown_13"]; got != "01" {
		t.Fatalf("Unknown_13 = %#v, want %q", got, "01")
	}
	if container["quotaManagementIndicator"] != false {
		t.Fatalf("quotaManagementIndicator = %#v, want false", container["quotaManagementIndicator"])
	}
}
