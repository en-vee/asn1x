package asn1x

import "github.com/en-vee/asn1x/cdrfile"

// CDRFileOptions configures 3GPP TS 32.297 CDR file framing.
type CDRFileOptions = cdrfile.Options

// FileHeader is the CDR file header from 3GPP TS 32.297 clause 6.1.1.
type FileHeader = cdrfile.FileHeader

// RecordHeader is the per-CDR header from 3GPP TS 32.297 clause 6.1.2.
type RecordHeader = cdrfile.RecordHeader

// CDRFileReader walks a CDR file payload with optional file and record headers.
type CDRFileReader = cdrfile.Reader

// NewCDRFileReader returns a reader over data using the given framing options.
func NewCDRFileReader(data []byte, opts CDRFileOptions) *CDRFileReader {
	return cdrfile.NewReader(data, opts)
}

// ParseFileHeader decodes a CDR file header from data.
func ParseFileHeader(data []byte) (FileHeader, int, error) {
	return cdrfile.ParseFileHeader(data)
}

// ParseRecordHeader decodes a CDR record header from data.
func ParseRecordHeader(data []byte) (RecordHeader, int, error) {
	return cdrfile.ParseRecordHeader(data)
}
