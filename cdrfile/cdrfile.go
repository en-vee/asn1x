package cdrfile

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/en-vee/asn1x/decode"
)

// Options configures how a 3GPP TS 32.297 CDR file is framed.
type Options struct {
	HasFileHeader bool
	HasCDRHeader  bool
}

// FileHeader is the CDR file header defined in 3GPP TS 32.297 clause 6.1.1.
type FileHeader struct {
	FileLength              uint32 `json:"fileLength"`
	HeaderLength            uint32 `json:"headerLength"`
	HighReleaseIdentifier   uint8  `json:"highReleaseIdentifier"`
	HighVersionIdentifier   uint8  `json:"highVersionIdentifier"`
	LowReleaseIdentifier    uint8  `json:"lowReleaseIdentifier"`
	LowVersionIdentifier    uint8  `json:"lowVersionIdentifier"`
	FileOpeningTimestamp    string `json:"fileOpeningTimestamp"`
	LastCDRAppendTimestamp  string `json:"lastCDRAppendTimestamp"`
	NumberOfCDRs            uint32 `json:"numberOfCDRsInFile"`
	FileSequenceNumber      uint32 `json:"fileSequenceNumber"`
	FileClosureTrigger      uint8  `json:"fileClosureTriggerReason"`
	FileClosureTriggerText  string `json:"fileClosureTriggerReasonText"`
	NodeIPAddress           string `json:"nodeIPAddress"`
	LostCDRIndicator        uint8  `json:"lostCDRIndicator"`
	CDRRoutingFilterLength  uint16 `json:"cdrRoutingFilterLength"`
	CDRRoutingFilter        string `json:"cdrRoutingFilter,omitempty"`
	PrivateExtensionLength  uint16 `json:"privateExtensionLength,omitempty"`
	PrivateExtension        string `json:"privateExtension,omitempty"`
	HighReleaseExtension    *uint8 `json:"highReleaseIdentifierExtension,omitempty"`
	LowReleaseExtension     *uint8 `json:"lowReleaseIdentifierExtension,omitempty"`
	HighRelease             string `json:"highRelease"`
	LowRelease              string `json:"lowRelease"`
}

// RecordHeader is the per-CDR header defined in 3GPP TS 32.297 clause 6.1.2.
type RecordHeader struct {
	CDRLength                      uint16 `json:"cdrLength"`
	ReleaseIdentifier              uint8  `json:"releaseIdentifier"`
	VersionIdentifier              uint8  `json:"versionIdentifier"`
	DataRecordFormat               uint8  `json:"dataRecordFormat"`
	DataRecordFormatText           string `json:"dataRecordFormatText"`
	TSNumber                       uint8  `json:"tsNumber"`
	TSNumberText                   string `json:"tsNumberText"`
	ReleaseIdentifierExtension     *uint8 `json:"releaseIdentifierExtension,omitempty"`
	Release                        string `json:"release"`
	HeaderLength                   int    `json:"headerLength"`
}

// ParseFileHeader decodes a CDR file header from data.
func ParseFileHeader(data []byte) (FileHeader, int, error) {
	if len(data) < 50 {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: file header truncated (have %d bytes, need at least 50)", len(data))
	}

	hdrLen := binary.BigEndian.Uint32(data[4:8])
	if hdrLen < 48 {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: invalid header length %d", hdrLen)
	}
	if int(hdrLen) > len(data) {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: file header truncated (header length %d, have %d bytes)", hdrLen, len(data))
	}

	highOct := data[8]
	lowOct := data[9]
	highRel := (highOct >> 5) & 0x07
	highVer := highOct & 0x1f
	lowRel := (lowOct >> 5) & 0x07
	lowVer := lowOct & 0x1f

	fh := FileHeader{
		FileLength:             binary.BigEndian.Uint32(data[0:4]),
		HeaderLength:           hdrLen,
		HighReleaseIdentifier:  highRel,
		HighVersionIdentifier:  highVer,
		LowReleaseIdentifier:   lowRel,
		LowVersionIdentifier:   lowVer,
		FileOpeningTimestamp:   formatPackedTimestamp(binary.BigEndian.Uint32(data[10:14])),
		LastCDRAppendTimestamp: formatPackedTimestamp(binary.BigEndian.Uint32(data[14:18])),
		NumberOfCDRs:           binary.BigEndian.Uint32(data[18:22]),
		FileSequenceNumber:     binary.BigEndian.Uint32(data[22:26]),
		FileClosureTrigger:     data[26],
		FileClosureTriggerText: closureTriggerText(data[26]),
		NodeIPAddress:          formatNodeIPAddress(data[27:47]),
		LostCDRIndicator:       data[47],
	}

	offset := 48
	routingLen := binary.BigEndian.Uint16(data[offset : offset+2])
	fh.CDRRoutingFilterLength = routingLen
	offset += 2
	if offset+int(routingLen) > int(hdrLen) {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: truncated CDR routing filter")
	}
	if routingLen > 0 {
		fh.CDRRoutingFilter = string(data[offset : offset+int(routingLen)])
	}
	offset += int(routingLen)

	releaseExtBytes := releaseExtensionBytes(highRel, lowRel)
	remaining := int(hdrLen) - offset

	if remaining > releaseExtBytes && remaining >= 2 {
		privLen := binary.BigEndian.Uint16(data[offset : offset+2])
		if offset+2+int(privLen)+releaseExtBytes == int(hdrLen) {
			fh.PrivateExtensionLength = privLen
			offset += 2
			if privLen > 0 {
				fh.PrivateExtension = hex.EncodeToString(data[offset : offset+int(privLen)])
			}
			offset += int(privLen)
			remaining = int(hdrLen) - offset
		}
	}

	if remaining != releaseExtBytes {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: unexpected file header tail (remaining %d bytes, expected %d release extension bytes)", remaining, releaseExtBytes)
	}

	if highRel == 7 {
		ext := data[offset]
		fh.HighReleaseExtension = &ext
		offset++
	}
	if lowRel == 7 {
		ext := data[offset]
		fh.LowReleaseExtension = &ext
		offset++
	}

	if offset != int(hdrLen) {
		return FileHeader{}, 0, fmt.Errorf("cdrfile: parsed %d header bytes, expected %d", offset, hdrLen)
	}

	fh.HighRelease = formatRelease(highRel, fh.HighReleaseExtension)
	fh.LowRelease = formatRelease(lowRel, fh.LowReleaseExtension)
	return fh, int(hdrLen), nil
}

// ParseRecordHeader decodes a CDR record header from data.
func ParseRecordHeader(data []byte) (RecordHeader, int, error) {
	if len(data) < 4 {
		return RecordHeader{}, 0, fmt.Errorf("cdrfile: CDR header truncated (have %d bytes, need at least 4)", len(data))
	}

	oct3 := data[2]
	oct4 := data[3]
	rel := (oct3 >> 5) & 0x07
	ver := oct3 & 0x1f
	format := (oct4 >> 5) & 0x07
	tsNum := oct4 & 0x1f

	rh := RecordHeader{
		CDRLength:            binary.BigEndian.Uint16(data[0:2]),
		ReleaseIdentifier:    rel,
		VersionIdentifier:    ver,
		DataRecordFormat:     format,
		DataRecordFormatText: dataRecordFormatText(format),
		TSNumber:             tsNum,
		TSNumberText:         tsNumberText(tsNum),
		HeaderLength:         4,
	}

	if rel == 7 {
		if len(data) < 5 {
			return RecordHeader{}, 0, fmt.Errorf("cdrfile: CDR header truncated (missing release identifier extension)")
		}
		ext := data[4]
		rh.ReleaseIdentifierExtension = &ext
		rh.HeaderLength = 5
	}

	rh.Release = formatRelease(rel, rh.ReleaseIdentifierExtension)
	_ = ver
	return rh, rh.HeaderLength, nil
}

func formatPackedTimestamp(v uint32) string {
	if v == 0 {
		return "0"
	}
	month := (v >> 28) & 0x0f
	day := (v >> 23) & 0x1f
	hour := (v >> 18) & 0x1f
	minute := (v >> 12) & 0x3f

	// The lower 12 bits carry an MS Time Zone (semi-octet octet + 2-bit DST),
	// not a separate UTC sign with hour/minute offset as the spec prose suggests.
	// For 0x35265101 the 5-bit field holds timezone octet 0x04 (+10:00) and
	// the low 2 bits hold DST 0x01 (+1), matching uETimeZone encoding.
	tzOctet := byte((v >> 6) & 0x1f)
	dstOctet := byte(v & 0x03)
	tz, err := decode.FormatMSTimeZoneOctets(tzOctet, dstOctet)
	if err != nil {
		return fmt.Sprintf("%02d%02d%02d%02d", month, day, hour, minute)
	}
	return fmt.Sprintf("%02d%02d%02d%02d%s", month, day, hour, minute, tz)
}

func formatNodeIPAddress(b []byte) string {
	if len(b) != 20 {
		return hex.EncodeToString(b)
	}
	ip := net.IP(b[4:20])
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}
	if ip.IsUnspecified() {
		return hex.EncodeToString(b)
	}
	return ip.String()
}

func releaseExtensionBytes(highRel, lowRel uint8) int {
	n := 0
	if highRel == 7 {
		n++
	}
	if lowRel == 7 {
		n++
	}
	return n
}

func formatRelease(rel uint8, ext *uint8) string {
	if rel == 7 {
		if ext == nil {
			return "beyond Rel-9"
		}
		return fmt.Sprintf("Rel-%d", 10+*ext)
	}
	if name, ok := releaseNames[rel]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", rel)
}

func closureTriggerText(v uint8) string {
	if text, ok := closureTriggers[v]; ok {
		return text
	}
	if v >= 128 {
		return fmt.Sprintf("abnormal(%d)", v)
	}
	return fmt.Sprintf("reserved(%d)", v)
}

func dataRecordFormatText(v uint8) string {
	switch v {
	case 1:
		return "BER"
	case 2:
		return "PER (unaligned)"
	case 3:
		return "PER (aligned)"
	case 4:
		return "XER"
	default:
		return fmt.Sprintf("unknown(%d)", v)
	}
}

func tsNumberText(v uint8) string {
	if text, ok := tsNumbers[v]; ok {
		return text
	}
	return fmt.Sprintf("unknown(%d)", v)
}

var releaseNames = map[uint8]string{
	0: "Rel-99",
	1: "Rel-4",
	2: "Rel-5",
	3: "Rel-6",
	4: "Rel-7",
	5: "Rel-8",
	6: "Rel-9",
}

var closureTriggers = map[uint8]string{
	0:   "normal closure",
	1:   "file size limit reached",
	2:   "file open-time limit reached",
	3:   "maximum number of CDRs reached",
	4:   "manual intervention",
	5:   "CDR release, version or encoding change",
	128: "abnormal closure",
	129: "file system error",
	130: "file system storage exhausted",
	131: "file integrity error",
}

var tsNumbers = map[uint8]string{
	0:  "32.005",
	1:  "32.015",
	2:  "32.205",
	3:  "32.215",
	4:  "32.225",
	5:  "32.235",
	6:  "32.250",
	7:  "32.251",
	9:  "32.260",
	10: "32.270",
	11: "32.271",
	12: "32.272",
	13: "32.273",
	14: "32.275",
	15: "32.274",
	16: "32.277",
	17: "32.296",
	18: "32.278",
	19: "32.253",
	20: "32.255",
	21: "32.254",
	22: "32.256",
	23: "28.201",
	24: "28.202",
	25: "32.257",
}
