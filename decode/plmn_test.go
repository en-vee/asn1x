package decode

import "testing"

func TestDecodePLMNIDTwoDigitMNC(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		wantMCC string
		wantMNC string
	}{
		{
			name:    "505-01",
			in:      []byte{0x05, 0xf5, 0x10},
			wantMCC: "505",
			wantMNC: "01",
		},
		{
			name:    "505-93",
			in:      []byte{0x05, 0xf5, 0x39},
			wantMCC: "505",
			wantMNC: "93",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodePLMNID(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got["mcc"] != tt.wantMCC || got["mnc"] != tt.wantMNC {
				t.Fatalf("decodePLMNID() = %#v, want mcc=%q mnc=%q", got, tt.wantMCC, tt.wantMNC)
			}
		})
	}
}

func TestDecodePLMNIDThreeDigitMNC(t *testing.T) {
	got, err := decodePLMNID([]byte{0x13, 0x40, 0x10})
	if err != nil {
		t.Fatal(err)
	}
	if got["mcc"] != "310" || got["mnc"] != "410" {
		t.Fatalf("decodePLMNID() = %#v, want mcc=310 mnc=410", got)
	}
}

func TestDecodeWithSpecTypePLMNID(t *testing.T) {
	val, err := decodeWithSpecType("PLMNID", []byte{0x05, 0xf5, 0x10})
	if err != nil {
		t.Fatal(err)
	}
	plmn := val.(map[string]string)
	if plmn["mcc"] != "505" || plmn["mnc"] != "01" {
		t.Fatalf("got %#v", plmn)
	}
}
