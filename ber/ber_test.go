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
