package cdrfile_test

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/en-vee/asn1x/ber"
	"github.com/en-vee/asn1x/cdrfile"
)

func buildMinimalFileHeader(t *testing.T, numCDRs uint32) []byte {
	t.Helper()
	hdr := make([]byte, 50)
	binary.BigEndian.PutUint32(hdr[0:4], 1000)
	binary.BigEndian.PutUint32(hdr[4:8], 50)
	hdr[8] = (6 << 5) | 17 // Rel-9, version 17
	hdr[9] = (6 << 5) | 17
	binary.BigEndian.PutUint32(hdr[18:22], numCDRs)
	binary.BigEndian.PutUint16(hdr[48:50], 0)
	return hdr
}

func loadFirstCHFRecord(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1")
	if err != nil {
		t.Skip(err)
	}
	_, rest, err := ber.ReadOne(data)
	if err != nil {
		t.Fatal(err)
	}
	return data[:len(data)-len(rest)]
}

func buildRecordHeader(t *testing.T, cdr []byte) []byte {
	t.Helper()
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint16(hdr[0:2], uint16(len(cdr)))
	hdr[2] = (6 << 5) | 17 // Rel-9, version 17
	hdr[3] = (1 << 5) | 7  // BER, TS 32.251
	return hdr
}

func TestParseFileHeaderMinimal(t *testing.T) {
	hdr := buildMinimalFileHeader(t, 3)
	fh, n, err := cdrfile.ParseFileHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if n != 50 {
		t.Fatalf("consumed %d bytes, want 50", n)
	}
	if fh.HeaderLength != 50 {
		t.Fatalf("headerLength = %d", fh.HeaderLength)
	}
	if fh.NumberOfCDRs != 3 {
		t.Fatalf("numberOfCDRs = %d", fh.NumberOfCDRs)
	}
	if fh.HighRelease != "Rel-9" {
		t.Fatalf("highRelease = %q", fh.HighRelease)
	}
}

func TestParseRecordHeader(t *testing.T) {
	cdr := loadFirstCHFRecord(t)
	raw := buildRecordHeader(t, cdr)
	rh, n, err := cdrfile.ParseRecordHeader(raw)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("consumed %d bytes, want 4", n)
	}
	if int(rh.CDRLength) != len(cdr) {
		t.Fatalf("cdrLength = %d, want %d", rh.CDRLength, len(cdr))
	}
	if rh.DataRecordFormatText != "BER" {
		t.Fatalf("format = %q", rh.DataRecordFormatText)
	}
	if rh.TSNumberText != "32.251" {
		t.Fatalf("tsNumber = %q", rh.TSNumberText)
	}
}

func TestReaderFileHeaderOnly(t *testing.T) {
	cdr := loadFirstCHFRecord(t)
	data := append(buildMinimalFileHeader(t, 1), cdr...)

	r := cdrfile.NewReader(data, cdrfile.Options{HasFileHeader: true})
	fh, err := r.ReadFileHeader()
	if err != nil {
		t.Fatal(err)
	}
	if fh.NumberOfCDRs != 1 {
		t.Fatalf("numberOfCDRs = %d", fh.NumberOfCDRs)
	}

	_, got, err := r.NextRecord()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(cdr) {
		t.Fatalf("cdr length = %d, want %d", len(got), len(cdr))
	}
}

func TestReaderCDRHeaderOnly(t *testing.T) {
	cdr := loadFirstCHFRecord(t)
	data := append(buildRecordHeader(t, cdr), cdr...)

	r := cdrfile.NewReader(data, cdrfile.Options{HasCDRHeader: true})
	rh, got, err := r.NextRecord()
	if err != nil {
		t.Fatal(err)
	}
	if rh == nil {
		t.Fatal("expected record header")
	}
	if len(got) != len(cdr) {
		t.Fatalf("cdr length = %d, want %d", len(got), len(cdr))
	}
}

func TestReaderFileAndCDRHeader(t *testing.T) {
	cdr := loadFirstCHFRecord(t)
	record := append(buildRecordHeader(t, cdr), cdr...)
	data := append(buildMinimalFileHeader(t, 1), record...)

	r := cdrfile.NewReader(data, cdrfile.Options{HasFileHeader: true, HasCDRHeader: true})
	if _, err := r.ReadFileHeader(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.NextRecord(); err != nil {
		t.Fatal(err)
	}
}

func TestParseRecordHeaderWithReleaseExtension(t *testing.T) {
	hdr := []byte{0x00, 0x10, 0xE0 | 17, 0x27, 3} // rel=7, ext=3 => Rel-13
	rh, n, err := cdrfile.ParseRecordHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("consumed %d bytes, want 5", n)
	}
	if rh.Release != "Rel-13" {
		t.Fatalf("release = %q", rh.Release)
	}
}

func TestParseFileHeaderWithReleaseExtensionsOnly(t *testing.T) {
	data, err := os.ReadFile("../../asn1-cdr-util/sample-files/PNY9_CGF01_-_000005.20220503_-_190520+1000")
	if err != nil {
		t.Skip(err)
	}

	fh, n, err := cdrfile.ParseFileHeader(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 68 {
		t.Fatalf("consumed %d bytes, want 68", n)
	}
	if fh.CDRRoutingFilter != "CHF:CAF:ASN1:SNK" {
		t.Fatalf("routing filter = %q", fh.CDRRoutingFilter)
	}
	if fh.HighRelease != "Rel-15" || fh.LowRelease != "Rel-15" {
		t.Fatalf("release range = %q..%q", fh.HighRelease, fh.LowRelease)
	}
	if fh.PrivateExtensionLength != 0 {
		t.Fatalf("private extension length = %d", fh.PrivateExtensionLength)
	}
	if fh.FileOpeningTimestamp != "03100937+10:00+1" {
		t.Fatalf("fileOpeningTimestamp = %q", fh.FileOpeningTimestamp)
	}
}
