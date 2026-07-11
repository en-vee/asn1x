package asn1x

import (
	"io"
	"os"

	"github.com/en-vee/asn1x/decode"
)

// Decoder decodes BER-encoded ASN.1 values using a parsed schema.
type Decoder = decode.Decoder

// DecodeOptions configures BER decoding behavior.
type DecodeOptions = decode.DecodeOptions

// NewDecoder returns a decoder bound to schema.
func NewDecoder(schema *Schema) *Decoder {
	return decode.NewDecoder(schema)
}

// NewDecoderWithOptions returns a decoder with field decode overrides.
func NewDecoderWithOptions(schema *Schema, opts DecodeOptions) *Decoder {
	return decode.NewDecoderWithOptions(schema, opts)
}

// LoadFieldSpecs loads YAML field decode overrides.
func LoadFieldSpecs(r io.Reader) (map[string]string, error) {
	return decode.LoadFieldSpecs(r)
}

// LoadFieldSpecsFile loads YAML field decode overrides from path.
func LoadFieldSpecsFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decode.LoadFieldSpecs(f)
}

// ToJSON marshals a decoded ASN.1 value as indented JSON.
func ToJSON(v any) ([]byte, error) {
	return decode.ToJSON(v)
}

// DecodeReader reads one BER value from r using rootType from schema.
func DecodeReader(schema *Schema, rootType string, r io.Reader) (any, error) {
	return decode.NewDecoder(schema).Decode(rootType, r)
}
