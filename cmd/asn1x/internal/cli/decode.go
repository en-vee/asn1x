package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/en-vee/asn1x"
	"github.com/en-vee/asn1x/cdrfile"
	"github.com/spf13/cobra"
)

type decodeOptions struct {
	schemaPath      string
	rootType        string
	limit           int
	compact         bool
	decodeSpecsPath string
	fileHeader      bool
	cdrHeader       bool
}

func newDecodeCmd() *cobra.Command {
	opts := &decodeOptions{}

	cmd := &cobra.Command{
		Use:   "decode <ber-file>",
		Short: "Decode BER-encoded ASN.1 values to JSON",
		Long: `Decode BER-encoded ASN.1 values using a schema and print JSON to stdout.

When the input contains multiple back-to-back values, each record is printed
on its own line (JSONL).

For 3GPP TS 32.297 CDR files, use --file-header and/or --cdr-header to declare
the presence of the file header and per-record CDR headers. Header metadata is
printed to stderr as JSON before each decoded CDR record.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDecode(opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.schemaPath, "schema", "", "path to ASN.1 schema file")
	cmd.Flags().StringVar(&opts.rootType, "type", "", "root type name (e.g. CHFRecord)")
	cmd.Flags().IntVar(&opts.limit, "limit", 0, "maximum number of records to decode (0 = all)")
	cmd.Flags().BoolVar(&opts.compact, "compact", false, "emit compact JSON instead of indented")
	cmd.Flags().StringVar(&opts.decodeSpecsPath, "decode-specs", "", "path to YAML file with per-field decode overrides (qualified fieldPath entries)")
	cmd.Flags().BoolVar(&opts.fileHeader, "file-header", false, "input contains a 3GPP TS 32.297 CDR file header")
	cmd.Flags().BoolVar(&opts.cdrHeader, "cdr-header", false, "each record is prefixed with a 3GPP TS 32.297 CDR record header")

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func runDecode(opts *decodeOptions, berPath string) error {
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
	if opts.decodeSpecsPath != "" {
		fieldSpecs, err = asn1x.LoadFieldSpecsFile(opts.decodeSpecsPath)
		if err != nil {
			return fmt.Errorf("load decode specs: %w", err)
		}
	}

	berFile, err := os.Open(berPath)
	if err != nil {
		return fmt.Errorf("open ber file: %w", err)
	}
	defer berFile.Close()

	data, err := io.ReadAll(berFile)
	if err != nil {
		return fmt.Errorf("read ber file: %w", err)
	}

	dec := asn1x.NewDecoderWithOptions(mod, asn1x.DecodeOptions{FieldSpecs: fieldSpecs})
	enc := json.NewEncoder(os.Stdout)
	if !opts.compact {
		enc.SetIndent("", "  ")
	}
	metaEnc := json.NewEncoder(os.Stderr)

	fileReader := cdrfile.NewReader(data, cdrfile.Options{
		HasFileHeader: opts.fileHeader,
		HasCDRHeader:  opts.cdrHeader,
	})

	if opts.fileHeader {
		fh, err := fileReader.ReadFileHeader()
		if err != nil {
			return fmt.Errorf("read file header: %w", err)
		}
		if err := metaEnc.Encode(map[string]any{
			"headerType": "file",
			"fileHeader": fh,
		}); err != nil {
			return fmt.Errorf("encode file header: %w", err)
		}
	}

	count := 0
	for fileReader.Remaining() > 0 {
		if opts.limit > 0 && count >= opts.limit {
			break
		}

		rh, cdrData, err := fileReader.NextRecord()
		if err != nil {
			return fmt.Errorf("read record %d: %w", count+1, err)
		}

		if opts.cdrHeader {
			if err := metaEnc.Encode(map[string]any{
				"headerType":   "cdr",
				"recordNumber": count + 1,
				"cdrHeader":    rh,
			}); err != nil {
				return fmt.Errorf("encode cdr header: %w", err)
			}
		}

		val, err := dec.DecodeBytes(opts.rootType, cdrData)
		if err != nil {
			return fmt.Errorf("decode record %d: %w", count+1, err)
		}
		if err := enc.Encode(val); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		count++
	}

	if count == 0 {
		return fmt.Errorf("no records decoded")
	}

	fmt.Fprintf(os.Stderr, "decoded %d record(s)\n", count)
	return nil
}
