package encode

import (
	"fmt"
	"math"
	"strings"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/schema"
)

func (e *Encoder) encodePrimitiveContent(t schema.Type, v any) ([]byte, error) {
	t = e.resolve(t)
	switch typ := t.(type) {
	case schema.IntegerType:
		return e.encodeIntegerContent(typ, v)
	case schema.EnumeratedType:
		return e.encodeEnumeratedContent(typ, v)
	case schema.BooleanType:
		b, err := asBool(v)
		if err != nil {
			return nil, err
		}
		return ber.EncodeBoolean(b), nil
	case schema.NullType:
		if v != nil {
			return nil, fmt.Errorf("NULL value must be null, got %T", v)
		}
		return nil, nil
	case schema.OctetStringType:
		return encodeOctetStringValue(v)
	case schema.StringType:
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case schema.BitStringType:
		return encodeBitStringContent(v)
	case schema.ObjectIdentifierType:
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return encodeObjectIdentifier(s)
	case schema.UTCTimeType, schema.GeneralizedTimeType:
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case schema.ReferenceType:
		if strings.Contains(typ.Name, ".&") {
			return encodeOctetStringValue(v)
		}
		return nil, fmt.Errorf("unresolved type reference %q", typ.Name)
	default:
		return encodeOctetStringValue(v)
	}
}

func (e *Encoder) encodeIntegerContent(typ schema.IntegerType, v any) ([]byte, error) {
	if s, ok := v.(string); ok {
		if n, ok := namedNumberValue(typ.NamedNumbers, s); ok {
			return encodeSignedIntegerContent(int64(n)), nil
		}
		return nil, fmt.Errorf("unknown INTEGER named value %q", s)
	}
	if vr, ok := schema.ValueRangeConstraintFrom(typ.Constraints); ok && isUnsignedIntegerRange(*vr) {
		n, err := asUint64(v)
		if err != nil {
			return nil, err
		}
		return ber.EncodeIntegerUnsigned(n), nil
	}
	n, err := asInt64(v)
	if err != nil {
		return nil, err
	}
	return encodeSignedIntegerContent(n), nil
}

func (e *Encoder) encodeEnumeratedContent(typ schema.EnumeratedType, v any) ([]byte, error) {
	if s, ok := v.(string); ok {
		if n, ok := namedNumberValue(typ.Values, s); ok {
			return encodeSignedIntegerContent(int64(n)), nil
		}
		return nil, fmt.Errorf("unknown ENUMERATED named value %q", s)
	}
	n, err := asInt64(v)
	if err != nil {
		return nil, err
	}
	return encodeSignedIntegerContent(n), nil
}

func encodeSignedIntegerContent(n int64) []byte {
	return ber.EncodeInteger(n)
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
