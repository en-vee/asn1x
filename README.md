# asn1x

Go library and command-line utility for parsing ASN.1 schema definitions and decoding BER-encoded ASN.1 data into JSON using those schemas.

## Features

- Parse ASN.1 module definitions (`.asn`, `.EXP`, and similar syntax files)
- Schema-driven BER decoding with field names taken from the schema
- JSON output with sensible type mappings (SEQUENCE/SET → object, SEQUENCE OF/SET OF → array, CHOICE → single-key object)
- Optional per-field decode overrides via YAML (`fieldPath` + `asn1DataType`) for cases where the on-wire encoding does not match the schema type
- `asn1x` CLI built with [Cobra](https://github.com/spf13/cobra)

JSON-to-ASN.1 encoding is not implemented yet.

## Requirements

- Go 1.22 or later

## Building the `asn1x` utility

From the repository root:

```bash
go build -o asn1x ./cmd/asn1x
```

Install into `$GOPATH/bin` or `$GOBIN`:

```bash
go install ./cmd/asn1x
```

## CLI usage

### Decode BER data to JSON

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1
```

| Flag | Required | Description |
|------|----------|-------------|
| `--schema` | yes | Path to the ASN.1 schema file |
| `--type` | yes | Root type name to decode (e.g. `CHFRecord`) |
| `--decode-specs` | no | Path to a YAML file with per-field decode overrides |
| `--limit` | no | Maximum number of records to decode (`0` = all) |
| `--compact` | no | Emit compact JSON instead of indented output |
| `--file-header` | no | Input contains a 3GPP TS 32.297 CDR file header (clause 6.1.1) |
| `--cdr-header` | no | Each record is prefixed with a 3GPP TS 32.297 CDR record header (clause 6.1.2) |

The BER file is passed as a positional argument.

**Output:** CDR JSON is written to stdout. When the input contains multiple records, each record is printed on its own line (JSONL). Progress is reported on stderr (`decoded N record(s)`).

When `--file-header` and/or `--cdr-header` are set, header metadata is printed to stderr as JSON before the corresponding CDR output. These flags are intended for 3GPP CDR **files** only; they do not apply when decoding individual BER records from a Kafka topic or similar stream.

### Examples

Decode the first record only:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --limit 1 \
  sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1
```

Decode all records to a JSONL file with field decode overrides:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --decode-specs decode/testdata/chf-decode-specs.yaml \
  sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1 > records.jsonl
```

Decode a 3GPP TS 32.297 CDR file with a file header but no per-record CDR headers:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --file-header=true \
  --cdr-header=false \
  cdr-file.dat
```

Decode a CDR file with both file and record headers (header JSON on stderr, CDR JSON on stdout):

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --file-header=true \
  --cdr-header=true \
  cdr-file.dat 2> headers.jsonl > records.jsonl
```

Show help:

```bash
asn1x --help
asn1x decode --help
```

## Library usage

Add the module to your project:

```bash
go get github.com/en-vee/asn1x
```

### Parse a schema

```go
package main

import (
    "fmt"
    "os"

    "github.com/en-vee/asn1x"
)

func main() {
    f, err := os.Open("schema/testdata/CHFChargingDataTypes.EXP")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    schema, err := asn1x.Parse(f)
    if err != nil {
        panic(err)
    }

    fmt.Println("module:", schema.ModuleName)
    fmt.Println("types:", len(schema.Types))

    typ, ok := schema.Lookup("CHFRecord")
    if ok {
        fmt.Println("CHFRecord kind:", typ.TypeKind())
    }
}
```

### Decode BER bytes to a Go value

```go
schema, _ := asn1x.Parse(schemaFile)

data, _ := os.ReadFile("record.asn1")

dec := asn1x.NewDecoder(schema)
val, err := dec.DecodeBytes("CHFRecord", data)
if err != nil {
    panic(err)
}

jsonBytes, err := asn1x.ToJSON(val)
if err != nil {
    panic(err)
}
fmt.Println(string(jsonBytes))
```

### Decode a stream of concatenated records

```go
data, _ := os.ReadFile("records.asn1")

dec := asn1x.NewDecoder(schema)
for len(data) > 0 {
    var val any
    var err error
    val, data, err = dec.DecodeNext("CHFRecord", data)
    if err != nil {
        panic(err)
    }

    jsonBytes, _ := asn1x.ToJSON(val)
    fmt.Println(string(jsonBytes))
}
```

### Per-field decode overrides

Some deployments encode values differently from what the schema declares (for example, timestamps carried as 9-byte BCD `TimeStamp` octet strings instead of BER `GeneralizedTime`). Use a YAML specs file and pass it to the decoder:

```go
specs, err := asn1x.LoadFieldSpecsFile("decode/testdata/chf-decode-specs.yaml")
if err != nil {
    panic(err)
}

dec := asn1x.NewDecoderWithOptions(schema, asn1x.DecodeOptions{
    FieldSpecs: specs,
})

val, err := dec.DecodeBytes("CHFRecord", data)
```

YAML format:

```yaml
asn1x:
  decodeSpecs:
    - fieldPath: chargingFunctionRecord.recordOpeningTime
      asn1DataType: TimeStamp
    - fieldPath: chargingFunctionRecord.listOfMultipleUnitUsage.usedUnitContainers.pDUContainerInformation.timeOfFirstUsage
      asn1DataType: TimeStamp
```

Each `fieldPath` must be a **qualified** dot-separated path matching the JSON structure (at least one `.` segment is required). Supported `asn1DataType` values include:

| Type | Behavior |
|------|----------|
| `TimeStamp` | 3GPP 9-byte BCD timestamp (`YYMMDDhhmmssShhmm`), or ASCII time text; output RFC3339 UTC |
| `UTCTime` | Parse UTCTime or GeneralizedTime text, or 9-byte BCD TimeStamp; output RFC3339 UTC |
| `GeneralizedTime` | Parse ISO-8601 / GeneralizedTime text, or 9-byte BCD TimeStamp; output RFC3339 UTC |
| `IA5String`, `UTF8String`, … | Force string decoding |
| `Integer` | Force BER integer decoding |
| `OctetString` | Default octet-string handling (text vs hex heuristic) |

A sample CHF specs file is included at `decode/testdata/chf-decode-specs.yaml`.

## JSON mapping

| ASN.1 type | JSON representation |
|------------|---------------------|
| `SEQUENCE`, `SET` | Object with schema field names as keys |
| `SEQUENCE OF`, `SET OF` | Array |
| `CHOICE` | Object with a single key (the chosen alternative name) |
| `INTEGER` / `ENUMERATED` | Symbolic name when defined in the schema, otherwise a number |
| `BOOLEAN` | `true` / `false` |
| `NULL` | `null` |
| `OCTET STRING` | Hex string for binary data; plain string when content is printable ASCII |
| Character string types (`IA5String`, `UTF8String`, …) | String |

Optional components omitted from the BER encoding are omitted from the JSON output.

## Project layout

```
asn1x/
  ber/           BER TLV parsing
  decode/        Schema-driven decoder and JSON output
  schema/        ASN.1 schema parser
  cmd/asn1x/     asn1x CLI (Cobra)
  sample-asn1-files/   Sample BER data files
```

## Running tests

```bash
go test ./...
```

## Limitations

- Indefinite-length BER is not supported
- Information object references (e.g. `DMI-EXTENSION.&id`) are decoded with best-effort fallback
- JSON-to-ASN.1 encoding is not implemented
