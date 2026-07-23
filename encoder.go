package asn1x

import (
	"github.com/en-vee/asn1x/encode"
)

// Encoder encodes JSON-friendly values to BER using a parsed schema.
type Encoder = encode.Encoder

// EncodeOptions configures BER encoding behavior.
type EncodeOptions = encode.EncodeOptions

// NewEncoder returns an encoder bound to schema.
func NewEncoder(schema *Schema) *Encoder {
	return encode.NewEncoder(schema)
}

// NewEncoderWithOptions returns an encoder with field encode overrides.
// FieldSpecs uses the same map produced by LoadFieldSpecs / LoadFieldSpecsFile.
func NewEncoderWithOptions(schema *Schema, opts EncodeOptions) *Encoder {
	return encode.NewEncoderWithOptions(schema, opts)
}

// EncodeJSON unmarshals jsonBytes and encodes it as rootType using schema.
func EncodeJSON(schema *Schema, rootType string, jsonBytes []byte) ([]byte, error) {
	return encode.NewEncoder(schema).EncodeJSON(rootType, jsonBytes)
}
