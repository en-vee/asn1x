package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/en-vee/asn1x/search"
	"github.com/spf13/cobra"
)

type grepOptions struct {
	sharedDecodeOptions
	jsonPath     string
	printMatches bool
	compact      bool
}

func newGrepCmd() *cobra.Command {
	opts := &grepOptions{}

	cmd := &cobra.Command{
		Use:   "grep <path>",
		Short: "Find CDR files containing a decoded field value",
		Long: `Search BER-encoded CDR files for records whose decoded JSON contains a field value.

Each file under path is decoded using the same options as the decode command (except
JSON output flags). Matching is performed on decoded values, so spec-overrides
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
	if err := requireSchemaAndType(opts.sharedDecodeOptions); err != nil {
		return err
	}

	searcher, err := search.NewSearcher(search.Config{
		SchemaPath:        opts.schemaPath,
		RootType:          opts.rootType,
		SpecOverridesPath: opts.specOverridesPath,
		FileHeader:        opts.fileHeader,
		CDRHeader:         opts.cdrHeader,
	})
	if err != nil {
		return err
	}

	matches, err := searcher.Grep(search.GrepOptions{
		Path:           searchPath,
		JSONPath:       opts.jsonPath,
		Limit:          opts.limit,
		IncludeRecords: opts.printMatches,
	})
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no matching CDRs found")
	}

	var enc *json.Encoder
	if opts.printMatches {
		enc = json.NewEncoder(os.Stdout)
		if !opts.compact {
			enc.SetIndent("", "  ")
		}
	}

	for _, match := range matches {
		fmt.Printf("%s:%d\n", match.File, match.RecordNum)
		if opts.printMatches {
			printGrepMatchBanner(match.File, match.RecordNum)
			if err := enc.Encode(match.Record); err != nil {
				return fmt.Errorf("encode match %s:%d: %w", match.File, match.RecordNum, err)
			}
		}
	}
	return nil
}

func printGrepMatchBanner(path string, recordNum int) {
	fmt.Fprintln(os.Stdout, "-----------------------------")
	fmt.Fprintf(os.Stdout, "CDR #%d (%s)\n", recordNum, path)
	fmt.Fprintln(os.Stdout, "-----------------------------")
}
