package decode

import (
	"fmt"
	"io"
	"strings"

	"github.com/en-vee/asn1x/schema"
	"gopkg.in/yaml.v3"
)

// FieldSpec maps a qualified schema field path to a logical ASN.1 type used
// for decoding when the on-wire representation does not match the schema type.
type FieldSpec struct {
	FieldPath    string
	ASN1DataType string
}

type decodeSpecsFile struct {
	ASN1x struct {
		DecodeSpecs []fieldSpecEntry `yaml:"decodeSpecs"`
	} `yaml:"asn1x"`
}

type fieldSpecEntry struct {
	FieldPath    string `yaml:"fieldPath"`
	ASN1DataType string `yaml:"asn1DataType"`
}

// LoadFieldSpecs reads decode field overrides from YAML.
func LoadFieldSpecs(r io.Reader) (map[string]string, error) {
	var cfg decodeSpecsFile
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode specs: %w", err)
	}

	specs := make(map[string]string, len(cfg.ASN1x.DecodeSpecs))
	for _, entry := range cfg.ASN1x.DecodeSpecs {
		path := strings.TrimSpace(entry.FieldPath)
		typ := strings.TrimSpace(entry.ASN1DataType)
		if err := validateFieldPath(path); err != nil {
			return nil, err
		}
		if typ == "" {
			return nil, fmt.Errorf("decode specs: asn1DataType is required for fieldPath %q", path)
		}
		specs[path] = typ
	}
	return specs, nil
}

// DecodeOptions configures a Decoder.
type DecodeOptions struct {
	FieldSpecs map[string]string
}

// NewDecoderWithOptions returns a decoder for s with optional overrides.
func NewDecoderWithOptions(s *schema.Schema, opts DecodeOptions) *Decoder {
	specs := opts.FieldSpecs
	if specs == nil {
		specs = map[string]string{}
	}
	return &Decoder{schema: s, fieldSpecs: specs}
}
