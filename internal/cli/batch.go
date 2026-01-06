package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
)

type batchFlags struct {
	jobs            int
	timeout         int
	recursive       bool
	maxDepth        int
	extensions      []string
	ignore          []string
	progress        string
	summaryOnly     bool
	inPlace         bool
	backup          bool
	backupDir       string
	saveMode        string
	continueOnError bool
}

func newBatchCmd(rootFlags *RootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch operations on multiple files",
		Long: `Process multiple EPUB and PDF files concurrently.

Batch operations use worker pools for efficient parallel processing.
The number of workers can be configured with the --jobs flag.`,
		Example: `  # Batch validate all files in a directory
  ebm batch validate ./books

  # Batch validate with custom worker count
  ebm batch validate ./library --jobs 8

  # Batch repair in-place with backups
  ebm batch repair ./books --in-place --backup

  # Batch operations with JSON output
  ebm batch validate ./library --format json --output results.json`,
	}

	cmd.AddCommand(newBatchValidateCmd(rootFlags))
	cmd.AddCommand(newBatchRepairCmd(rootFlags))

	return cmd
}

func newBatchValidateCmd(rootFlags *RootFlags) *cobra.Command {
	flags := &batchFlags{}

	cmd := &cobra.Command{
		Use:   "validate <directory>",
		Short: "Validate multiple EPUB and PDF files",
		Long: `Validate all EPUB and PDF files in a directory concurrently.

Files are processed using a worker pool for efficient parallel validation.
Progress is displayed in real-time showing completed/total files.`,
		Example: `  # Validate all files in directory
  ebm batch validate ./books

  # Validate with 8 workers
  ebm batch validate ./library --jobs 8

  # Validate recursively with JSON output
  ebm batch validate ./books --recursive --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchValidate(cmd.Context(), args[0], flags, rootFlags)
		},
	}

	cmd.Flags().IntVarP(&flags.jobs, "jobs", "j", runtime.NumCPU(), "Number of concurrent workers")
	cmd.Flags().IntVar(&flags.timeout, "timeout", 30, "Timeout per file in seconds")
	cmd.Flags().BoolVarP(&flags.recursive, "recursive", "r", true, "Process subdirectories recursively")
	cmd.Flags().IntVar(&flags.maxDepth, "max-depth", -1, "Maximum directory depth (-1 = unlimited)")
	cmd.Flags().StringSliceVar(&flags.extensions, "ext", nil, "File extensions to include (default: .epub, .pdf)")
	cmd.Flags().StringSliceVar(&flags.ignore, "ignore", nil, "Glob patterns to ignore")
	cmd.Flags().StringVar(&flags.progress, "progress", "auto", "Progress output mode (auto, simple, none)")
	cmd.Flags().BoolVar(&flags.summaryOnly, "summary-only", false, "Only print summary output")
	cmd.Flags().BoolVar(&flags.continueOnError, "continue-on-error", true, "Continue processing on individual file errors")

	return cmd
}

func newBatchRepairCmd(rootFlags *RootFlags) *cobra.Command {
	flags := &batchFlags{}

	cmd := &cobra.Command{
		Use:   "repair <directory>",
		Short: "Repair multiple EPUB and PDF files",
		Long: `Repair all EPUB and PDF files in a directory concurrently.

Files are processed using a worker pool for efficient parallel repairs.
Progress is displayed in real-time showing completed/total files.`,
		Example: `  # Repair all files in-place with backups
  ebm batch repair ./books --in-place --backup

  # Repair with custom backup directory
  ebm batch repair ./library --in-place --backup --backup-dir ./backups

  # Repair by renaming repaired files
  ebm batch repair ./books --in-place --save-mode rename-repaired

  # Repair with 4 workers
  ebm batch repair ./books --in-place --jobs 4`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchRepair(cmd.Context(), args[0], flags, rootFlags)
		},
	}

	cmd.Flags().IntVarP(&flags.jobs, "jobs", "j", runtime.NumCPU(), "Number of concurrent workers")
	cmd.Flags().IntVar(&flags.timeout, "timeout", 60, "Timeout per file in seconds")
	cmd.Flags().BoolVarP(&flags.recursive, "recursive", "r", true, "Process subdirectories recursively")
	cmd.Flags().IntVar(&flags.maxDepth, "max-depth", -1, "Maximum directory depth (-1 = unlimited)")
	cmd.Flags().StringSliceVar(&flags.extensions, "ext", nil, "File extensions to include (default: .epub, .pdf)")
	cmd.Flags().StringSliceVar(&flags.ignore, "ignore", nil, "Glob patterns to ignore")
	cmd.Flags().StringVar(&flags.progress, "progress", "auto", "Progress output mode (auto, simple, none)")
	cmd.Flags().BoolVar(&flags.summaryOnly, "summary-only", false, "Only print summary output")
	cmd.Flags().BoolVar(&flags.inPlace, "in-place", false, "Repair files in place")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backups before in-place repair (legacy alias for --save-mode backup-original)")
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory for backup files")
	cmd.Flags().StringVar(&flags.saveMode, "save-mode", "", "Save mode for in-place repairs: backup-original, rename-repaired, or no-backup")
	cmd.Flags().BoolVar(&flags.continueOnError, "continue-on-error", true, "Continue processing on individual file errors")

	return cmd
}

func runBatchValidate(ctx context.Context, dir string, flags *batchFlags, rootFlags *RootFlags) error {
	// Handle default jobs if not set (e.g. called from root command)
	if flags.jobs <= 0 {
		flags.jobs = runtime.NumCPU()
	}

	// Find all matching files
	findOpts := operations.FindFilesOptions{
		Recursive:  flags.recursive,
		MaxDepth:   flags.maxDepth,
		Extensions: flags.extensions,
		Ignore:     flags.ignore,
	}
	files, err := operations.FindFiles(dir, findOpts)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no matching files found in %s", dir)
	}

	// Create batch processor
	config := operations.BatchConfig{
		NumWorkers:   flags.jobs,
		QueueSize:    100,
		ProgressRate: 100 * time.Millisecond,
		Timeout:      time.Duration(flags.timeout) * time.Second,
	}
	processor := operations.NewBatchProcessor(ctx, config)

	// Set up progress reporting
	done := make(chan struct{})
	var bar *progressbar.ProgressBar

	if flags.progress != "none" {
		if flags.progress == "simple" || (flags.progress == "auto" && !isTerminal()) {
			// Simple text progress
			go func() {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						update := <-processor.ProgressChannel()
						fmt.Fprintf(os.Stderr, "Progress: %d/%d files completed...\n", update.Completed, update.Total)
					}
				}
			}()
		} else {
			// Progress bar (auto with terminal or explicit auto/bar)
			bar = progressbar.NewOptions(len(files),
				progressbar.OptionSetDescription("Validating"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowCount(),
				progressbar.OptionShowIts(),
				progressbar.OptionSetWidth(40),
				progressbar.OptionThrottle(100*time.Millisecond),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionEnableColorCodes(rootFlags.Color),
			)

			go func() {
				for {
					select {
					case <-done:
						return
					case update := <-processor.ProgressChannel():
						_ = bar.Set(update.Completed)
					}
				}
			}()
		}
	}

	// Execute batch validation
	start := time.Now()
	results := processor.Execute(files, operations.OperationValidate)
	duration := time.Since(start)
	close(done)
	if bar != nil {
		_ = bar.Finish()
	}

	// Aggregate results
	batchResult := operations.AggregateResults(results, duration, operations.OperationValidate)

	// Create report options
	opts, err := NewReportOptions(rootFlags)
	if err != nil {
		return fmt.Errorf("invalid report options: %w", err)
	}
	opts.SummaryOnly = flags.summaryOnly

	// Write batch report
	if err := WriteBatchValidationReport(&batchResult, opts); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with non-zero if any files failed
	if len(batchResult.Invalid) > 0 || len(batchResult.Errored) > 0 {
		osExit(1)
	}

	return nil
}

func runBatchRepair(ctx context.Context, dir string, flags *batchFlags, rootFlags *RootFlags) error {
	// Handle default jobs if not set
	if flags.jobs <= 0 {
		flags.jobs = runtime.NumCPU()
	}

	// Validate flags
	if !flags.inPlace {
		return fmt.Errorf("batch repair currently only supports --in-place mode")
	}

	mode, err := resolveRepairSaveMode(flags.saveMode, flags.backup)
	if err != nil {
		return err
	}
	if (mode == operations.RepairSaveModeRenameRepaired || mode == operations.RepairSaveModeNoBackup) && flags.backupDir != "" {
		return fmt.Errorf("--backup-dir is not supported with save-mode rename-repaired or no-backup")
	}

	// Find all matching files
	findOpts := operations.FindFilesOptions{
		Recursive:  flags.recursive,
		MaxDepth:   flags.maxDepth,
		Extensions: flags.extensions,
		Ignore:     flags.ignore,
	}
	files, err := operations.FindFiles(dir, findOpts)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no matching files found in %s", dir)
	}

	// Create backup directory if needed
	if mode == operations.RepairSaveModeBackupOriginal && flags.backupDir != "" {
		if err := os.MkdirAll(flags.backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}
	}

	// Create batch processor
	config := operations.BatchConfig{
		NumWorkers:   flags.jobs,
		QueueSize:    100,
		ProgressRate: 100 * time.Millisecond,
		Timeout:      time.Duration(flags.timeout) * time.Second,
		RepairMode:   mode,
		BackupDir:    flags.backupDir,
	}
	processor := operations.NewBatchProcessor(ctx, config)

	// Set up progress reporting
	done := make(chan struct{})
	var bar *progressbar.ProgressBar

	if flags.progress != "none" {
		if flags.progress == "simple" || (flags.progress == "auto" && !isTerminal()) {
			// Simple text progress
			go func() {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						update := <-processor.ProgressChannel()
						fmt.Fprintf(os.Stderr, "Progress: %d/%d files completed...\n", update.Completed, update.Total)
					}
				}
			}()
		} else {
			// Progress bar
			bar = progressbar.NewOptions(len(files),
				progressbar.OptionSetDescription("Repairing"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowCount(),
				progressbar.OptionShowIts(),
				progressbar.OptionSetWidth(40),
				progressbar.OptionThrottle(100*time.Millisecond),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionEnableColorCodes(rootFlags.Color),
			)

			go func() {
				for {
					select {
					case <-done:
						return
					case update := <-processor.ProgressChannel():
						_ = bar.Set(update.Completed)
					}
				}
			}()
		}
	}

	// Execute batch repair
	start := time.Now()
	results := processor.Execute(files, operations.OperationRepair)
	duration := time.Since(start)
	close(done)
	if bar != nil {
		_ = bar.Finish()
	}

	// Aggregate results
	batchResult := operations.AggregateResults(results, duration, operations.OperationRepair)

	// Create report options
	opts, err := NewReportOptions(rootFlags)
	if err != nil {
		return fmt.Errorf("invalid report options: %w", err)
	}
	opts.SummaryOnly = flags.summaryOnly

	// Write batch report
	if err := WriteBatchRepairReport(&batchResult, opts); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with non-zero if any files failed
	if len(batchResult.Invalid) > 0 || len(batchResult.Errored) > 0 {
		osExit(1)
	}

	return nil
}

// isTerminal returns true if stderr is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stderr.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
