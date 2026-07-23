package ber

// EncodeTag encodes a BER identifier.
func EncodeTag(tag Tag) []byte {
	first := byte(int(tag.Class)<<6) & 0xc0
	if tag.Constructed {
		first |= 0x20
	}
	if tag.Number < 0x1f {
		return []byte{first | byte(tag.Number)}
	}
	first |= 0x1f
	n := tag.Number
	var digits []byte
	for {
		digits = append(digits, byte(n&0x7f))
		n >>= 7
		if n == 0 {
			break
		}
	}
	// digits are least-significant first; reverse and set continuation bits.
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	for i := 0; i < len(digits)-1; i++ {
		digits[i] |= 0x80
	}
	out := make([]byte, 1+len(digits))
	out[0] = first
	copy(out[1:], digits)
	return out
}

// EncodeLength encodes a definite BER length.
func EncodeLength(n int) []byte {
	if n < 0 {
		n = 0
	}
	if n <= 0x7f {
		return []byte{byte(n)}
	}
	var body []byte
	for x := n; x > 0; x >>= 8 {
		body = append(body, byte(x))
	}
	for i, j := 0, len(body)-1; i < j; i, j = i+1, j-1 {
		body[i], body[j] = body[j], body[i]
	}
	out := make([]byte, 1+len(body))
	out[0] = 0x80 | byte(len(body))
	copy(out[1:], body)
	return out
}

// EncodeTLV encodes a tag-length-value element.
func EncodeTLV(tag Tag, value []byte) []byte {
	tagBytes := EncodeTag(tag)
	lenBytes := EncodeLength(len(value))
	out := make([]byte, 0, len(tagBytes)+len(lenBytes)+len(value))
	out = append(out, tagBytes...)
	out = append(out, lenBytes...)
	out = append(out, value...)
	return out
}

// EncodeInteger encodes a signed BER INTEGER content octet string.
func EncodeInteger(n int64) []byte {
	if n >= 0 {
		return EncodeIntegerUnsigned(uint64(n))
	}

	var raw [8]byte
	u := uint64(n)
	for i := 7; i >= 0; i-- {
		raw[i] = byte(u)
		u >>= 8
	}
	start := 0
	for start < 7 && raw[start] == 0xff && raw[start+1]&0x80 != 0 {
		start++
	}
	out := make([]byte, 8-start)
	copy(out, raw[start:])
	return out
}

// EncodeIntegerUnsigned encodes an unsigned BER INTEGER content octet string.
func EncodeIntegerUnsigned(n uint64) []byte {
	if n == 0 {
		return []byte{0x00}
	}
	var buf []byte
	for x := n; x > 0; x >>= 8 {
		buf = append(buf, byte(x))
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	if buf[0]&0x80 != 0 {
		buf = append([]byte{0x00}, buf...)
	}
	return buf
}

// EncodeBoolean encodes a BER BOOLEAN content octet string.
func EncodeBoolean(v bool) []byte {
	if v {
		return []byte{0xff}
	}
	return []byte{0x00}
}
