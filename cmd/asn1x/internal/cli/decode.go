package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/en-vee/asn1x"
	"github.com/spf13/cobra"
)

type decodeOptions struct {
	schemaPath      string
	rootType        string
	limit           int
	compact         bool
	decodeSpecsPath string
}

func newDecodeCmd() *cobra.Command {
	opts := &decodeOptions{}

	cmd := &cobra.Command{
		Use:   "decode <ber-file>",
		Short: "Decode BER-encoded ASN.1 values to JSON",
		Long: `Decode BER-encoded ASN.1 values using a schema and print JSON to stdout.

When the input contains multiple back-to-back values, each record is printed
on its own line (JSONL).`,
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

	count := 0
	for len(data) > 0 {
		if opts.limit > 0 && count >= opts.limit {
			break
		}
		val, rest, err := dec.DecodeNext(opts.rootType, data)
		if err != nil {
			return fmt.Errorf("decode record %d: %w", count+1, err)
		}
		if err := enc.Encode(val); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		data = rest
		count++
	}

	if count == 0 {
		return fmt.Errorf("no records decoded")
	}

	fmt.Fprintf(os.Stderr, "decoded %d record(s)\n", count)
	return nil
}
