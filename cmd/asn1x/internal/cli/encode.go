package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/en-vee/asn1x"
	"github.com/spf13/cobra"
)

type encodeOptions struct {
	schemaPath        string
	rootType          string
	specOverridesPath string
	outputPath        string
	limit             int
}

func newEncodeCmd() *cobra.Command {
	opts := &encodeOptions{}

	cmd := &cobra.Command{
		Use:   "encode [json-file]",
		Short: "Encode JSON to BER-encoded ASN.1",
		Long: `Encode one or more JSON values to BER using an ASN.1 schema.

JSON may be read from a file argument or from stdin when no file is given.

Input formats:
  - a single JSON object (one record)
  - a JSON array of objects (multiple records)
  - JSON Lines (NDJSON): one JSON object per line

Each record is encoded with the given root type and written back-to-back as
concatenated BER. Use --output to write to a file; otherwise BER is written
to stdout.

The same --spec-overrides YAML used for decoding also drives encode wire formats
for overridden fields (TimeStamp, PLMNID, MSTimeZone, etc.).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonPath string
			if len(args) == 1 {
				jsonPath = args[0]
			}
			return runEncode(opts, jsonPath)
		},
	}

	cmd.Flags().StringVar(&opts.schemaPath, "schema", "", "path to ASN.1 schema file")
	cmd.Flags().StringVar(&opts.rootType, "type", "", "root type name (e.g. CHFRecord)")
	cmd.Flags().StringVar(&opts.specOverridesPath, "spec-overrides", "", "path to YAML file with per-field wire-type overrides (same format as decode)")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "", "write BER output to this file (default: stdout)")
	cmd.Flags().IntVar(&opts.limit, "limit", 0, "maximum number of records to encode (0 = all)")
	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func runEncode(opts *encodeOptions, jsonPath string) error {
	schemaFile, err := os.Open(opts.schemaPath)
	if err != nil {
		return fmt.Errorf("open schema: %w", err)
	}
	defer schemaFile.Close()

	mod, err := asn1x.Parse(schemaFile)
	if err != nil {
		return fmt.Errorf("parse schema: %w", err)
	}

	var fieldSpecs map[string]string
	if opts.specOverridesPath != "" {
		fieldSpecs, err = asn1x.LoadFieldSpecsFile(opts.specOverridesPath)
		if err != nil {
			return fmt.Errorf("load spec overrides: %w", err)
		}
	}

	jsonBytes, err := readJSONInput(jsonPath)
	if err != nil {
		return err
	}

	records, err := parseEncodeRecords(jsonBytes)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("no records to encode")
	}
	if opts.limit > 0 && len(records) > opts.limit {
		records = records[:opts.limit]
	}

	enc := asn1x.NewEncoderWithOptions(mod, asn1x.EncodeOptions{FieldSpecs: fieldSpecs})

	out, closer, err := openEncodeOutput(opts.outputPath)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer()
	}

	for i, rec := range records {
		berBytes, err := enc.EncodeBytes(opts.rootType, rec)
		if err != nil {
			return fmt.Errorf("encode record %d: %w", i+1, err)
		}
		if _, err := out.Write(berBytes); err != nil {
			return fmt.Errorf("write ber record %d: %w", i+1, err)
		}
	}

	fmt.Fprintf(os.Stderr, "encoded %d record(s)\n", len(records))
	return nil
}

func openEncodeOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, nil, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("create output file: %w", err)
	}
	return f, func() { _ = f.Close() }, nil
}

func readJSONInput(path string) ([]byte, error) {
	if path == "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read json file: %w", err)
	}
	return data, nil
}

// parseEncodeRecords accepts a single JSON object, a JSON array of objects,
// or JSON Lines (one object per non-empty line).
func parseEncodeRecords(data []byte) ([]any, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty JSON input")
	}

	switch trimmed[0] {
	case '[':
		var arr []any
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return nil, fmt.Errorf("parse JSON array: %w", err)
		}
		for i, item := range arr {
			if _, ok := item.(map[string]any); !ok {
				return nil, fmt.Errorf("JSON array element %d must be an object, got %T", i+1, item)
			}
		}
		return arr, nil
	case '{':
		// Prefer a single object when the whole input is one JSON value.
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		var first any
		if err := dec.Decode(&first); err != nil {
			return nil, fmt.Errorf("parse JSON: %w", err)
		}
		if _, ok := first.(map[string]any); !ok {
			return nil, fmt.Errorf("JSON value must be an object, got %T", first)
		}
		if !dec.More() {
			return []any{first}, nil
		}
		// Multiple top-level JSON values (JSON sequence / concatenated objects).
		records := []any{first}
		for dec.More() {
			var next any
			if err := dec.Decode(&next); err != nil {
				return nil, fmt.Errorf("parse JSON record %d: %w", len(records)+1, err)
			}
			if _, ok := next.(map[string]any); !ok {
				return nil, fmt.Errorf("JSON record %d must be an object, got %T", len(records)+1, next)
			}
			records = append(records, next)
		}
		return records, nil
	default:
		// JSON Lines: one object per non-empty line.
		return parseJSONLines(trimmed)
	}
}

func parseJSONLines(data []byte) ([]any, error) {
	lines := bytes.Split(data, []byte("\n"))
	var records []any
	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var v any
		if err := json.Unmarshal(line, &v); err != nil {
			return nil, fmt.Errorf("parse JSON line %d: %w", i+1, err)
		}
		if _, ok := v.(map[string]any); !ok {
			return nil, fmt.Errorf("JSON line %d must be an object, got %T", i+1, v)
		}
		records = append(records, v)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no JSON objects found")
	}
	return records, nil
}
