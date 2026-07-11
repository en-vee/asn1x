package cdrfile

import (
	"fmt"

	"github.com/en-vee/asn1x/ber"
)

// Reader walks a 3GPP TS 32.297 CDR file payload.
type Reader struct {
	data []byte
	opts Options
}

// NewReader returns a reader over data using the given framing options.
func NewReader(data []byte, opts Options) *Reader {
	return &Reader{data: data, opts: opts}
}

// Remaining returns unread bytes.
func (r *Reader) Remaining() int {
	return len(r.data)
}

// ReadFileHeader parses the file header when enabled by options.
func (r *Reader) ReadFileHeader() (FileHeader, error) {
	if !r.opts.HasFileHeader {
		return FileHeader{}, fmt.Errorf("cdrfile: file header not enabled")
	}
	fh, n, err := ParseFileHeader(r.data)
	if err != nil {
		return FileHeader{}, err
	}
	r.data = r.data[n:]
	return fh, nil
}

// NextRecord returns the optional CDR header and raw BER bytes for one record.
func (r *Reader) NextRecord() (*RecordHeader, []byte, error) {
	if len(r.data) == 0 {
		return nil, nil, fmt.Errorf("cdrfile: no data")
	}

	var rh *RecordHeader
	if r.opts.HasCDRHeader {
		hdr, hdrLen, err := ParseRecordHeader(r.data)
		if err != nil {
			return nil, nil, err
		}
		rh = &hdr
		r.data = r.data[hdrLen:]

		cdrLen := int(hdr.CDRLength)
		if cdrLen == 0 {
			return rh, nil, fmt.Errorf("cdrfile: CDR length is 0")
		}
		if len(r.data) < cdrLen {
			return rh, nil, fmt.Errorf("cdrfile: CDR truncated (length %d, have %d bytes)", cdrLen, len(r.data))
		}
		cdr := r.data[:cdrLen]
		r.data = r.data[cdrLen:]
		return rh, cdr, nil
	}

	_, rest, err := ber.ReadOne(r.data)
	if err != nil {
		return nil, nil, err
	}
	cdr := r.data[:len(r.data)-len(rest)]
	r.data = rest
	return nil, cdr, nil
}
