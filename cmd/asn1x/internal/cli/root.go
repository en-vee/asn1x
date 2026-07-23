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
	Short: "Encode and decode ASN.1 BER data using ASN.1 schema definitions",
	Long: `asn1x parses ASN.1 schema files and converts between BER-encoded data and JSON.

Use decode to transform BER records to JSON, encode to transform JSON to BER, or
grep to find records matching a decoded field value.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(newDecodeCmd())
	rootCmd.AddCommand(newEncodeCmd())
	rootCmd.AddCommand(newGrepCmd())
	rootCmd.AddCommand(newVersionCmd())
}
