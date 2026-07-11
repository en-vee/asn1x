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
