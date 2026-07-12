package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/en-vee/asn1x"
	"github.com/en-vee/asn1x/cdrfile"
	"github.com/en-vee/asn1x/decode"
	"github.com/spf13/cobra"
)

type grepOptions struct {
	sharedDecodeOptions
	jsonPath      string
	printMatches  bool
	compact       bool
}

func newGrepCmd() *cobra.Command {
	opts := &grepOptions{}

	cmd := &cobra.Command{
		Use:   "grep <path>",
		Short: "Find CDR files containing a decoded field value",
		Long: `Search BER-encoded CDR files for records whose decoded JSON contains a field value.

Each file under path is decoded using the same options as the decode command (except
JSON output flags). Matching is performed on decoded values, so decode-specs
transformations such as TimeStamp, MSTimeZone, and PLMNID apply.

The --json-path flag uses the form field.path==value. Arrays along the path are
searched; a match in any element satisfies the filter.

Output lines are written as:

  filename:recordNumber

Use --print-matches to also emit the decoded JSON for each matching record.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGrep(opts, args[0])
		},
	}

	bindSharedDecodeFlags(cmd, &opts.sharedDecodeOptions)
	cmd.Flags().StringVar(&opts.jsonPath, "json-path", "", `field filter in the form field.path==value`)
	cmd.Flags().BoolVar(&opts.printMatches, "print-matches", false, "also print decoded JSON for each matching record")
	cmd.Flags().BoolVar(&opts.compact, "compact", false, "emit compact JSON when --print-matches is set")
	_ = cmd.MarkFlagRequired("json-path")

	return cmd
}

func runGrep(opts *grepOptions, searchPath string) error {
	filter, err := decode.ParsePathFilter(opts.jsonPath)
	if err != nil {
		return err
	}

	mod, fieldSpecs, err := loadSchemaAndSpecs(opts.sharedDecodeOptions)
	if err != nil {
		return err
	}

	info, err := os.Stat(searchPath)
	if err != nil {
		return err
	}

	dec := asn1x.NewDecoderWithOptions(mod, asn1x.DecodeOptions{FieldSpecs: fieldSpecs})
	matches := 0

	if !info.IsDir() {
		n, err := grepFile(dec, filter, searchPath, opts)
		if err != nil {
			return err
		}
		matches += n
	} else {
		err = filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			n, err := grepFile(dec, filter, path, opts)
			if err != nil {
				return err
			}
			matches += n
			return nil
		})
		if err != nil {
			return err
		}
	}

	if matches == 0 {
		return fmt.Errorf("no matching CDRs found")
	}
	return nil
}

func grepFile(dec *asn1x.Decoder, filter decode.PathFilter, path string, opts *grepOptions) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return 0, nil
	}

	fileReader := cdrfile.NewReader(data, cdrfile.Options{
		HasFileHeader: opts.fileHeader,
		HasCDRHeader:  opts.cdrHeader,
	})

	if opts.fileHeader {
		if _, err := fileReader.ReadFileHeader(); err != nil {
			return 0, nil
		}
	}

	var enc *json.Encoder
	if opts.printMatches {
		enc = json.NewEncoder(os.Stdout)
		if !opts.compact {
			enc.SetIndent("", "  ")
		}
	}

	matches := 0
	recordNum := 0
	for fileReader.Remaining() > 0 {
		if opts.limit > 0 && recordNum >= opts.limit {
			break
		}

		_, cdrData, err := fileReader.NextRecord()
		if err != nil {
			break
		}
		recordNum++

		val, err := dec.DecodeBytes(opts.rootType, cdrData)
		if err != nil {
			continue
		}
		if !filter.Match(val) {
			continue
		}

		fmt.Printf("%s:%d\n", path, recordNum)
		if opts.printMatches {
			printGrepMatchBanner(path, recordNum)
			if err := enc.Encode(val); err != nil {
				return matches, fmt.Errorf("encode match %s:%d: %w", path, recordNum, err)
			}
		}
		matches++
	}

	return matches, nil
}

func printGrepMatchBanner(path string, recordNum int) {
	fmt.Fprintln(os.Stdout, "-----------------------------")
	fmt.Fprintf(os.Stdout, "CDR #%d (%s)\n", recordNum, path)
	fmt.Fprintln(os.Stdout, "-----------------------------")
}

func loadSchemaAndSpecs(opts sharedDecodeOptions) (*asn1x.Schema, map[string]string, error) {
	schemaFile, err := os.Open(opts.schemaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open schema: %w", err)
	}
	defer schemaFile.Close()

	mod, err := asn1x.Parse(schemaFile)
	if err != nil {
		return nil, nil, fmt.Errorf("parse schema: %w", err)
	}

	var fieldSpecs map[string]string
	if opts.decodeSpecsPath != "" {
		fieldSpecs, err = asn1x.LoadFieldSpecsFile(opts.decodeSpecsPath)
		if err != nil {
			return nil, nil, fmt.Errorf("load decode specs: %w", err)
		}
	}

	return mod, fieldSpecs, nil
}
