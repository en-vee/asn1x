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
	sharedDecodeOptions
	compact            bool
	suppressFileHeader bool
	suppressCDRHeader  bool
}

func newDecodeCmd() *cobra.Command {
	opts := &decodeOptions{}

	cmd := &cobra.Command{
		Use:   "decode <ber-file>",
		Short: "Decode BER-encoded ASN.1 values to JSON",
		Long: `Decode BER-encoded ASN.1 values using a schema and print JSON to stdout.

When the input contains multiple records, each is preceded by a CDR #N banner
followed by indented or compact JSON.

For 3GPP TS 32.297 CDR files, --file-header and --cdr-header declare the input
framing (both default to true). Use --suppress-file-header and/or
--suppress-cdr-header to skip printing parsed header metadata while still
skipping those bytes during decode.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDecode(opts, args[0])
		},
	}

	bindSharedDecodeFlags(cmd, &opts.sharedDecodeOptions)
	cmd.Flags().BoolVar(&opts.compact, "compact", false, "emit compact JSON instead of indented")
	cmd.Flags().BoolVar(&opts.suppressFileHeader, "suppress-file-header", false, "do not print parsed file header metadata")
	cmd.Flags().BoolVar(&opts.suppressCDRHeader, "suppress-cdr-header", false, "do not print parsed CDR record header metadata")

	return cmd
}

func runDecode(opts *decodeOptions, berPath string) error {
	mod, fieldSpecs, err := loadSchemaAndSpecs(opts.sharedDecodeOptions)
	if err != nil {
		return err
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

	fileReader := cdrfile.NewReader(data, cdrfile.Options{
		HasFileHeader: opts.fileHeader,
		HasCDRHeader:  opts.cdrHeader,
	})

	if opts.fileHeader {
		fh, err := fileReader.ReadFileHeader()
		if err != nil {
			return fmt.Errorf("read file header: %w", err)
		}
		if !opts.suppressFileHeader {
			if err := enc.Encode(map[string]any{
				"headerType": "file",
				"fileHeader": fh,
			}); err != nil {
				return fmt.Errorf("encode file header: %w", err)
			}
		}
	}

	count := 0
	for fileReader.Remaining() > 0 {
		if opts.limit > 0 && count >= opts.limit {
			break
		}

		printCDRBanner(count + 1)

		rh, cdrData, err := fileReader.NextRecord()
		if err != nil {
			return fmt.Errorf("read record %d: %w", count+1, err)
		}

		if opts.cdrHeader && !opts.suppressCDRHeader {
			if err := enc.Encode(map[string]any{
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

func printCDRBanner(n int) {
	fmt.Fprintln(os.Stdout, "-----------------------------")
	fmt.Fprintf(os.Stdout, "CDR #%d\n", n)
	fmt.Fprintln(os.Stdout, "-----------------------------")
}
