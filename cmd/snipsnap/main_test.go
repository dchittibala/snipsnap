package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/dchittibala/snipsnap/pkg/fileops"
)

// runCLI invokes the CLI entrypoint exactly as main() does and returns the exit code.
func runCLI(t *testing.T, args ...string) int {
	t.Helper()
	return run(context.Background(), args, fileops.New(), io.Discard)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestIntegration_HappyPath exercises the full create -> copy -> combine -> delete
// flow end to end through the CLI argument parser and real file operations.
func TestIntegration_HappyPath(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	cp := filepath.Join(dir, "copy.txt")
	combined := filepath.Join(dir, "combined.txt")

	if code := runCLI(t, "create", "-text=hello", a); code != ExitSuccess {
		t.Fatalf("create a: exit %d", code)
	}
	if got := readFile(t, a); got != "hello" {
		t.Fatalf("create a content = %q, want %q", got, "hello")
	}

	if code := runCLI(t, "create", b); code != ExitSuccess {
		t.Fatalf("create empty b: exit %d", code)
	}
	if got := readFile(t, b); got != "" {
		t.Fatalf("create empty b content = %q, want empty", got)
	}

	if code := runCLI(t, "copy", a, cp); code != ExitSuccess {
		t.Fatalf("copy: exit %d", code)
	}
	if got := readFile(t, cp); got != "hello" {
		t.Fatalf("copy content = %q, want %q", got, "hello")
	}

	if code := runCLI(t, "combine", "-sep=\n", a, cp, combined); code != ExitSuccess {
		t.Fatalf("combine: exit %d", code)
	}
	if got := readFile(t, combined); got != "hello\nhello" {
		t.Fatalf("combine content = %q, want %q", got, "hello\nhello")
	}

	if code := runCLI(t, "delete", a); code != ExitSuccess {
		t.Fatalf("delete: exit %d", code)
	}
	if _, err := os.Stat(a); !os.IsNotExist(err) {
		t.Fatalf("delete: file %s still exists", a)
	}
}

// TestIntegration_UsageErrors verifies the CLI returns the usage exit code for
// missing or unknown subcommands and missing required arguments.
func TestIntegration_UsageErrors(t *testing.T) {
	if code := runCLI(t); code != ExitUsage {
		t.Fatalf("no args: exit %d, want %d", code, ExitUsage)
	}
	if code := runCLI(t, "bogus"); code != ExitUsage {
		t.Fatalf("unknown subcommand: exit %d, want %d", code, ExitUsage)
	}
	if code := runCLI(t, "create"); code != ExitUsage {
		t.Fatalf("create without path: exit %d, want %d", code, ExitUsage)
	}
}

// TestIntegration_ForceOverwrite verifies the -force safety semantics.
func TestIntegration_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")

	if code := runCLI(t, "create", "-text=one", f); code != ExitSuccess {
		t.Fatalf("create: exit %d", code)
	}
	if code := runCLI(t, "create", "-text=two", f); code != ExitFailure {
		t.Fatalf("overwrite without -force: exit %d, want %d", code, ExitFailure)
	}
	if code := runCLI(t, "create", "-force", "-text=two", f); code != ExitSuccess {
		t.Fatalf("overwrite with -force: exit %d", code)
	}
	if got := readFile(t, f); got != "two" {
		t.Fatalf("content = %q, want %q", got, "two")
	}
}

// TestIntegration_VersionCommand verifies the version subcommand succeeds.
func TestIntegration_VersionCommand(t *testing.T) {
	if code := runCLI(t, "version"); code != ExitSuccess {
		t.Fatalf("version: exit %d", code)
	}
}
