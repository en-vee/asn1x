package decode

import "testing"

func TestDecodeMSTimeZone0400(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x04, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00+0" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZone0401(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x04, 0x01})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00+1" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneOneHour(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x40, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+01:00+0" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneDSTTwoHours(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x04, 0x02})
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00+2" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneText(t *testing.T) {
	val, err := decodeMSTimeZone([]byte("+10:00+0"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00+0" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneTextWithDST(t *testing.T) {
	val, err := decodeMSTimeZone([]byte("+10:00+1"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "+10:00+1" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneNegative(t *testing.T) {
	val, err := decodeMSTimeZone([]byte{0x4a, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val != "-06:00+0" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeMSTimeZoneReservedDST(t *testing.T) {
	_, err := decodeMSTimeZone([]byte{0x04, 0x03})
	if err == nil {
		t.Fatal("expected error for reserved DST value")
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
