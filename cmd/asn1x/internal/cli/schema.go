package cli

import (
	"fmt"
	"os"

	"github.com/en-vee/asn1x"
)

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
	if opts.specOverridesPath != "" {
		fieldSpecs, err = asn1x.LoadFieldSpecsFile(opts.specOverridesPath)
		if err != nil {
			return nil, nil, fmt.Errorf("load spec overrides: %w", err)
		}
	}

	return mod, fieldSpecs, nil
}
