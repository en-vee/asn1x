package cli

import "github.com/spf13/cobra"

type sharedDecodeOptions struct {
	schemaPath      string
	rootType        string
	limit           int
	decodeSpecsPath string
	fileHeader      bool
	cdrHeader       bool
}

func bindSharedDecodeFlags(cmd *cobra.Command, opts *sharedDecodeOptions) {
	cmd.Flags().StringVar(&opts.schemaPath, "schema", "", "path to ASN.1 schema file")
	cmd.Flags().StringVar(&opts.rootType, "type", "", "root type name (e.g. CHFRecord)")
	cmd.Flags().IntVar(&opts.limit, "limit", 0, "maximum number of records to decode per file (0 = all)")
	cmd.Flags().StringVar(&opts.decodeSpecsPath, "decode-specs", "", "path to YAML file with per-field decode overrides (qualified fieldPath entries)")
	cmd.Flags().BoolVar(&opts.fileHeader, "file-header", true, "input contains a 3GPP TS 32.297 CDR file header")
	cmd.Flags().BoolVar(&opts.cdrHeader, "cdr-header", true, "each record is prefixed with a 3GPP TS 32.297 CDR record header")

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("type")
}
