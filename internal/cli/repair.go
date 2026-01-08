package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

type repairFlags struct {
	backupDir    string
	skipValidate bool
	noBackup     bool
	aggressive   bool
}

func newRepairCmd(rootFlags *RootFlags) *cobra.Command {
	flags := &repairFlags{}

	cmd := &cobra.Command{
		Use:   "repair <file>",
		Short: "Repair a single EPUB or PDF file",
		Long: `Repair an EPUB or PDF file by fixing detected issues.

Repairs run in-place by default. Backups are created unless disabled.`,
		Example: `  # Repair in-place with backup (default)
  ebm repair book.epub

  # Repair with custom backup directory
  ebm repair book.epub --backup-dir ./backups

  # Repair without backup
  ebm repair book.epub --no-backup

  # Repair without post-validation
  ebm repair book.epub --skip-validate`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepair(cmd.Context(), args[0], flags, rootFlags)
		},
	}

	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory for backup files (default: same as input)")
	cmd.Flags().BoolVar(&flags.skipValidate, "skip-validate", false, "Skip post-repair validation")
	cmd.Flags().BoolVar(&flags.noBackup, "no-backup", false, "Skip backup before in-place repair")
	cmd.Flags().BoolVar(&flags.aggressive, "aggressive", false, "Enable aggressive repairs (may drop content/structure)")

	return cmd
}

func runRepair(ctx context.Context, filePath string, flags *repairFlags, rootFlags *RootFlags) error {
	// Validate flags
	if flags.noBackup && flags.backupDir != "" {
		return fmt.Errorf("--backup-dir is not supported with --no-backup")
	}
	if flags.aggressive {
		fmt.Fprintln(os.Stderr, "Warning: aggressive repairs may discard content or restructure the book.")
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

	op := operations.NewRepairOperation(ctx).WithAggressive(flags.aggressive)

	mode := operations.RepairSaveModeBackupOriginal
	if flags.noBackup {
		mode = operations.RepairSaveModeNoBackup
	}

	result, outputPath, repairErr = repairInPlace(op, filePath, mode, flags.backupDir)

	// Handle repair errors
	if repairErr != nil {
		return fmt.Errorf("repair failed: %w", repairErr)
	}

	// Validate the repaired file unless skipped
	var validationReport *ebmlib.ValidationReport
	if !flags.skipValidate && result.Success && outputPath != "" {
		validateOp := operations.NewValidateOperation(ctx)
		validationReport, _ = validateOp.Execute(outputPath)
	}

	// Write the repair report
	if err := WriteRepairReport(result, validationReport, opts); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with non-zero if repair failed (but don't return error to avoid double printing)
	if !result.Success {
		osExit(1)
	}

	return nil
}

func repairInPlace(op *operations.RepairOperation, filePath string, mode operations.RepairSaveMode, backupDir string) (*ebmlib.RepairResult, string, error) {
	return op.ExecuteWithSaveMode(filePath, mode, backupDir)
}
