package encode

import "github.com/en-vee/asn1x/schema"

// EncodeOptions configures an Encoder.
type EncodeOptions struct {
	// FieldSpecs maps qualified field paths to logical ASN.1 wire types,
	// using the same YAML-loaded map as decode.LoadFieldSpecs.
	FieldSpecs map[string]string
}

// NewEncoder returns an encoder for s.
func NewEncoder(s *schema.Schema) *Encoder {
	return NewEncoderWithOptions(s, EncodeOptions{})
}

// NewEncoderWithOptions returns an encoder for s with optional overrides.
func NewEncoderWithOptions(s *schema.Schema, opts EncodeOptions) *Encoder {
	specs := opts.FieldSpecs
	if specs == nil {
		specs = map[string]string{}
	}
	return &Encoder{schema: s, fieldSpecs: specs}
}
