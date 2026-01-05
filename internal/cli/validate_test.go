package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFromFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")
	if err := os.WriteFile(tmpFile, []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validateFromFile(ctx, tmpFile)

	if err != nil {
		t.Errorf("validateFromFile returned error: %v", err)
	}

	if report == nil {
		t.Fatal("expected report to be non-nil")
	}

	if report.FilePath != tmpFile {
		t.Errorf("expected report filepath %s, got %s", tmpFile, report.FilePath)
	}
}

func TestRunValidate_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "book.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	ctx := context.Background()
	flags := &validateFlags{}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runValidate(ctx, src, flags, rootFlags)
}

func TestRunValidate_Stdin(t *testing.T) {
	// Mock stdin
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

		go func() {
			_, _ = w.Write([]byte("dummy content"))
			_ = w.Close()
		}()

	ctx := context.Background()
	flags := &validateFlags{fileType: "epub"}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runValidate(ctx, "-", flags, rootFlags)
}

func TestRunValidate_Stdin_NoType(t *testing.T) {
	ctx := context.Background()
	flags := &validateFlags{fileType: ""}
	rootFlags := &RootFlags{}

	err := runValidate(ctx, "-", flags, rootFlags)
	if err == nil {
		t.Error("Expected error for missing type with stdin")
	}
}

func TestRunValidate_WriteError(t *testing.T) {

	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "book.epub")

	if err := os.WriteFile(src, []byte("t"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	flags := &validateFlags{}

	rootFlags := &RootFlags{Output: "/invalid/path"}

	err := runValidate(ctx, src, flags, rootFlags)

	if err == nil {

		t.Error("Expected error for invalid output path")

	}

}
