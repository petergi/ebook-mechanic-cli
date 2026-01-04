# ADR 0005: Concurrency Model for Batch Processing

## Status

Accepted

## Context

The ebook-mechanic-cli needs to efficiently process multiple ebook files in batch operations (validate/repair multiple files). We need to decide:

- How to parallelize file processing
- How to manage worker goroutines
- How to report progress for concurrent operations
- How to handle errors in concurrent processing
- How to ensure UI remains responsive during batch operations

### Requirements

1. **Performance**: Process multiple files concurrently to maximize throughput
2. **Resource Control**: Limit concurrent operations to prevent resource exhaustion
3. **Progress Reporting**: Real-time progress updates to the UI
4. **Error Handling**: Graceful handling of individual file errors without stopping batch
5. **Cancellation**: Allow users to cancel long-running batch operations
6. **UI Responsiveness**: Never block the UI thread

### Considered Alternatives

1. **Sequential Processing**

   - Pros: Simple, predictable
   - Cons: Slow for large batches, doesn't utilize multi-core

2. **Unbounded Concurrency** (goroutine per file)

   - Pros: Maximum parallelism
   - Cons: Resource exhaustion, system overload, hard to control

3. **Worker Pool Pattern**

   - Pros: Controlled concurrency, predictable resource usage, good throughput
   - Cons: More complex implementation

4. **errgroup Package**

   - Pros: Built-in error handling, easy cancellation
   - Cons: Stops on first error (not desired for batch), less control

## Decision

We will implement a **Worker Pool Pattern** with the following characteristics:

- **Fixed number of workers** (default: `runtime.NumCPU()`, configurable)
- **Task queue** using buffered channels
- **Result aggregation** with concurrent-safe collection
- **Progress tracking** via atomic counters and periodic updates
- **Context-based cancellation** for graceful shutdown
- **Bubbletea messages** for UI updates (non-blocking)

## Rationale

### Worker Pool Benefits

1. **Controlled Concurrency**: Limit workers to prevent system overload
2. **Predictable Performance**: Consistent resource usage
3. **Backpressure**: Task queue naturally limits memory usage
4. **Flexibility**: Easy to adjust worker count based on system resources

### Why Not errgroup

While `errgroup` is excellent for many use cases, it's designed to fail-fast:

- Stops all workers on first error
- We need to continue processing and collect all results
- We want fine-grained error collection per file

### Why Not Unbounded Goroutines

Creating a goroutine per file:

- Could spawn thousands of goroutines for large batches
- Each operation involves file I/O (expensive resources)
- Could exhaust file descriptors, memory, or CPU
- No backpressure mechanism

### Progress Reporting via Messages

Bubbletea's message-passing architecture:

- Workers send progress messages to UI
- UI updates asynchronously
- No blocking or shared state
- Natural fit for concurrent operations

## Architecture

### Worker Pool Structure

```go
type BatchProcessor struct {
    numWorkers  int
    taskQueue   chan Task
    resultQueue chan Result
    progressCh  chan ProgressMsg
    ctx         context.Context
    cancel      context.CancelFunc
}

type Task struct {
    FilePath  string
    Operation OperationType // Validate or Repair
}

type Result struct {
    FilePath string
    Report   *ValidationReport
    Error    error
}

type ProgressMsg struct {
    Completed int
    Total     int
    Current   string
}
```

### Worker Implementation

```go
func (bp *BatchProcessor) worker(id int) {
    for {
        select {
        case <-bp.ctx.Done():
            return // Graceful shutdown
        case task, ok := <-bp.taskQueue:
            if !ok {
                return // Queue closed
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
            bp.updateProgress(task.FilePath)
        }
    }
}
```

### Batch Execution

```go
func (bp *BatchProcessor) Execute(files []string, op OperationType) []Result {
    // Start workers
    for i := 0; i < bp.numWorkers; i++ {
        go bp.worker(i)
    }

    // Feed tasks
    go func() {
        for _, file := range files {
            select {
            case bp.taskQueue <- Task{FilePath: file, Operation: op}:
            case <-bp.ctx.Done():
                return
            }
        }
        close(bp.taskQueue)
    }()

    // Collect results
    results := make([]Result, 0, len(files))
    for i := 0; i < len(files); i++ {
        select {
        case result := <-bp.resultQueue:
            results = append(results, result)
        case <-bp.ctx.Done():
            return results
        }
    }

    return results
}
```

## Consequences

### Positive

1. **Scalability**: Efficiently processes 10s to 1000s of files
2. **Resource Control**: Prevents system overload
3. **Responsiveness**: UI remains responsive during batch processing
4. **Error Isolation**: One file's error doesn't stop others
5. **Cancellation**: User can cancel anytime
6. **Progress Tracking**: Real-time progress updates

### Negative

1. **Complexity**: More complex than sequential processing
2. **Testing**: Concurrent code is harder to test
3. **Debugging**: Race conditions and timing issues possible

### Mitigation

- Comprehensive testing with race detector (`go test -race`)
- Clear worker lifecycle management
- Use atomic operations for shared counters
- Extensive logging for debugging
- Integration tests with various batch sizes

## Implementation Details

### Configuration

```go
type BatchConfig struct {
    NumWorkers    int           // Number of concurrent workers
    QueueSize     int           // Task queue buffer size
    ProgressRate  time.Duration // Progress update frequency
    Timeout       time.Duration // Per-file operation timeout
}

func DefaultBatchConfig() BatchConfig {
    return BatchConfig{
        NumWorkers:   runtime.NumCPU(),
        QueueSize:    100,
        ProgressRate: 100 * time.Millisecond,
        Timeout:      30 * time.Second,
    }
}
```

### Progress Tracking

Use atomic counters for thread-safe progress:

```go
type ProgressTracker struct {
    total     int
    completed atomic.Int64
    current   atomic.Value // string
}

func (pt *ProgressTracker) Update(filePath string) {
    pt.completed.Add(1)
    pt.current.Store(filePath)
}

func (pt *ProgressTracker) Progress() (int, int, string) {
    completed := int(pt.completed.Load())
    current := pt.current.Load().(string)
    return completed, pt.total, current
}
```

### Error Collection

Collect all errors without stopping:

```go
type BatchResult struct {
    Successful []Result
    Failed     []Result
    Duration   time.Duration
}

func (bp *BatchProcessor) collectResults(count int) BatchResult {
    result := BatchResult{
        Successful: make([]Result, 0),
        Failed:     make([]Result, 0),
    }

    for i := 0; i < count; i++ {
        r := <-bp.resultQueue
        if r.Error != nil {
            result.Failed = append(result.Failed, r)
        } else {
            result.Successful = append(result.Successful, r)
        }
    }

    return result
}
```

### Integration with Bubbletea

Send progress messages to UI:

```go
// In worker
func (bp *BatchProcessor) updateProgress(filePath string) {
    progress := bp.tracker.Progress()
    msg := ProgressMsg{
        Completed: progress.Completed,
        Total:     progress.Total,
        Current:   filePath,
    }

    // Non-blocking send to UI
    select {
    case bp.progressCh <- msg:
    default:
        // Skip if channel is full (UI is busy)
    }
}

// In Bubbletea model
func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ProgressMsg:
        m.completed = msg.Completed
        m.total = msg.Total
        m.current = msg.Current
        return m, m.listenForProgress()
    }
    return m, nil
}

func (m progressModel) listenForProgress() tea.Cmd {
    return func() tea.Msg {
        return <-m.progressCh
    }
}
```

### Cancellation

Support user cancellation:

```go
// User presses ESC or Ctrl+C
func (m batchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
            m.processor.Cancel() // Triggers context cancellation
            return m, nil
        }
    }
    return m, nil
}
```

## Performance Considerations

### Worker Count Tuning

Default: `runtime.NumCPU()`

**Rationale**:

- File validation/repair is CPU-bound (parsing, checksums)
- I/O is mostly sequential within each operation
- NumCPU provides good balance

**Tuning**:

- User can override via config
- May increase for I/O-heavy operations
- May decrease for resource-constrained systems

### Queue Sizing

Default: 100 tasks

**Rationale**:

- Provides backpressure
- Prevents memory exhaustion with large batches
- Small enough to not waste memory

### Progress Update Rate

Default: 100ms

**Rationale**:

- Responsive enough for user feedback
- Not so frequent as to impact performance
- Bubbletea handles rendering efficiently

## Testing Strategy

### Unit Tests

Test worker pool components in isolation:

```go
func TestWorkerPool_ProcessTask(t *testing.T) {
    bp := NewBatchProcessor(DefaultBatchConfig())
    task := Task{FilePath: "test.epub", Operation: Validate}

    result := bp.processTask(task)

    assert.NotNil(t, result)
    assert.Equal(t, "test.epub", result.FilePath)
}
```

### Concurrency Tests

Use race detector:

```go
func TestBatchProcessor_Concurrent(t *testing.T) {
    // Run with: go test -race
    bp := NewBatchProcessor(BatchConfig{NumWorkers: 4})

    files := make([]string, 100)
    for i := range files {
        files[i] = fmt.Sprintf("file%d.epub", i)
    }

    results := bp.Execute(files, Validate)

    assert.Equal(t, len(files), len(results))
}
```

### Integration Tests

Test with real files:

```go
func TestBatchProcessor_Integration(t *testing.T) {
    bp := NewBatchProcessor(DefaultBatchConfig())

    // Create test files
    tmpDir := t.TempDir()
    files := createTestFiles(t, tmpDir, 10)

    results := bp.Execute(files, Validate)

    assert.Equal(t, 10, len(results))
    // Verify each result
}
```

### Performance Benchmarks

Measure throughput:

```go
func BenchmarkBatchProcessor_100Files(b *testing.B) {
    files := createBenchFiles(b, 100)
    bp := NewBatchProcessor(DefaultBatchConfig())

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bp.Execute(files, Validate)
    }
}
```

## Monitoring

### Metrics to Track

- **Throughput**: Files processed per second
- **Latency**: Average time per file
- **Error Rate**: Percentage of failed files
- **Worker Utilization**: Busy vs idle time

### Debugging

Enable verbose logging:

```go
func (bp *BatchProcessor) worker(id int) {
    log.Printf("Worker %d started", id)
    defer log.Printf("Worker %d stopped", id)

    for task := range bp.taskQueue {
        log.Printf("Worker %d processing %s", id, task.FilePath)
        // ...
    }
}
```

## Future Enhancements

1. **Dynamic Worker Scaling**: Adjust workers based on load
2. **Priority Queue**: Process high-priority files first
3. **Work Stealing**: Load balancing between workers
4. **Batch Resumption**: Resume interrupted batch operations
5. **Distributed Processing**: Process across multiple machines

## References

- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Worker Pool Pattern](https://gobyexample.com/worker-pools)
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)
