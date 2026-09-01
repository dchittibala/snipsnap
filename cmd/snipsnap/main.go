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
		_, _ = fmt.Fprintln(stderr, "Error: missing subcommand")
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
		_, _ = fmt.Fprintf(stderr, "Error: unknown subcommand %q\n", cmd)
		printUsage(stderr)
		return ExitUsage
	}

	// Pre-process flags and positionals to support flags positioned AFTER arguments
	flagArgs, positionals := parseInterspersedArgs(fs, args[1:])
	if err := fs.Parse(flagArgs); err != nil {
		return ExitUsage
	}

	var err error
	switch cmd {
	case "create":
		if len(positionals) < 1 {
			_, _ = fmt.Fprintln(stderr, "Error: create requires a target filename")
			_, _ = fmt.Fprintln(stderr, "Usage: snipsnap create [-text=...] [-force] <file>")
			return ExitUsage
		}
		err = ops.Create(ctx, positionals[0], text, force)

	case "copy":
		if len(positionals) < 2 {
			_, _ = fmt.Fprintln(stderr, "Error: copy requires source and destination paths")
			_, _ = fmt.Fprintln(stderr, "Usage: snipsnap copy [-force] <src> <dst>")
			return ExitUsage
		}
		err = ops.Copy(ctx, positionals[0], positionals[1], force)

	case "combine":
		if len(positionals) < 3 {
			_, _ = fmt.Fprintln(stderr, "Error: combine requires file1, file2, and destination paths")
			_, _ = fmt.Fprintln(stderr, "Usage: snipsnap combine [-sep=...] [-force] <file1> <file2> <dst>")
			return ExitUsage
		}
		err = ops.Combine(ctx, positionals[0], positionals[1], positionals[2], sep, force)

	case "delete":
		if len(positionals) < 1 {
			_, _ = fmt.Fprintln(stderr, "Error: delete requires a file path")
			_, _ = fmt.Fprintln(stderr, "Usage: snipsnap delete <file>")
			return ExitUsage
		}
		err = ops.Delete(ctx, positionals[0])
	}

	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return ExitFailure
	}

	return ExitSuccess
}

// parseInterspersedArgs separates flags (-force, -text=foo) from positional parameters
// allowing flags to be supplied anywhere in the CLI invocation.
func parseInterspersedArgs(fs *flag.FlagSet, args []string) (flags []string, positionals []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}

		if len(arg) > 1 && arg[0] == '-' {
			flags = append(flags, arg)
			// If flag requires a value provided as next arg (e.g. -text "hello")
			name := arg[1:]
			if len(name) > 0 && name[0] == '-' {
				name = name[1:]
			}
			if f := fs.Lookup(name); f != nil {
				// If it's not a boolean flag and doesn't use '=', grab the next argument
				if getter, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok || !getter.IsBoolFlag() {
					if !containsEquals(arg) && i+1 < len(args) {
						i++
						flags = append(flags, args[i])
					}
				}
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return flags, positionals
}

func containsEquals(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return true
		}
	}
	return false
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "\nUsage: snipsnap <command> [flags] [arguments]")
	_, _ = fmt.Fprintln(w, "\nAvailable Commands:")
	_, _ = fmt.Fprintln(w, "  create   Create a new file (empty or with -text)")
	_, _ = fmt.Fprintln(w, "  copy     Copy a file to a new location")
	_, _ = fmt.Fprintln(w, "  combine  Combine two files sequentially into a third file")
	_, _ = fmt.Fprintln(w, "  delete   Delete an existing file")
}
