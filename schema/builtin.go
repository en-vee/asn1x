package schema

// builtin string and time type names from ITU-T X.680.
var builtinTypeNames = map[string]Kind{
	"NumericString":     KindString,
	"PrintableString":   KindString,
	"TeletexString":     KindString,
	"T61String":         KindString,
	"VideotexString":    KindString,
	"IA5String":         KindString,
	"GraphicString":     KindString,
	"VisibleString":     KindString,
	"ISO646String":      KindString,
	"GeneralString":     KindString,
	"UniversalString":   KindString,
	"BMPString":         KindString,
	"UTF8String":          KindString,
	"UTCTime":           KindUTCTime,
	"GeneralizedTime":   KindGeneralizedTime,
	"RELATIVE-OID":      KindRelativeOID,
	"RELATIVE-OID-IRI":  KindRelativeOID,
	"OID-IRI":           KindObjectIdentifier,
}

func builtinTypeFromName(name string) (Type, bool) {
	kind, ok := builtinTypeNames[name]
	if !ok {
		return nil, false
	}
	switch kind {
	case KindString:
		return StringType{Name: name}, true
	case KindUTCTime:
		return UTCTimeType{}, true
	case KindGeneralizedTime:
		return GeneralizedTimeType{}, true
	case KindRelativeOID:
		return RelativeOIDType{}, true
	case KindObjectIdentifier:
		return ObjectIdentifierType{}, true
	default:
		return nil, false
	}
}

func isKnownStringType(name string) bool {
	kind, ok := builtinTypeNames[name]
	return ok && kind == KindString
}
