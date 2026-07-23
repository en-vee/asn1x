package ber_test

import (
	"encoding/hex"
	"testing"

	"github.com/en-vee/asn1x/ber"
)

func TestReadOneCHFRecordTag(t *testing.T) {
	raw, err := hex.DecodeString("bf814803010203")
	if err != nil {
		t.Fatal(err)
	}
	tlv, rest, err := ber.ReadOne(raw)
	if err != nil {
		t.Fatal(err)
	}
	if tlv.Tag.Class != ber.ClassContext || tlv.Tag.Number != 200 || !tlv.Tag.Constructed {
		t.Fatalf("tag = %+v", tlv.Tag)
	}
	if len(tlv.Value) != 3 || tlv.Value[0] != 0x01 {
		t.Fatalf("value = %x", tlv.Value)
	}
	if len(rest) != 0 {
		t.Fatalf("rest = %x", rest)
	}
}

func TestDecodeInteger(t *testing.T) {
	n, err := ber.DecodeInteger([]byte{0x00, 0xc8})
	if err != nil {
		t.Fatal(err)
	}
	if n != 200 {
		t.Fatalf("got %d", n)
	}
}

func TestDecodeIntegerUnsigned(t *testing.T) {
	n, err := ber.DecodeIntegerUnsigned([]byte{0xe0, 0x63, 0x07, 0xcc})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3764586444 {
		t.Fatalf("got %d", n)
	}
}

func TestEncodeTLVRoundTripCHFTag(t *testing.T) {
	raw := ber.EncodeTLV(ber.Tag{Class: ber.ClassContext, Number: 200, Constructed: true}, []byte{0x01, 0x02, 0x03})
	want, err := hex.DecodeString("bf814803010203")
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(raw) != hex.EncodeToString(want) {
		t.Fatalf("EncodeTLV = %x, want %x", raw, want)
	}
	tlv, rest, err := ber.ReadOne(raw)
	if err != nil {
		t.Fatal(err)
	}
	if tlv.Tag.Class != ber.ClassContext || tlv.Tag.Number != 200 || !tlv.Tag.Constructed {
		t.Fatalf("tag = %+v", tlv.Tag)
	}
	if len(rest) != 0 {
		t.Fatalf("rest = %x", rest)
	}
}

func TestEncodeIntegerRoundTrip(t *testing.T) {
	cases := []int64{0, 1, 127, 128, 200, -1, -128, -129, 3764586444}
	for _, want := range cases {
		var content []byte
		if want < 0 {
			content = ber.EncodeInteger(want)
		} else {
			content = ber.EncodeIntegerUnsigned(uint64(want))
		}
		got, err := ber.DecodeInteger(content)
		if err != nil {
			t.Fatalf("DecodeInteger(%x) error = %v", content, err)
		}
		if got != want {
			t.Fatalf("round-trip %d: content=%x got %d", want, content, got)
		}
	}
}

func TestEncodeBoolean(t *testing.T) {
	if got := ber.EncodeBoolean(true); len(got) != 1 || got[0] == 0 {
		t.Fatalf("true = %x", got)
	}
	if got := ber.EncodeBoolean(false); len(got) != 1 || got[0] != 0 {
		t.Fatalf("false = %x", got)
	}
}
