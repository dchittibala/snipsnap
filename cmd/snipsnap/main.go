package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/dchittibala/snipsnap/pkg/fileops"
)

const (
	ExitSuccess = 0
	ExitFailure = 1
	ExitUsage   = 64
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ops := fileops.New()
	code := run(ctx, os.Args[1:], ops, os.Stderr)
	os.Exit(code)
}

func run(ctx context.Context, args []string, ops fileops.FileOps, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Error: missing subcommand")
		printUsage(stderr)
		return ExitUsage
	}

	cmd := args[0]
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var text, sep string
	var force bool

	switch cmd {
	case "create":
		fs.StringVar(&text, "text", "", "optional text content to populate file")
		fs.BoolVar(&force, "force", false, "overwrite destination file if it exists")
	case "copy":
		fs.BoolVar(&force, "force", false, "overwrite destination file if it exists")
	case "combine":
		fs.StringVar(&sep, "sep", "", "optional string separator between files")
		fs.BoolVar(&force, "force", false, "overwrite destination file if it exists")
	case "delete":
		// no flags required
	default:
		fmt.Fprintf(stderr, "Error: unknown subcommand %q\n", cmd)
		printUsage(stderr)
		return ExitUsage
	}

	if err := fs.Parse(args[1:]); err != nil {
		return ExitUsage
	}

	positionals := fs.Args()

	var err error
	switch cmd {
	case "create":
		if len(positionals) < 1 {
			fmt.Fprintln(stderr, "Error: create requires a target filename")
			fmt.Fprintln(stderr, "Usage: snipsnap create [-text=...] [-force] <file>")
			return ExitUsage
		}
		err = ops.Create(ctx, positionals[0], text, force)

	case "copy":
		if len(positionals) < 2 {
			fmt.Fprintln(stderr, "Error: copy requires source and destination paths")
			fmt.Fprintln(stderr, "Usage: snipsnap copy [-force] <src> <dst>")
			return ExitUsage
		}
		err = ops.Copy(ctx, positionals[0], positionals[1], force)

	case "combine":
		if len(positionals) < 3 {
			fmt.Fprintln(stderr, "Error: combine requires file1, file2, and destination paths")
			fmt.Fprintln(stderr, "Usage: snipsnap combine [-sep=...] [-force] <file1> <file2> <dst>")
			return ExitUsage
		}
		err = ops.Combine(ctx, positionals[0], positionals[1], positionals[2], sep, force)

	case "delete":
		if len(positionals) < 1 {
			fmt.Fprintln(stderr, "Error: delete requires a file path")
			fmt.Fprintln(stderr, "Usage: snipsnap delete <file>")
			return ExitUsage
		}
		err = ops.Delete(ctx, positionals[0])
	}

	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ExitFailure
	}

	return ExitSuccess
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "\nUsage: snipsnap <command> [flags] [arguments]")
	fmt.Fprintln(w, "\nAvailable Commands:")
	fmt.Fprintln(w, "  create   Create a new file (empty or with -text)")
	fmt.Fprintln(w, "  copy     Copy a file to a new location")
	fmt.Fprintln(w, "  combine  Combine two files sequentially into a third file")
	fmt.Fprintln(w, "  delete   Delete an existing file")
}
