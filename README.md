# asn1x

Go library and command-line utility for parsing ASN.1 schema definitions and converting between BER-encoded ASN.1 data and JSON using those schemas.

## Features

- Parse ASN.1 module definitions (`.asn`, `.EXP`, and similar syntax files)
- Schema-driven BER decoding and encoding with field names taken from the schema
- JSON mapping with sensible type conversions (SEQUENCE/SET → object, SEQUENCE OF/SET OF → array, CHOICE → single-key object)
- Optional per-field wire-type overrides via YAML (`fieldPath` + `asn1DataType`) for cases where the on-wire encoding does not match the schema type (used by both decode and encode)
- `grep` subcommand to find CDR records matching a decoded field value across files or directories
- `asn1x` CLI built with [Cobra](https://github.com/spf13/cobra)

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

### Encode JSON to BER

```bash
asn1x encode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --output records.ber \
  records.json
```

JSON may also be read from stdin when no file argument is given.

**Input formats:**

- A single JSON object (one record)
- A JSON array of objects (multiple records), encoded back-to-back as concatenated BER
- JSON Lines / NDJSON (one object per line)

Example multi-record array:

```json
[
  { "chargingFunctionRecord": { "...": "..." } },
  { "chargingFunctionRecord": { "...": "..." } }
]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--schema` | yes | Path to the ASN.1 schema file |
| `--type` | yes | Root type name to encode (e.g. `CHFRecord`) |
| `--spec-overrides` | no | Path to a YAML file with per-field wire-type overrides (same file as decode) |
| `--output` / `-o` | no | Write BER to this file (default: stdout) |
| `--limit` | no | Maximum number of records to encode (`0` = all) |

**Output:** Concatenated BER records are written to `--output`, or to **stdout** if omitted. Progress is reported on stderr (`encoded N record(s)`).

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
| `--spec-overrides` | no | Path to a YAML file with per-field wire-type overrides |
| `--limit` | no | Maximum number of records to decode (`0` = all) |
| `--compact` | no | Emit compact JSON instead of indented output |
| `--file-header` | no | Input contains a 3GPP TS 32.297 CDR file header (default: `true`) |
| `--cdr-header` | no | Each record is prefixed with a 3GPP TS 32.297 CDR record header (default: `true`) |
| `--suppress-file-header` | no | Do not print parsed file header metadata (default: `false`) |
| `--suppress-cdr-header` | no | Do not print parsed CDR record header metadata (default: `false`) |

The BER file is passed as a positional argument.

**Output:** Header metadata (when present and not suppressed), a `CDR #N` banner before each record, and CDR JSON are written to **stdout**. Progress is reported on stderr (`decoded N record(s)`).

`--file-header` and `--cdr-header` control **input framing** (whether those bytes are skipped before decode). `--suppress-*` controls **printing only** — headers are still parsed and consumed when framing is enabled.

For raw back-to-back BER records (no 3GPP file wrapper), pass `--file-header=false --cdr-header=false`. These flags do not apply when decoding individual BER records from a Kafka topic or similar stream.

### Examples

Decode the first record only (raw BER stream, no 3GPP file headers):

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --file-header=false \
  --cdr-header=false \
  --limit 1 \
  sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1
```

Decode all records to a JSONL file with field decode overrides:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  sample-asn1-files/vvsl22183_-_87150.20220429_._1113+1000.asn1 > records.jsonl
```

Decode a 3GPP TS 32.297 CDR file (default framing; headers printed to stdout):

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  cdr-file.dat
```

Decode a CDR file but omit header metadata from the output:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --suppress-file-header=true \
  --suppress-cdr-header=true \
  cdr-file.dat
```

Decode a CDR file with a file header but no per-record CDR headers:

```bash
asn1x decode \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --cdr-header=false \
  cdr-file.dat
```

Show help:

```bash
asn1x --help
asn1x decode --help
asn1x encode --help
```

### Search CDR files (`grep`)

Find records whose **decoded** JSON contains a specific field value. Matching uses the same schema and spec-overrides as `decode`, so values like `TimeStamp`, `MSTimeZone`, and `PLMNID` are compared after transformation — not as raw BER bytes.

Pass a **single file** or a **directory** as the positional argument. Directories are searched recursively.

```bash
asn1x grep \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --json-path 'chargingFunctionRecord.listOfMultipleUnitUsage.usedUnitContainers.pDUContainerInformation.uETimeZone==+10:00+1' \
  sample-asn1-files
```

| Flag | Required | Description |
|------|----------|-------------|
| `--schema` | yes | Path to the ASN.1 schema file |
| `--type` | yes | Root type name to decode (e.g. `CHFRecord`) |
| `--json-path` | yes | Field filter in the form `field.path==value` |
| `--spec-overrides` | no | Path to a YAML file with per-field wire-type overrides |
| `--limit` | no | Maximum number of records to scan per file (`0` = all) |
| `--file-header` | no | Input contains a 3GPP TS 32.297 CDR file header (default: `true`) |
| `--cdr-header` | no | Each record is prefixed with a 3GPP TS 32.297 CDR record header (default: `true`) |
| `--print-matches` | no | Also print decoded JSON for each matching record (default: `false`) |
| `--compact` | no | Emit compact JSON when `--print-matches` is set |

**Filter syntax:** `--json-path` uses `field.path==value`. Path segments match decoded JSON field names. Arrays along the path are searched — a match in any element satisfies the filter.

**Output:** One line per match on stdout:

```text
filename:recordNumber
```

Exit status is non-zero when no matches are found.

#### Examples

Search a single CDR file for a specific duration:

```bash
asn1x grep \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --file-header=false \
  --cdr-header=false \
  --json-path 'chargingFunctionRecord.duration==148' \
  sample-asn1-files/iot-1.ber
```

Search a 3GPP CDR file for records with a given SMF trigger:

```bash
asn1x grep \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --json-path 'chargingFunctionRecord.listOfMultipleUnitUsage.usedUnitContainers.triggers.sMFTrigger==endOfPDUSession' \
  sample-asn1-files/iot-sftp-sink-m3c1-0_-_000001.20260709_-_125346+1000
```

Search by PLMN (requires `PLMNID` entries in spec-overrides):

```bash
asn1x grep \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --json-path 'chargingFunctionRecord.nFunctionConsumerInformation.networkFunctionPLMNIdentifier.mcc==505' \
  sample-asn1-files
```

Print matching records as JSON:

```bash
asn1x grep \
  --schema schema/testdata/CHFChargingDataTypes.EXP \
  --type CHFRecord \
  --spec-overrides decode/testdata/chf-decode-specs.yaml \
  --json-path 'chargingFunctionRecord.duration==148' \
  --print-matches \
  sample-asn1-files/iot-1.ber
```

Show help:

```bash
asn1x grep --help
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

### Encode a Go value or JSON to BER

```go
schema, _ := asn1x.Parse(schemaFile)
specs, _ := asn1x.LoadFieldSpecsFile("decode/testdata/chf-decode-specs.yaml")

enc := asn1x.NewEncoderWithOptions(schema, asn1x.EncodeOptions{
    FieldSpecs: specs,
})

// From a decoded (or hand-built) Go value:
berBytes, err := enc.EncodeBytes("CHFRecord", val)
if err != nil {
    panic(err)
}

// Or directly from JSON:
berBytes, err = enc.EncodeJSON("CHFRecord", jsonBytes)
```

### Per-field wire-type overrides

Some deployments encode values differently from what the schema declares (for example, timestamps carried as 9-byte BCD `TimeStamp` octet strings instead of BER `GeneralizedTime`). Use a YAML specs file and pass it to the decoder and encoder:

```go
specs, err := asn1x.LoadFieldSpecsFile("decode/testdata/chf-decode-specs.yaml")
if err != nil {
    panic(err)
}

dec := asn1x.NewDecoderWithOptions(schema, asn1x.DecodeOptions{
    FieldSpecs: specs,
})

enc := asn1x.NewEncoderWithOptions(schema, asn1x.EncodeOptions{
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
  ber/           BER TLV parsing and encoding
  decode/        Schema-driven decoder and JSON output
  encode/        Schema-driven JSON/Go-value → BER encoder
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
- Information object references (e.g. `DMI-EXTENSION.&id`) are decoded/encoded with best-effort fallback
- Extension fields decoded as `Unknown_<tag>` are not re-encoded
- CDR file header wrapping is not applied on encode (encode emits raw BER records)
