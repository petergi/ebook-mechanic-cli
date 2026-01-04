package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

type repairFlags struct {
	output    string
	inPlace   bool
	backup    bool
	backupDir string
}

func newRepairCmd(rootFlags *RootFlags) *cobra.Command {
	flags := &repairFlags{}

	cmd := &cobra.Command{
		Use:   "repair <file>",
		Short: "Repair a single EPUB or PDF file",
		Long: `Repair an EPUB or PDF file by fixing detected issues.

The repaired file can be written to a new location or replace the original.
Automatic backups are supported before in-place repairs.`,
		Example: `  # Repair to a new file
  ebm repair book.epub --output fixed.epub

  # Repair in-place with backup
  ebm repair book.epub --in-place --backup

  # Repair with custom backup directory
  ebm repair book.epub --in-place --backup --backup-dir ./backups

  # Repair with JSON report
  ebm repair book.epub --output fixed.epub --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepair(cmd.Context(), args[0], flags, rootFlags)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output path for repaired file")
	cmd.Flags().BoolVar(&flags.inPlace, "in-place", false, "Repair file in place (replaces original)")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backup before in-place repair")
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory for backup files (default: same as input)")

	return cmd
}

func runRepair(ctx context.Context, filePath string, flags *repairFlags, rootFlags *RootFlags) error {
	// Validate flags
	if !flags.inPlace && flags.output == "" {
		return fmt.Errorf("either --in-place or --output must be specified")
	}

	if flags.inPlace && flags.output != "" {
		return fmt.Errorf("--in-place and --output are mutually exclusive")
	}

	if flags.backup && !flags.inPlace {
		return fmt.Errorf("--backup requires --in-place")
	}

	if flags.backupDir != "" && !flags.backup {
		return fmt.Errorf("--backup-dir requires --backup")
	}

	// Create report options
	opts, err := NewReportOptions(rootFlags)
	if err != nil {
		return fmt.Errorf("invalid report options: %w", err)
	}

	// Perform repair
	var result *ebmlib.RepairResult
	var repairErr error
	var outputPath string

	op := operations.NewRepairOperation(ctx)

	if flags.inPlace {
		// In-place repair
		result, repairErr = repairInPlace(op, filePath, flags.backup, flags.backupDir)
		outputPath = filePath
	} else {
		// Repair to new file
		result, repairErr = repairToOutput(op, filePath, flags.output)
		outputPath = flags.output
	}

	// Handle repair errors
	if repairErr != nil {
		return fmt.Errorf("repair failed: %w", repairErr)
	}

	// Validate the repaired file
	var validationReport *ebmlib.ValidationReport
	validateOp := operations.NewValidateOperation(ctx)
	validationReport, _ = validateOp.Execute(outputPath)

	// Write the repair report
	if err := WriteRepairReport(result, validationReport, opts); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with non-zero if repair failed (but don't return error to avoid double printing)
	if !result.Success {
		os.Exit(1)
	}

	return nil
}

func repairInPlace(op *operations.RepairOperation, filePath string, createBackup bool, backupDir string) (*ebmlib.RepairResult, error) {
	var backupPath string

	if createBackup {
		// Create backup
		if backupDir == "" {
			backupDir = filepath.Dir(filePath)
		}

		// Ensure backup directory exists
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Generate backup filename
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		backupPath = filepath.Join(backupDir, fmt.Sprintf("%s.backup%s", name, ext))

		// Copy original to backup
		if err := copyFile(filePath, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Perform repair
	result, err := op.Execute(filePath)
	if err != nil {
		return nil, err
	}

	// Add backup path to result
	if backupPath != "" {
		result.BackupPath = backupPath
	}

	return result, nil
}

func repairToOutput(op *operations.RepairOperation, filePath, outputPath string) (*ebmlib.RepairResult, error) {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy input to output location
	if err := copyFile(filePath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to copy file to output location: %w", err)
	}

	// Perform repair on the output file
	result, err := op.Execute(outputPath)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
