package decode

import "testing"

func TestDecodeMSTimeZone0400(t *testing.T) {
	// Kafka / charging wire format: hex "0400" → +10:00 (semi-octet BCD).
	val, err := decodeMSTimeZone([]byte{0x04, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneOneHour(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x40, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+01:00" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneText(t *testing.T) {
	val, err := decodeMSTimeZone([]byte("+10:00+0"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneNegative(t *testing.T) {
	// Wireshark example: 0x4A → GMT-6
	val, err := decodeMSTimeZone([]byte{0x4a, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "-06:00" {
		t.Fatalf("got %#v", val)
	}
}

func TestGSMTimeZoneUnits(t *testing.T) {
	tests := []struct {
		b     byte
		units int
	}{
		{0x04, 40},
		{0x40, 4},
		{0x4a, -24},
		{0x00, 0},
	}
	for _, tc := range tests {
		if got := gsmTimeZoneUnits(tc.b); got != tc.units {
			t.Fatalf("%02x: got %d want %d", tc.b, got, tc.units)
		}
	}
}
