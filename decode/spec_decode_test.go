package decode

import "testing"

func TestDecodeWithSpecTypeUTCTime(t *testing.T) {
	val, err := decodeWithSpecType("UTCTime", []byte("220429011139Z"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "2022-04-29T01:11:39Z" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeWithSpecTypeGeneralizedTime(t *testing.T) {
	val, err := decodeWithSpecType("GeneralizedTime", []byte("2022-04-29T01:04:42Z"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "2022-04-29T01:04:42Z" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeWithSpecTypeBCDTimeStamp(t *testing.T) {
	// 2022-04-29 01:11:39 +10:00 local -> 2022-04-28T15:11:39Z UTC
	content := []byte{0x22, 0x04, 0x29, 0x01, 0x11, 0x39, '+', 0x10, 0x00}
	val, err := decodeWithSpecType("GeneralizedTime", content)
	if err != nil {
		t.Fatal(err)
	}
	if val != "2022-04-28T15:11:39Z" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeWithSpecTypeBCDTimeStampSignZero(t *testing.T) {
	// 3GPP uses ASCII '0' for '+'.
	content := []byte{0x22, 0x04, 0x29, 0x01, 0x11, 0x39, '0', 0x10, 0x00}
	val, err := decodeWithSpecType("TimeStamp", content)
	if err != nil {
		t.Fatal(err)
	}
	if val != "2022-04-28T15:11:39Z" {
		t.Fatalf("got %#v", val)
	}
}

func TestDecodeBCDTimeStampNegativeOffset(t *testing.T) {
	// 2022-04-29 01:11:39 -05:00 local -> 2022-04-29T06:11:39Z UTC
	content := []byte{0x22, 0x04, 0x29, 0x01, 0x11, 0x39, '-', 0x05, 0x00}
	val, err := decodeBCDTimeStamp(content)
	if err != nil {
		t.Fatal(err)
	}
	if val != "2022-04-29T06:11:39Z" {
		t.Fatalf("got %#v", val)
	}
}
