package cli

import (
	"github.com/spf13/cobra"
)

// Execute runs the asn1x CLI.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "asn1x",
	Short: "Decode ASN.1 BER data using ASN.1 schema definitions",
	Long: `asn1x parses ASN.1 schema files and decodes BER-encoded data into JSON.

Use the decode command to transform one or more concatenated BER records, or grep
to find records matching a decoded field value.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(newDecodeCmd())
	rootCmd.AddCommand(newGrepCmd())
}
