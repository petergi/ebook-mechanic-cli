package operations

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

func TestDefaultBatchConfig(t *testing.T) {
	config := DefaultBatchConfig()

	if config.NumWorkers != runtime.NumCPU() {
		t.Errorf("Expected NumWorkers to be %d, got %d", runtime.NumCPU(), config.NumWorkers)
	}

	if config.QueueSize != 100 {
		t.Errorf("Expected QueueSize to be 100, got %d", config.QueueSize)
	}

	if config.ProgressRate != 100*time.Millisecond {
		t.Errorf("Expected ProgressRate to be 100ms, got %v", config.ProgressRate)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", config.Timeout)
	}
}

func TestNewBatchProcessor(t *testing.T) {
	config := DefaultBatchConfig()
	bp := NewBatchProcessor(context.Background(), config)

	if bp == nil {
		t.Fatal("Expected non-nil batch processor")
	}

	if bp.config.NumWorkers != config.NumWorkers {
		t.Errorf("Expected NumWorkers to be %d, got %d", config.NumWorkers, bp.config.NumWorkers)
	}

	if bp.taskQueue == nil {
		t.Error("Expected taskQueue to be initialized")
	}

	if bp.resultQueue == nil {
		t.Error("Expected resultQueue to be initialized")
	}

	if bp.progressCh == nil {
		t.Error("Expected progressCh to be initialized")
	}
}

func TestBatchProcessor_Cancel(t *testing.T) {
	config := DefaultBatchConfig()
	bp := NewBatchProcessor(context.Background(), config)

	// Cancel the processor
	bp.Cancel()

	// Context should be done
	select {
	case <-bp.ctx.Done():
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context to be done after cancel")
	}
}

func TestBatchProcessor_ProgressChannel(t *testing.T) {
	config := DefaultBatchConfig()
	bp := NewBatchProcessor(context.Background(), config)

	ch := bp.ProgressChannel()

	if ch == nil {
		t.Error("Expected non-nil progress channel")
	}
}

func TestFindFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested structure
	// root/
	//   book1.epub
	//   book2.pdf
	//   other.txt
	//   subdir/
	//     book3.epub
	//     nested/
	//       book4.pdf

	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "book1.epub"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "book2.pdf"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "book3.epub"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "nested", "book4.pdf"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		opts     FindFilesOptions
		expected int
	}{
		{"Recursive all", FindFilesOptions{Recursive: true, MaxDepth: -1}, 4},
		{"Non-recursive", FindFilesOptions{Recursive: false, MaxDepth: -1}, 2},
		{"Max depth 1", FindFilesOptions{Recursive: true, MaxDepth: 1}, 2},
		{"Extensions filter", FindFilesOptions{Recursive: true, MaxDepth: -1, Extensions: []string{".epub"}}, 2},
		{"Ignore pattern", FindFilesOptions{Recursive: true, MaxDepth: -1, Ignore: []string{"subdir"}}, 2},
		{"Multiple extensions", FindFilesOptions{Recursive: true, MaxDepth: -1, Extensions: []string{"epub", "pdf"}}, 4},
		{"Ignore file pattern", FindFilesOptions{Recursive: true, MaxDepth: -1, Ignore: []string{"book1.epub"}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := FindFiles(tmpDir, tt.opts)
			if err != nil {
				t.Fatalf("FindFiles failed: %v", err)
			}
			if len(files) != tt.expected {
				t.Errorf("Expected %d files, got %d: %v", tt.expected, len(files), files)
			}
		})
	}

	t.Run("Extensions without dots", func(t *testing.T) {
		files, _ := FindFiles(tmpDir, FindFilesOptions{Recursive: true, MaxDepth: -1, Extensions: []string{"epub"}})
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
	})

	t.Run("Ignore directory", func(t *testing.T) {
		files, _ := FindFiles(tmpDir, FindFilesOptions{Recursive: true, MaxDepth: -1, Ignore: []string{"subdir"}})
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
	})

	t.Run("No extensions filter", func(t *testing.T) {
		files, _ := FindFiles(tmpDir, FindFilesOptions{Recursive: true, MaxDepth: -1, Extensions: []string{}})
		if len(files) != 4 {
			t.Errorf("Expected 4 files, got %d", len(files))
		}
	})
}

func TestBatchProcessor_ProgressAndCancel(t *testing.T) {
	config := DefaultBatchConfig()
	config.NumWorkers = 1
	bp := NewBatchProcessor(context.Background(), config)

	if bp.ProgressChannel() == nil {
		t.Error("Expected ProgressChannel")
	}

	bp.Cancel()
	// bp.ctx should be cancelled
}

func TestAggregateResults_Detailed(t *testing.T) {
	results := []Result{
		{FilePath: "valid.epub", Report: &ebmlib.ValidationReport{IsValid: true}},
		{FilePath: "invalid.epub", Report: &ebmlib.ValidationReport{IsValid: false}},
		{FilePath: "system_err.epub", Error: errors.New("io error")},
		{FilePath: "repair_ok.epub", Repair: &ebmlib.RepairResult{Success: true, ActionsApplied: []ebmlib.RepairAction{{}}}},
		{FilePath: "repair_noop.epub", Repair: &ebmlib.RepairResult{Success: true}},
		{FilePath: "repair_fail.epub", Repair: &ebmlib.RepairResult{Success: false, ActionsApplied: []ebmlib.RepairAction{{}}}},
		{FilePath: "repair_fail_no_actions.epub", Repair: &ebmlib.RepairResult{Success: false}},
	}

	br := AggregateResults(results, time.Second, OperationRepair)

	if len(br.Valid) != 3 { // valid.epub, repair_ok.epub, repair_noop.epub
		t.Errorf("Expected 3 valid, got %d", len(br.Valid))
	}
	if len(br.Invalid) != 3 { // invalid.epub, repair_fail.epub, repair_fail_no_actions.epub
		t.Errorf("Expected 3 invalid, got %d", len(br.Invalid))
	}
	if len(br.Errored) != 1 { // system_err.epub
		t.Errorf("Expected 1 errored, got %d", len(br.Errored))
	}
	if br.RepairsAttempted != 4 { // repair_ok, repair_fail, repair_fail_no_actions, system_err
		t.Errorf("Expected 4 repairs attempted, got %d", br.RepairsAttempted)
	}
	if br.RepairsSucceeded != 1 { // repair_ok
		t.Errorf("Expected 1 repair succeeded, got %d", br.RepairsSucceeded)
	}
	if br.RepairsNoOp != 1 { // repair_noop
		t.Errorf("Expected 1 no-op repair, got %d", br.RepairsNoOp)
	}
}

func TestAggregateResults_AllSuccessful(t *testing.T) {
	results := []Result{
		{FilePath: "file1.epub", Error: nil},
		{FilePath: "file2.epub", Error: nil},
	}

	batchResult := AggregateResults(results, 0, OperationValidate)

	if len(batchResult.Failed) != 0 {
		t.Errorf("Expected 0 failed results, got %d", len(batchResult.Failed))
	}

	if len(batchResult.Successful) != 2 {
		t.Errorf("Expected 2 successful results, got %d", len(batchResult.Successful))
	}
}

func TestAggregateResults_AllFailed(t *testing.T) {
	results := []Result{
		{FilePath: "file1.epub", Error: context.DeadlineExceeded},
		{FilePath: "file2.epub", Error: context.DeadlineExceeded},
	}

	batchResult := AggregateResults(results, 0, OperationValidate)

	if len(batchResult.Successful) != 0 {
		t.Errorf("Expected 0 successful results, got %d", len(batchResult.Successful))
	}

	if len(batchResult.Failed) != 2 {
		t.Errorf("Expected 2 failed results, got %d", len(batchResult.Failed))
	}
}

func TestAggregateResults_Empty(t *testing.T) {
	results := []Result{}

	batchResult := AggregateResults(results, 0, OperationValidate)

	if batchResult.Total != 0 {
		t.Errorf("Expected Total to be 0, got %d", batchResult.Total)
	}

	if len(batchResult.Successful) != 0 {
		t.Errorf("Expected 0 successful results, got %d", len(batchResult.Successful))
	}

	if len(batchResult.Failed) != 0 {
		t.Errorf("Expected 0 failed results, got %d", len(batchResult.Failed))
	}
}

func TestOperationType(t *testing.T) {
	// Test that operation types exist and can be compared
	if OperationValidate != "validate" {
		t.Errorf("Expected OperationValidate to be 'validate', got '%s'", OperationValidate)
	}

	if OperationRepair != "repair" {
		t.Errorf("Expected OperationRepair to be 'repair', got '%s'", OperationRepair)
	}
}

func TestTask(t *testing.T) {
	// Test that Task type can be created
	task := Task{
		FilePath:  "test.epub",
		Operation: OperationValidate,
	}

	if task.FilePath != "test.epub" {
		t.Errorf("Expected FilePath 'test.epub', got '%s'", task.FilePath)
	}

	if task.Operation != OperationValidate {
		t.Errorf("Expected Operation 'validate', got '%s'", task.Operation)
	}
}

func TestResult(t *testing.T) {
	// Test that Result type can be created
	result := Result{
		FilePath: "test.epub",
		Report:   nil,
		Repair:   nil,
		Error:    nil,
	}

	if result.FilePath != "test.epub" {
		t.Errorf("Expected FilePath 'test.epub', got '%s'", result.FilePath)
	}
}

func TestBatchProcessor_ProcessTask_UnknownOperation(t *testing.T) {
	config := DefaultBatchConfig()
	config.Timeout = 50 * time.Millisecond
	bp := NewBatchProcessor(context.Background(), config)

	result := bp.processTask(Task{FilePath: "file.txt", Operation: "unknown"})

	if result.Error != nil {
		t.Errorf("Expected no error for unknown operation, got %v", result.Error)
	}
}

func TestBatchProcessor_Execute_ValidateUnsupported(t *testing.T) {
	config := DefaultBatchConfig()
	config.NumWorkers = 1
	config.QueueSize = 2
	config.ProgressRate = 10 * time.Millisecond
	config.Timeout = 50 * time.Millisecond

	bp := NewBatchProcessor(context.Background(), config)
	results := bp.Execute([]string{"file1.txt", "file2.txt"}, OperationValidate)
	bp.Cancel()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error == nil {
			t.Errorf("Expected error for file %s", result.FilePath)
		}
		if result.Report != nil {
			t.Errorf("Expected no report for unsupported file %s", result.FilePath)
		}
	}
}

func TestBatchProcessor_Execute_RepairUnsupported(t *testing.T) {
	config := DefaultBatchConfig()
	config.NumWorkers = 1
	config.QueueSize = 1
	config.ProgressRate = 10 * time.Millisecond
	config.Timeout = 50 * time.Millisecond

	bp := NewBatchProcessor(context.Background(), config)
	results := bp.Execute([]string{"file1.txt"}, OperationRepair)
	bp.Cancel()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected error for unsupported repair file")
	}

	if results[0].Repair != nil {
		t.Error("Expected no repair result for unsupported file")
	}
}

func TestBatchProcessor_Execute_Canceled(t *testing.T) {
	config := DefaultBatchConfig()
	config.NumWorkers = 1
	config.QueueSize = 2
	config.Timeout = 50 * time.Millisecond

	bp := NewBatchProcessor(context.Background(), config)
	bp.Cancel()

	results := bp.Execute([]string{"file1.txt", "file2.txt"}, OperationValidate)

	if len(results) != 0 {
		t.Fatalf("Expected no results after cancellation, got %d", len(results))
	}
}

func TestBatchProcessor_ReportProgress(t *testing.T) {
	config := DefaultBatchConfig()
	config.ProgressRate = 5 * time.Millisecond

	bp := NewBatchProcessor(context.Background(), config)
	bp.total = 3
	bp.completed.Store(1)
	bp.currentFile.Store("file1.txt")

	go bp.reportProgress()
	defer bp.Cancel()

	select {
	case update := <-bp.ProgressChannel():
		if update.Completed != 1 {
			t.Errorf("Expected completed to be 1, got %d", update.Completed)
		}
		if update.Total != 3 {
			t.Errorf("Expected total to be 3, got %d", update.Total)
		}
		if update.Current != "file1.txt" {
			t.Errorf("Expected current to be file1.txt, got %s", update.Current)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected progress update")
	}
}

func TestProgressUpdate(t *testing.T) {
	// Test that ProgressUpdate type can be created
	update := ProgressUpdate{
		Completed: 5,
		Total:     10,
		Current:   "file5.epub",
	}

	if update.Completed != 5 {
		t.Errorf("Expected Completed to be 5, got %d", update.Completed)
	}

	if update.Total != 10 {
		t.Errorf("Expected Total to be 10, got %d", update.Total)
	}

	if update.Current != "file5.epub" {
		t.Errorf("Expected Current to be 'file5.epub', got '%s'", update.Current)
	}
}
