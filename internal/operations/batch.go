package operations

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// OperationType represents the type of batch operation
type OperationType string

const (
	// OperationValidate for batch validation
	OperationValidate OperationType = "validate"
	// OperationRepair for batch repair
	OperationRepair OperationType = "repair"
)

// BatchConfig configures batch processing behavior
type BatchConfig struct {
	NumWorkers   int           // Number of concurrent workers
	QueueSize    int           // Task queue buffer size
	ProgressRate time.Duration // Progress update frequency
	Timeout      time.Duration // Per-file operation timeout
}

// DefaultBatchConfig returns sensible defaults for batch processing
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		NumWorkers:   runtime.NumCPU(),
		QueueSize:    100,
		ProgressRate: 100 * time.Millisecond,
		Timeout:      30 * time.Second,
	}
}

// Task represents a single file to process
type Task struct {
	FilePath  string
	Operation OperationType
}

// Result contains the result of processing a single file
type Result struct {
	FilePath string
	Report   *ebmlib.ValidationReport
	Repair   *ebmlib.RepairResult
	Error    error
}

// ProgressUpdate contains progress information
type ProgressUpdate struct {
	Completed int
	Total     int
	Current   string
}

// BatchProcessor handles concurrent batch processing of files
type BatchProcessor struct {
	config      BatchConfig
	ctx         context.Context
	cancel      context.CancelFunc
	taskQueue   chan Task
	resultQueue chan Result
	progressCh  chan ProgressUpdate
	completed   atomic.Int64
	total       int
	currentFile atomic.Value // stores string
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(config BatchConfig) *BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &BatchProcessor{
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
		taskQueue:   make(chan Task, config.QueueSize),
		resultQueue: make(chan Result, config.QueueSize),
		progressCh:  make(chan ProgressUpdate, 10),
	}
}

// Execute processes a batch of files with the given operation
func (bp *BatchProcessor) Execute(files []string, operation OperationType) []Result {
	bp.total = len(files)
	bp.completed.Store(0)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bp.config.NumWorkers; i++ {
		wg.Add(1)
		go bp.worker(i, &wg)
	}

	// Feed tasks
	go func() {
		for _, file := range files {
			select {
			case bp.taskQueue <- Task{FilePath: file, Operation: operation}:
			case <-bp.ctx.Done():
				return
			}
		}
		close(bp.taskQueue)
	}()

	// Start progress reporter
	go bp.reportProgress()

	// Collect results
	results := make([]Result, 0, len(files))
	for i := 0; i < len(files); i++ {
		select {
		case result := <-bp.resultQueue:
			results = append(results, result)
		case <-bp.ctx.Done():
			// Collect any remaining results
			for len(bp.resultQueue) > 0 {
				results = append(results, <-bp.resultQueue)
			}
			return results
		}
	}

	// Wait for all workers to finish
	wg.Wait()

	return results
}

// worker processes tasks from the queue
func (bp *BatchProcessor) worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case task, ok := <-bp.taskQueue:
			if !ok {
				return
			}

			// Process task
			result := bp.processTask(task)

			// Send result
			select {
			case bp.resultQueue <- result:
			case <-bp.ctx.Done():
				return
			}

			// Update progress
			bp.completed.Add(1)
			bp.currentFile.Store(task.FilePath)
		}
	}
}

// processTask processes a single task
func (bp *BatchProcessor) processTask(task Task) Result {
	// Create timeout context for this operation
	ctx, cancel := context.WithTimeout(bp.ctx, bp.config.Timeout)
	defer cancel()

	result := Result{
		FilePath: task.FilePath,
	}

	switch task.Operation {
	case OperationValidate:
		validator := NewValidateOperation(ctx)
		report, err := validator.Execute(task.FilePath)
		result.Report = report
		result.Error = err

	case OperationRepair:
		repairer := NewRepairOperation(ctx)
		repairResult, err := repairer.Execute(task.FilePath)
		result.Repair = repairResult
		result.Error = err

	default:
		result.Error = nil
	}

	return result
}

// reportProgress periodically sends progress updates
func (bp *BatchProcessor) reportProgress() {
	ticker := time.NewTicker(bp.config.ProgressRate)
	defer ticker.Stop()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case <-ticker.C:
			completed := int(bp.completed.Load())
			current := ""
			if val := bp.currentFile.Load(); val != nil {
				current = val.(string)
			}

			update := ProgressUpdate{
				Completed: completed,
				Total:     bp.total,
				Current:   current,
			}

			// Non-blocking send
			select {
			case bp.progressCh <- update:
			default:
				// Skip if channel is full
			}
		}
	}
}

// ProgressChannel returns the channel for receiving progress updates
func (bp *BatchProcessor) ProgressChannel() <-chan ProgressUpdate {
	return bp.progressCh
}

// Cancel cancels the batch processing
func (bp *BatchProcessor) Cancel() {
	bp.cancel()
}

// BatchResult contains aggregated results of a batch operation
type BatchResult struct {
	Successful []Result
	Failed     []Result
	Duration   time.Duration
	Total      int
}

// AggregateResults aggregates a list of results into a BatchResult
func AggregateResults(results []Result, duration time.Duration) BatchResult {
	br := BatchResult{
		Successful: make([]Result, 0),
		Failed:     make([]Result, 0),
		Duration:   duration,
		Total:      len(results),
	}

	for _, r := range results {
		if r.Error != nil {
			br.Failed = append(br.Failed, r)
		} else {
			br.Successful = append(br.Successful, r)
		}
	}

	return br
}
