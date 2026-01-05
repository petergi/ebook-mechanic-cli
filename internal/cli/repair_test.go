package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
)

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	content := []byte("hello world")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}

	if string(got) != string(content) {
		t.Errorf("expected %s, got %s", content, got)
	}
}

func TestRepairToOutput(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.epub")
	dst := filepath.Join(tmpDir, "dst.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	op := operations.NewRepairOperation(context.Background())
	result, err := repairToOutput(op, src, dst)

	if err != nil {
		t.Fatalf("repairToOutput failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result")
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("Destination file not created")
	}
}

func TestRepairInPlace_NoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "book.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	op := operations.NewRepairOperation(context.Background())
	result, err := repairInPlace(op, src, false, "")

	if err != nil {
		t.Fatalf("repairInPlace failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result")
	}
}

func TestRunRepair_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "book.epub")
	dst := filepath.Join(tmpDir, "fixed.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	ctx := context.Background()
	flags := &repairFlags{
		output: dst,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runRepair(ctx, src, flags, rootFlags)
}

func TestRunRepair_InPlace(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "book.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	ctx := context.Background()
	flags := &repairFlags{
		inPlace: true,
		backup:  true,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runRepair(ctx, src, flags, rootFlags)
}

func TestRepairInPlace_InvalidFile(t *testing.T) {
	op := operations.NewRepairOperation(context.Background())
	_, err := repairInPlace(op, "/non/existent", true, "")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestRepairToOutput_ExistingDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "newdir")
	src := filepath.Join(tmpDir, "src.epub")
		dst := filepath.Join(subDir, "dst.epub")
		
		if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		
		op := operations.NewRepairOperation(context.Background())
	
	_, _ = repairToOutput(op, src, dst)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("Subdirectory should have been created")
	}
}
