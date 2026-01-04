package operations

import (
	"context"
	"runtime"
	"testing"
	"time"
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

func TestAggregateResults(t *testing.T) {
	results := []Result{
		{FilePath: "file1.epub", Error: nil},
		{FilePath: "file2.epub", Error: nil},
		{FilePath: "file3.epub", Error: context.DeadlineExceeded},
		{FilePath: "file4.epub", Error: nil},
	}

	duration := 5 * time.Second
	batchResult := AggregateResults(results, duration)

	if batchResult.Total != 4 {
		t.Errorf("Expected Total to be 4, got %d", batchResult.Total)
	}

	if len(batchResult.Successful) != 3 {
		t.Errorf("Expected 3 successful results, got %d", len(batchResult.Successful))
	}

	if len(batchResult.Failed) != 1 {
		t.Errorf("Expected 1 failed result, got %d", len(batchResult.Failed))
	}

	if batchResult.Duration != duration {
		t.Errorf("Expected duration to be %v, got %v", duration, batchResult.Duration)
	}
}

func TestAggregateResults_AllSuccessful(t *testing.T) {
	results := []Result{
		{FilePath: "file1.epub", Error: nil},
		{FilePath: "file2.epub", Error: nil},
	}

	batchResult := AggregateResults(results, 0)

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

	batchResult := AggregateResults(results, 0)

	if len(batchResult.Successful) != 0 {
		t.Errorf("Expected 0 successful results, got %d", len(batchResult.Successful))
	}

	if len(batchResult.Failed) != 2 {
		t.Errorf("Expected 2 failed results, got %d", len(batchResult.Failed))
	}
}

func TestAggregateResults_Empty(t *testing.T) {
	results := []Result{}

	batchResult := AggregateResults(results, 0)

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
