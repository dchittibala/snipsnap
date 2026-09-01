package fileops_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dchittibala/snipsnap/pkg/fileops"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	ops := fileops.New()
	ctx := context.Background()

	t.Run("Create empty file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "empty.txt")

		err := ops.Create(ctx, target, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("failed to stat created file: %v", err)
		}
		if info.Size() != 0 {
			t.Errorf("expected size 0, got %d", info.Size())
		}
	})

	t.Run("Create file with text content", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "hello.txt")
		expected := "Hello, Snipsnap!"

		err := ops.Create(ctx, target, expected, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}
		if string(data) != expected {
			t.Errorf("expected %q, got %q", expected, string(data))
		}
	})

	t.Run("Auto-create directory tree", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "nested", "deeply", "file.txt")

		err := ops.Create(ctx, target, "nested data", false)
		if err != nil {
			t.Fatalf("expected directory auto-creation to succeed, got: %v", err)
		}

		if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Fatal("nested file was not created")
		}
	})

	t.Run("Refuse overwrite without force flag", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "existing.txt")
		_ = ops.Create(ctx, target, "initial", false)

		err := ops.Create(ctx, target, "overwrite attempt", false)
		if !errors.Is(err, fileops.ErrDestinationExists) {
			t.Errorf("expected ErrDestinationExists, got: %v", err)
		}
	})

	t.Run("Allow overwrite with force flag", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "existing.txt")
		_ = ops.Create(ctx, target, "initial", false)

		err := ops.Create(ctx, target, "new content", true)
		if err != nil {
			t.Fatalf("unexpected error with force=true: %v", err)
		}

		data, _ := os.ReadFile(target)
		if string(data) != "new content" {
			t.Errorf("expected 'new content', got %q", string(data))
		}
	})
}

func TestCopy(t *testing.T) {
	t.Parallel()
	ops := fileops.New()
	ctx := context.Background()

	t.Run("Copy file successfully", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		content := "Copy this stream payload"

		_ = ops.Create(ctx, src, content, false)

		err := ops.Copy(ctx, src, dst, false)
		if err != nil {
			t.Fatalf("unexpected copy error: %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != content {
			t.Errorf("expected copied content %q, got %q", content, string(data))
		}
	})

	t.Run("Fail copying non-existent source", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		src := filepath.Join(dir, "missing.txt")
		dst := filepath.Join(dir, "dst.txt")

		err := ops.Copy(ctx, src, dst, false)
		if err == nil {
			t.Fatal("expected error copying non-existent file, got nil")
		}
	})
}

func TestCombine(t *testing.T) {
	t.Parallel()
	ops := fileops.New()
	ctx := context.Background()

	t.Run("Combine two files with separator", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.txt")
		f2 := filepath.Join(dir, "b.txt")
		dst := filepath.Join(dir, "combined.txt")

		_ = ops.Create(ctx, f1, "PartA", false)
		_ = ops.Create(ctx, f2, "PartB", false)

		err := ops.Combine(ctx, f1, f2, dst, " --- ", false)
		if err != nil {
			t.Fatalf("unexpected combine error: %v", err)
		}

		data, _ := os.ReadFile(dst)
		expected := "PartA --- PartB"
		if string(data) != expected {
			t.Errorf("expected %q, got %q", expected, string(data))
		}
	})

	t.Run("Combine into one of the source files (In-place overwrite with force)", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.txt")
		f2 := filepath.Join(dir, "b.txt")

		_ = ops.Create(ctx, f1, "Foo", false)
		_ = ops.Create(ctx, f2, "Bar", false)

		// Combining f1 and f2 directly INTO f1 (requires force=true)
		err := ops.Combine(ctx, f1, f2, f1, "-", true)
		if err != nil {
			t.Fatalf("unexpected in-place combine error: %v", err)
		}

		data, _ := os.ReadFile(f1)
		if string(data) != "Foo-Bar" {
			t.Errorf("expected 'Foo-Bar', got %q", string(data))
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ops := fileops.New()
	ctx := context.Background()

	t.Run("Delete existing file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "delete_me.txt")
		_ = ops.Create(ctx, target, "temp", false)

		err := ops.Delete(ctx, target)
		if err != nil {
			t.Fatalf("unexpected delete error: %v", err)
		}

		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Error("file still exists after delete call")
		}
	})

	t.Run("Refuse to delete directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		subDir := filepath.Join(dir, "myFolder")
		_ = os.Mkdir(subDir, 0o755)

		err := ops.Delete(ctx, subDir)
		if !errors.Is(err, fileops.ErrIsDirectory) {
			t.Errorf("expected ErrIsDirectory, got %v", err)
		}
	})
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()
	ops := fileops.New()

	t.Run("Abort operation when context is canceled", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		target := filepath.Join(dir, "cancelled.txt")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately before triggering operation

		err := ops.Create(ctx, target, "should not write", false)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got: %v", err)
		}

		// Ensure no temp file or target file was left behind
		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Errorf("expected empty directory on cancellation rollback, found %d entries", len(entries))
		}
	})
}
