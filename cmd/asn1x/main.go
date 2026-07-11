package main

import (
	"os"

	"github.com/en-vee/asn1x/cmd/asn1x/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
