package decode

import (
	"math"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/schema"
)

func decodeIntegerType(content []byte, typ schema.IntegerType) (any, error) {
	if vr, ok := schema.ValueRangeConstraintFrom(typ.Constraints); ok && isUnsignedIntegerRange(*vr) {
		n, err := ber.DecodeIntegerUnsigned(content)
		if err != nil {
			return nil, err
		}
		if name := namedNumberFromUint(typ.NamedNumbers, n); name != "" {
			return name, nil
		}
		if n <= math.MaxInt64 {
			return int64(n), nil
		}
		return n, nil
	}

	n, err := ber.DecodeInteger(content)
	if err != nil {
		return nil, err
	}
	if name := namedNumber(typ.NamedNumbers, int(n)); name != "" {
		return name, nil
	}
	return n, nil
}

func isUnsignedIntegerRange(vr schema.ValueRangeConstraint) bool {
	if vr.Lower.Kind == schema.EndpointNumber && vr.Lower.Value < 0 {
		return false
	}
	if vr.Upper.Kind == schema.EndpointNumber && vr.Upper.Value > math.MaxInt32 {
		return true
	}
	return false
}

func namedNumberFromUint(values map[string]int, n uint64) string {
	if n > math.MaxInt {
		return ""
	}
	return namedNumber(values, int(n))
}
