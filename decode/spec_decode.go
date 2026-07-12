package decode

import (
	"fmt"
	"strings"

	"github.com/en-vee/asn1x/ber"
)

func decodeWithSpecType(typeName string, content []byte) (any, error) {
	switch normalizeSpecType(typeName) {
	case "UTCTime":
		return decodeTimeValue(content, true)
	case "GeneralizedTime":
		return decodeTimeValue(content, false)
	case "TimeStamp":
		return decodeTimeValue(content, true)
	case "MSTimeZone", "TimeZone":
		return decodeMSTimeZone(content)
	case "PLMNID", "PLMNIdentifier", "PLMN":
		return decodePLMNID(content)
	case "IA5String", "UTF8String", "PrintableString", "VisibleString", "NumericString":
		return string(content), nil
	case "Integer":
		return decodeIntegerValue(content)
	case "OctetString":
		return decodeOctetString(content), nil
	default:
		return nil, fmt.Errorf("unsupported decode spec type %q", typeName)
	}
}

func normalizeSpecType(typeName string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "utctime":
		return "UTCTime"
	case "generalizedtime":
		return "GeneralizedTime"
	case "timestamp":
		return "TimeStamp"
	case "mstimezone", "timezone":
		return "MSTimeZone"
	case "plmnid", "plmnidentifier", "plmn":
		return "PLMNID"
	case "ia5string":
		return "IA5String"
	case "utf8string":
		return "UTF8String"
	case "printablestring":
		return "PrintableString"
	case "visiblestring":
		return "VisibleString"
	case "numericstring":
		return "NumericString"
	case "integer":
		return "Integer"
	case "octetstring":
		return "OctetString"
	default:
		return typeName
	}
}

func decodeIntegerValue(content []byte) (any, error) {
	n, err := ber.DecodeInteger(content)
	if err != nil {
		return nil, err
	}
	return n, nil
}
