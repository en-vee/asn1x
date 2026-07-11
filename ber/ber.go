package ber

import (
	"fmt"
	"io"
)

// Class is the tag class in BER encoding.
type Class int

const (
	ClassUniversal Class = iota
	ClassApplication
	ClassContext
	ClassPrivate
)

// Tag is a BER identifier octet sequence.
type Tag struct {
	Class       Class
	Number      int
	Constructed bool
}

// TLV is one BER tag-length-value element.
type TLV struct {
	Tag   Tag
	Value []byte
}

// Reader reads BER TLV elements from a byte slice.
type Reader struct {
	data []byte
}

// NewReader returns a reader over data.
func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

// Remaining returns unread bytes.
func (r *Reader) Remaining() int {
	return len(r.data)
}

// Read reads the next TLV element.
func (r *Reader) Read() (TLV, error) {
	if len(r.data) == 0 {
		return TLV{}, io.EOF
	}
	off := 0
	tag, off, err := readTag(r.data, off)
	if err != nil {
		return TLV{}, err
	}
	length, off, err := readLength(r.data, off)
	if err != nil {
		return TLV{}, err
	}
	if off+length > len(r.data) {
		return TLV{}, fmt.Errorf("ber: truncated value for tag %d", tag.Number)
	}
	val := r.data[off : off+length]
	r.data = r.data[off+length:]
	return TLV{Tag: tag, Value: val}, nil
}

// ReadOne reads a single TLV from data and returns the rest.
func ReadOne(data []byte) (TLV, []byte, error) {
	r := NewReader(data)
	tlv, err := r.Read()
	if err != nil {
		return TLV{}, nil, err
	}
	return tlv, r.data, nil
}

func readTag(data []byte, off int) (Tag, int, error) {
	if off >= len(data) {
		return Tag{}, off, fmt.Errorf("ber: unexpected end reading tag")
	}
	b := data[off]
	off++
	tag := Tag{
		Class:       Class(b >> 6),
		Constructed: (b>>5)&1 == 1,
		Number:      int(b & 0x1f),
	}
	if tag.Number == 0x1f {
		tag.Number = 0
		for {
			if off >= len(data) {
			 return Tag{}, off, fmt.Errorf("ber: unexpected end reading tag number")
			}
			b = data[off]
			off++
			tag.Number = (tag.Number << 7) | int(b&0x7f)
			if b&0x80 == 0 {
				break
			}
		}
	}
	return tag, off, nil
}

func readLength(data []byte, off int) (int, int, error) {
	if off >= len(data) {
		return 0, off, fmt.Errorf("ber: unexpected end reading length")
	}
	b := data[off]
	off++
	if b&0x80 == 0 {
		return int(b), off, nil
	}
	n := int(b & 0x7f)
	if n == 0 {
		return 0, off, fmt.Errorf("ber: indefinite length not supported")
	}
	if off+n > len(data) {
		return 0, off, fmt.Errorf("ber: truncated length")
	}
	length := 0
	for i := 0; i < n; i++ {
		length = (length << 8) | int(data[off])
		off++
	}
	return length, off, nil
}

// DecodeInteger decodes a BER INTEGER content octet string as signed.
func DecodeInteger(content []byte) (int64, error) {
	if len(content) == 0 {
		return 0, fmt.Errorf("ber: empty integer")
	}
	var n int64
	negative := content[0]&0x80 != 0
	for _, b := range content {
		n = (n << 8) | int64(b)
	}
	if negative {
		bits := uint(len(content) * 8)
		n -= int64(1) << bits
	}
	return n, nil
}

// DecodeIntegerUnsigned decodes a BER INTEGER content octet string as unsigned.
func DecodeIntegerUnsigned(content []byte) (uint64, error) {
	if len(content) == 0 {
		return 0, fmt.Errorf("ber: empty integer")
	}
	var n uint64
	for _, b := range content {
		n = (n << 8) | uint64(b)
	}
	return n, nil
}

// DecodeBoolean decodes a BER BOOLEAN content octet string.
func DecodeBoolean(content []byte) (bool, error) {
	if len(content) != 1 {
		return false, fmt.Errorf("ber: invalid boolean length %d", len(content))
	}
	return content[0] != 0, nil
}
