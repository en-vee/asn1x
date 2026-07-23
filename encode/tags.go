package encode

import (
	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/schema"
)

func (e *Encoder) expectedTag(comp schema.Component) ber.Tag {
	typ := e.resolve(comp.Type)
	if comp.Tag != nil {
		return ber.Tag{
			Class:       ber.Class(comp.Tag.Class),
			Number:      comp.Tag.Number,
			Constructed: isConstructedType(typ),
		}
	}
	return universalTag(typ)
}

func isConstructedType(t schema.Type) bool {
	switch t.(type) {
	case schema.SetType, schema.SequenceType, schema.ChoiceType, schema.SequenceOfType, schema.SetOfType:
		return true
	default:
		return false
	}
}

func universalTag(t schema.Type) ber.Tag {
	switch typ := t.(type) {
	case schema.IntegerType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 2, Constructed: false}
	case schema.EnumeratedType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 10, Constructed: false}
	case schema.BooleanType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 1, Constructed: false}
	case schema.NullType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 5, Constructed: false}
	case schema.OctetStringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 4, Constructed: false}
	case schema.BitStringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 3, Constructed: false}
	case schema.ObjectIdentifierType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 6, Constructed: false}
	case schema.StringType:
		return ber.Tag{Class: ber.ClassUniversal, Number: stringTypeTag(typ.Name), Constructed: false}
	case schema.SequenceType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 16, Constructed: true}
	case schema.SetType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 17, Constructed: true}
	case schema.SequenceOfType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 16, Constructed: true}
	case schema.SetOfType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 17, Constructed: true}
	case schema.ChoiceType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 0, Constructed: true}
	case schema.UTCTimeType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 23, Constructed: false}
	case schema.GeneralizedTimeType:
		return ber.Tag{Class: ber.ClassUniversal, Number: 24, Constructed: false}
	default:
		return ber.Tag{Class: ber.ClassUniversal, Number: 4, Constructed: false}
	}
}

func stringTypeTag(name string) int {
	switch name {
	case "UTF8String":
		return 12
	case "NumericString":
		return 18
	case "PrintableString":
		return 19
	case "IA5String":
		return 22
	case "VisibleString", "ISO646String":
		return 26
	case "GeneralizedTime":
		return 24
	default:
		return 12
	}
}
