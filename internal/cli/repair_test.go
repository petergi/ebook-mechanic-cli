package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
)

func TestRepairInPlace_BackupOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "book.epub")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	op := operations.NewRepairOperation(context.Background())
	result, _, err := repairInPlace(op, src, operations.RepairSaveModeBackupOriginal, "")

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

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}

	ctx := context.Background()
	flags := &repairFlags{
		noBackup:     true,
		skipValidate: true,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runRepair(ctx, src, flags, rootFlags)
}

func TestRepairInPlace_InvalidFile(t *testing.T) {
	op := operations.NewRepairOperation(context.Background())
	_, _, err := repairInPlace(op, "/non/existent", operations.RepairSaveModeBackupOriginal, "")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestRepairToOutput_ExistingDir(t *testing.T) {
	t.Skip("repair output mode is no longer supported")
}
