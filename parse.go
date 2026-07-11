package asn1x

import (
	"io"

	"github.com/en-vee/asn1x/schema"
)

// Schema holds a parsed ASN.1 module.
type Schema = schema.Schema

// Parse reads an ASN.1 syntax schema from r and returns the parsed module.
func Parse(r io.Reader) (*Schema, error) {
	return schema.Parse(r)
}
