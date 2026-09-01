package fileops

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Sentinel errors exported for caller inspection
var (
	ErrDestinationExists = errors.New("destination exists (use -force to overwrite)")
	ErrIsDirectory       = errors.New("target is a directory, not a file")
)

const ioBufferSize = 32 * 1024

// FileOps defines exported file operations interface.
type FileOps interface {
	Create(ctx context.Context, path, text string, force bool) error
	Copy(ctx context.Context, src, dst string, force bool) error
	Combine(ctx context.Context, a, b, dst, sep string, force bool) error
	Delete(ctx context.Context, path string) error
}

type RealFileOps struct{}

func New() FileOps {
	return RealFileOps{}
}

func (RealFileOps) Create(ctx context.Context, path, text string, force bool) error {
	cleanPath := filepath.Clean(path)
	return AtomicWrite(cleanPath, force, func(w io.Writer) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := io.WriteString(w, text)
			return err
		}
	})
}

func (RealFileOps) Copy(ctx context.Context, src, dst string, force bool) error {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	in, err := os.Open(cleanSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", cleanSrc, err)
	}
	defer in.Close()

	return AtomicWrite(cleanDst, force, func(w io.Writer) error {
		return CopyWithContext(ctx, w, in)
	})
}

func (RealFileOps) Combine(ctx context.Context, a, b, dst, sep string, force bool) error {
	cleanA := filepath.Clean(a)
	cleanB := filepath.Clean(b)
	cleanDst := filepath.Clean(dst)

	return AtomicWrite(cleanDst, force, func(w io.Writer) error {
		inA, err := os.Open(cleanA)
		if err != nil {
			return fmt.Errorf("failed to open input file A %q: %w", cleanA, err)
		}
		err = CopyWithContext(ctx, w, inA)
		inA.Close()
		if err != nil {
			return err
		}

		if sep != "" {
			if _, err := io.WriteString(w, sep); err != nil {
				return fmt.Errorf("failed writing separator: %w", err)
			}
		}

		inB, err := os.Open(cleanB)
		if err != nil {
			return fmt.Errorf("failed to open input file B %q: %w", cleanB, err)
		}
		err = CopyWithContext(ctx, w, inB)
		inB.Close()
		return err
	})
}

func (RealFileOps) Delete(_ context.Context, path string) error {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to stat path %q: %w", cleanPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%w: %q", ErrIsDirectory, cleanPath)
	}
	if err := os.Remove(cleanPath); err != nil {
		return fmt.Errorf("failed to delete file %q: %w", cleanPath, err)
	}
	return nil
}

func CopyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	buf := make([]byte, ioBufferSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			nr, rErr := src.Read(buf)
			if nr > 0 {
				nw, wErr := dst.Write(buf[:nr])
				if wErr != nil {
					return fmt.Errorf("write error during copy: %w", wErr)
				}
				if nr != nw {
					return io.ErrShortWrite
				}
			}
			if rErr != nil {
				if errors.Is(rErr, io.EOF) {
					return nil
				}
				return fmt.Errorf("read error during copy: %w", rErr)
			}
		}
	}
}

func AtomicWrite(path string, force bool, writeFn func(io.Writer) error) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %q", ErrDestinationExists, path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect destination path %q: %w", path, err)
		}
	}

	targetDir := filepath.Dir(path)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory tree %q: %w", targetDir, err)
	}

	tmp, err := os.CreateTemp(targetDir, ".snipsnap-tmp-*")
	if err != nil {
		return fmt.Errorf("failed creating temporary file: %w", err)
	}
	tmpName := tmp.Name()

	defer func() {
		_ = os.Remove(tmpName)
	}()

	bufWriter := bufio.NewWriterSize(tmp, ioBufferSize)
	if err := writeFn(bufWriter); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := bufWriter.Flush(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to flush write buffer: %w", err)
	}

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed setting permissions: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("atomic rename failed from %q to %q: %w", tmpName, path, err)
	}

	return nil
}
