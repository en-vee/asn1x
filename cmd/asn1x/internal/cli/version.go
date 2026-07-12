package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

const author = "Neeraj Vaidya"

// Version and GitCommit are set at build time via -ldflags.
var (
	Version   = "dev"
	GitCommit = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print asn1x version information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}

func printVersion() {
	version, commit, goVersion := resolveVersionInfo()

	fmt.Println("-------------------------------------")
	fmt.Println("asn1x - asn.1 ber utility")
	fmt.Printf("Version : %s\n", version)
	fmt.Printf("Author : %s\n", author)
	fmt.Println("--------------------------------------")
	fmt.Printf("GoVersion : %s\n", goVersion)
	fmt.Printf("GitCommit : %s\n", commit)
}

func resolveVersionInfo() (version, commit, goVersion string) {
	version = Version
	commit = GitCommit
	goVersion = formatGoVersion(runtime.Version())

	if info, ok := debug.ReadBuildInfo(); ok {
		goVersion = formatGoVersion(info.GoVersion)
		if commit == "unknown" {
			commit = vcsRevision(info)
		}
	}

	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	return version, commit, goVersion
}

func formatGoVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "go")
}

func vcsRevision(info *debug.BuildInfo) string {
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			rev := setting.Value
			if len(rev) > 7 {
				return rev[:7]
			}
			return rev
		}
	}
	return "unknown"
}
