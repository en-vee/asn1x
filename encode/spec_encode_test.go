package encode

import (
	"bytes"
	"testing"

	"github.com/en-vee/asn1x/decode"
)

func TestEncodePLMNIDRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		wire []byte
		mcc  string
		mnc  string
	}{
		{name: "505-01", wire: []byte{0x05, 0xf5, 0x10}, mcc: "505", mnc: "01"},
		{name: "505-93", wire: []byte{0x05, 0xf5, 0x39}, mcc: "505", mnc: "93"},
		{name: "310-410", wire: []byte{0x13, 0x40, 0x10}, mcc: "310", mnc: "410"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodePLMNID(map[string]string{"mcc": tt.mcc, "mnc": tt.mnc})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.wire) {
				t.Fatalf("encodePLMNID() = %x, want %x", got, tt.wire)
			}
		})
	}
}

func TestEncodeMSTimeZoneRoundTrip(t *testing.T) {
	tests := []struct {
		text string
		wire []byte
	}{
		{text: "+10:00+0", wire: []byte{0x04, 0x00}},
		{text: "+10:00+1", wire: []byte{0x04, 0x01}},
		{text: "+01:00+0", wire: []byte{0x40, 0x00}},
		{text: "+10:00+2", wire: []byte{0x04, 0x02}},
		{text: "-06:00+0", wire: []byte{0x4a, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := encodeMSTimeZone(tt.text)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.wire) {
				t.Fatalf("encodeMSTimeZone() = %x, want %x", got, tt.wire)
			}
			back, err := decode.FormatMSTimeZoneOctets(got[0], got[1])
			if err != nil {
				t.Fatal(err)
			}
			if back != tt.text {
				t.Fatalf("decode after encode = %q, want %q", back, tt.text)
			}
		})
	}
}

func TestEncodeTimeStamp(t *testing.T) {
	got, err := encodeTimeStamp("2022-04-29T01:11:39Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 9 {
		t.Fatalf("len = %d, want 9", len(got))
	}
	want := []byte{0x22, 0x04, 0x29, 0x01, 0x11, 0x39, '+', 0x00, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
}
