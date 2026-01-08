package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunBatchValidate_WithFile(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {

		t.Fatal(err)

	}

	ctx := context.Background()

	flags := &batchFlags{
		recursive: true,
		jobs:      1,
		progress:  "none",
		maxDepth:  -1,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runBatchValidate(ctx, tmpDir, flags, rootFlags)
}

func TestRunBatchRepair_WithFile(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {

		t.Fatal(err)

	}

	ctx := context.Background()

	flags := &batchFlags{
		recursive: true,
		jobs:      1,
		progress:  "none",
		maxDepth:  -1,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runBatchRepair(ctx, tmpDir, flags, rootFlags)
}

func TestRunBatchValidate_ProgressModes(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {

		t.Fatal(err)

	}

	modes := []string{"auto", "simple", "none"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			flags := &batchFlags{
				progress: mode,
				jobs:     1,
				maxDepth: -1,
			}
			rootFlags := &RootFlags{Color: false}
			_ = runBatchValidate(context.Background(), tmpDir, flags, rootFlags)
		})
	}
}

func TestRunBatchRepair_WithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	flags := &batchFlags{
		recursive: true,
		jobs:      1,
		progress:  "none",
		maxDepth:  -1,
		backupDir: filepath.Join(tmpDir, "backups"),
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runBatchRepair(ctx, tmpDir, flags, rootFlags)
}

func TestIsTerminal(t *testing.T) {
	_ = isTerminal()
}

func TestRunBatchRepair_DefaultJobs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	flags := &batchFlags{
		jobs:     0,
		maxDepth: -1,
	}
	rootFlags := &RootFlags{Color: false}
	_ = runBatchRepair(ctx, tmpDir, flags, rootFlags)
}

func TestRunBatchValidate_SummaryOnly(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {

		t.Fatal(err)

	}

	ctx := context.Background()

	flags := &batchFlags{
		summaryOnly: true,
		jobs:        1,
		maxDepth:    -1,
	}
	rootFlags := &RootFlags{Color: false, Format: "text"}

	_ = runBatchValidate(ctx, tmpDir, flags, rootFlags)
}

func TestRunBatchValidate_WithIgnore(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("t"), 0644); err != nil {

		t.Fatal(err)

	}

	ctx := context.Background()

	flags := &batchFlags{
		ignore:   []string{"*.epub"},
		jobs:     1,
		maxDepth: -1,
	}
	rootFlags := &RootFlags{}

	err := runBatchValidate(ctx, tmpDir, flags, rootFlags)
	if err == nil {
		t.Error("Expected error because all files were ignored")
	}
}

func TestRunBatchRepair_InvalidOpts(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Test invalid backup dir
	flags := &batchFlags{
		backupDir: "/invalid/path/that/cannot/exist",
	}
	rootFlags := &RootFlags{}

	err := runBatchRepair(ctx, tmpDir, flags, rootFlags)
	if err == nil {
		t.Error("Expected error for invalid backup dir")
	}
}

func TestRunBatchValidate_InvalidReport(t *testing.T) {

	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("t"), 0644); err != nil {

		t.Fatal(err)

	}

	ctx := context.Background()

	flags := &batchFlags{jobs: 1, maxDepth: -1}
	rootFlags := &RootFlags{Output: "/invalid/path"}

	err := runBatchValidate(ctx, tmpDir, flags, rootFlags)
	if err == nil {
		t.Error("Expected error for invalid output path")
	}
}

func TestBatchRepair_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := context.Background()
	flags := &batchFlags{
		recursive: true,
		jobs:      1,
	}
	rootFlags := &RootFlags{Color: false}

	err := runBatchRepair(ctx, tmpDir, flags, rootFlags)
	if err == nil {
		t.Error("expected error when no files found in directory")
	}
}

func TestRunBatchRepair_InvalidBackupDir(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("t"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	flags := &batchFlags{
		backupDir: "/proc/invalid/path",
		maxDepth:  -1,
	}
	rootFlags := &RootFlags{}
	err := runBatchRepair(ctx, tmpDir, flags, rootFlags)
	if err == nil {
		t.Error("Expected error for invalid backup dir")
	}
}

func TestRunBatchRepair_InvalidDir(t *testing.T) {
	ctx := context.Background()
	flags := &batchFlags{jobs: 1}
	rootFlags := &RootFlags{}
	err := runBatchRepair(ctx, "/non/existent/dir/12345", flags, rootFlags)
	if err == nil {
		t.Error("Expected error for non-existent dir")
	}
}

func TestRunBatchValidate_DefaultJobs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	flags := &batchFlags{jobs: 0, maxDepth: -1}
	rootFlags := &RootFlags{Color: false}
	_ = runBatchValidate(ctx, tmpDir, flags, rootFlags)
}
