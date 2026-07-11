package decode

import (
	"testing"

	"github.com/en-vee/asn1x/schema"
)

func TestDecodeIntegerTypeUnsignedRange(t *testing.T) {
	typ := schema.IntegerType{
		Constraints: []schema.Constraint{
			schema.ValueRangeConstraint{
				Lower: schema.RangeEndpoint{Kind: schema.EndpointNumber, Value: 0},
				Upper: schema.RangeEndpoint{Kind: schema.EndpointNumber, Value: 4294967295},
			},
		},
	}

	// 3764586444 == 0xE060A08C, negative when decoded as signed int32/int64 BER INTEGER.
	content := []byte{0xe0, 0x63, 0x07, 0xcc}
	val, err := decodeIntegerType(content, typ)
	if err != nil {
		t.Fatal(err)
	}
	if val != int64(3764586444) {
		t.Fatalf("got %#v, want 3764586444", val)
	}
}

func TestDecodeIntegerTypeSignedRange(t *testing.T) {
	typ := schema.IntegerType{
		Constraints: []schema.Constraint{
			schema.ValueRangeConstraint{
				Lower: schema.RangeEndpoint{Kind: schema.EndpointNumber, Value: 0},
				Upper: schema.RangeEndpoint{Kind: schema.EndpointNumber, Value: 255},
			},
		},
	}

	content := []byte{0x00, 0xc8}
	val, err := decodeIntegerType(content, typ)
	if err != nil {
		t.Fatal(err)
	}
	if val != int64(200) {
		t.Fatalf("got %#v", val)
	}
}
