package cli

import (
	"runtime/debug"
	"testing"
)

func TestFormatGoVersion(t *testing.T) {
	if got := formatGoVersion("go1.22.5"); got != "1.22.5" {
		t.Fatalf("formatGoVersion() = %q, want %q", got, "1.22.5")
	}
}

func TestResolveVersionInfoDefaults(t *testing.T) {
	version, commit, goVersion := resolveVersionInfo()
	if version == "" {
		t.Fatal("expected non-empty version")
	}
	if commit == "" {
		t.Fatal("expected non-empty commit")
	}
	if goVersion == "" {
		t.Fatal("expected non-empty go version")
	}
}

func TestVCSRevision(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("build info unavailable")
	}

	rev := vcsRevision(info)
	if rev == "unknown" {
		t.Skip("vcs revision not embedded in build")
	}
	if len(rev) > 7 {
		t.Fatalf("vcsRevision() = %q, want at most 7 chars", rev)
	}
}
