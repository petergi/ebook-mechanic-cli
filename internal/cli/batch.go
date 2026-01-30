package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
)

type batchFlags struct {
	jobs               int
	timeout            int
	recursive          bool
	maxDepth           int
	extensions         []string
	ignore             []string
	progress           string
	summaryOnly        bool
	backupDir          string
	continueOnError    bool
	noBackup           bool
	aggressive         bool
	skipValidation     bool
	removeSystemErrors bool
	moveFailedRepairs  bool
	cleanupEmptyDirs   bool
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

  # Batch repair in-place with backups (default)
  ebm batch repair ./books

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
	cmd.Flags().BoolVar(&flags.removeSystemErrors, "remove-system-errors", false, "Remove files with system errors after processing")
	cmd.Flags().BoolVar(&flags.cleanupEmptyDirs, "cleanup-empty-dirs", true, "Clean up empty parent directories after file removal")

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
		Example: `  # Repair all files in-place with backups (default)
  ebm batch repair ./books

  # Repair with custom backup directory
  ebm batch repair ./library --backup-dir ./backups

  # Repair without backups
  ebm batch repair ./books --no-backup

  # Repair with 4 workers
  ebm batch repair ./books --jobs 4`,
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
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory for backup files")
	cmd.Flags().BoolVar(&flags.noBackup, "no-backup", false, "Skip backup before in-place repair")
	cmd.Flags().BoolVar(&flags.aggressive, "aggressive", false, "Enable aggressive repairs (may drop content/structure)")
	cmd.Flags().BoolVar(&flags.continueOnError, "continue-on-error", true, "Continue processing on individual file errors")
	cmd.Flags().BoolVar(&flags.skipValidation, "skip-validation", false, "Skip post-repair validation")
	cmd.Flags().BoolVar(&flags.removeSystemErrors, "remove-system-errors", false, "Remove files with system errors after processing")
	cmd.Flags().BoolVar(&flags.moveFailedRepairs, "move-failed-repairs", false, "Move unrepairable files to INVALID folder")
	cmd.Flags().BoolVar(&flags.cleanupEmptyDirs, "cleanup-empty-dirs", true, "Clean up empty parent directories and Calibre metadata folders")

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

	// Record the options used for this batch
	batchResult.Options = operations.BatchOptions{
		NumWorkers:         flags.jobs,
		SkipValidation:     false, // N/A for validation
		NoBackup:           false, // N/A for validation
		Aggressive:         false, // N/A for validation
		RemoveSystemErrors: flags.removeSystemErrors,
		MoveFailedRepairs:  false, // N/A for validation
		CleanupEmptyDirs:   flags.cleanupEmptyDirs,
	}

	// Perform post-processing cleanup if requested
	if flags.removeSystemErrors && len(batchResult.Errored) > 0 {
		batchResult.RemovedFiles = make([]string, 0, len(batchResult.Errored))
		for _, r := range batchResult.Errored {
			batchResult.RemovedFiles = append(batchResult.RemovedFiles, r.FilePath)
			_ = os.Remove(r.FilePath)
			if flags.cleanupEmptyDirs {
				removeEmptyParentDirs(filepath.Dir(r.FilePath), dir)
			}
		}
	}

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

	if flags.noBackup && flags.backupDir != "" {
		return fmt.Errorf("--backup-dir is not supported with --no-backup")
	}
	if flags.aggressive {
		fmt.Fprintln(os.Stderr, "Warning: aggressive repairs may discard content or restructure the book.")
	}

	mode := operations.RepairSaveModeBackupOriginal
	if flags.noBackup {
		mode = operations.RepairSaveModeNoBackup
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
		Aggressive:   flags.aggressive,
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

	// Record the options used for this batch
	batchResult.Options = operations.BatchOptions{
		NumWorkers:         flags.jobs,
		SkipValidation:     flags.skipValidation,
		NoBackup:           flags.noBackup,
		Aggressive:         flags.aggressive,
		RemoveSystemErrors: flags.removeSystemErrors,
		MoveFailedRepairs:  flags.moveFailedRepairs,
		CleanupEmptyDirs:   flags.cleanupEmptyDirs,
	}

	// Perform post-processing cleanup if requested
	if flags.removeSystemErrors && len(batchResult.Errored) > 0 {
		batchResult.RemovedFiles = make([]string, 0, len(batchResult.Errored))
		for _, r := range batchResult.Errored {
			batchResult.RemovedFiles = append(batchResult.RemovedFiles, r.FilePath)
			_ = os.Remove(r.FilePath)
			if flags.cleanupEmptyDirs {
				removeEmptyParentDirs(filepath.Dir(r.FilePath), dir)
			}
		}
	}
	if flags.moveFailedRepairs && len(batchResult.Invalid) > 0 {
		invalidDir := filepath.Join(dir, "INVALID")
		_ = os.MkdirAll(invalidDir, 0755)
		batchResult.MovedFiles = make([]string, 0, len(batchResult.Invalid))
		for _, r := range batchResult.Invalid {
			batchResult.MovedFiles = append(batchResult.MovedFiles, r.FilePath)
			dstPath := filepath.Join(invalidDir, filepath.Base(r.FilePath))
			_ = os.Rename(r.FilePath, dstPath)
			if flags.cleanupEmptyDirs {
				removeEmptyParentDirs(filepath.Dir(r.FilePath), dir)
			}
		}
	}

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

// isCalibreMetadataDirectory checks if a directory contains only Calibre metadata files
// (no ebooks, only cover.jpg, metadata.opf, .DS_Store, etc.)
func isCalibreMetadataDirectory(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	// Empty directory
	if len(entries) == 0 {
		return true
	}

	// Check for ebook files - if any exist, not metadata-only
	hasEbook := false
	hasMetadata := false

	for _, entry := range entries {
		name := strings.ToLower(entry.Name())

		// Skip hidden files and common system files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			// Nested directory means not a simple Calibre metadata folder
			return false
		}

		ext := filepath.Ext(name)
		if ext == ".epub" || ext == ".pdf" {
			hasEbook = true
			break
		}

		// Check for Calibre metadata files
		if name == "cover.jpg" || name == "cover.jpeg" || name == "cover.png" ||
			name == "metadata.opf" {
			hasMetadata = true
		}
	}

	// If no ebooks and has metadata files, or just empty/system files, it's metadata-only
	return !hasEbook && (hasMetadata || len(entries) == 0)
}

// removeEmptyParentDirs removes empty parent directories and Calibre metadata-only directories up to the batch root
// Bails out if a parent has more than 3 siblings to avoid scanning large shallow hierarchies
func removeEmptyParentDirs(dir string, rootPath string) {
	// Ensure both paths are absolute and clean
	rootPath = filepath.Clean(rootPath)
	dir = filepath.Clean(dir)

	// Walk up the directory tree
	for dir != rootPath && strings.HasPrefix(dir, rootPath) {
		// Check sibling count before attempting removal
		parent := filepath.Dir(dir)
		if parent != rootPath { // Don't bail at root level
			entries, err := os.ReadDir(parent)
			if err == nil && len(entries) > 3 {
				// Too many siblings, skip this entire branch to avoid performance hit
				break
			}
		}

		// First try to remove if it's a Calibre metadata-only directory
		if isCalibreMetadataDirectory(dir) {
			// Remove all files in the directory first
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						_ = os.Remove(filepath.Join(dir, entry.Name()))
					}
				}
			}
		}

		// Try to remove the directory if empty
		err := os.Remove(dir)
		if err != nil {
			// Directory not empty or other error, stop
			break
		}

		// Move to parent
		dir = parent
	}
}
