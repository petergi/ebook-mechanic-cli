package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

type validateFlags struct {
	fileType string
}

func newValidateCmd(rootFlags *RootFlags) *cobra.Command {
	flags := &validateFlags{}

	cmd := &cobra.Command{
		Use:   "validate <file>|-",
		Short: "Validate a single EPUB or PDF file",
		Long: `Validate an EPUB or PDF file for errors, warnings, and informational issues.

The file can be provided as a path or read from stdin using '-'.
When reading from stdin, the --type flag is required to specify the file format.`,
		Example: `  # Validate a file
  ebm validate book.epub
  ebm validate document.pdf

  # Validate with JSON output
  ebm validate book.epub --format json

  # Validate from stdin
  cat book.epub | ebm validate - --type epub

  # Save report to file
  ebm validate book.epub --output report.md --format markdown

  # Filter by severity
  ebm validate book.epub --min-severity error
  ebm validate book.epub --severity error --severity warning`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context(), args[0], flags, rootFlags)
		},
	}

	cmd.Flags().StringVar(&flags.fileType, "type", "", "File type when reading from stdin (epub, pdf)")

	return cmd
}

func runValidate(ctx context.Context, target string, flags *validateFlags, rootFlags *RootFlags) error {
	// Create report options
	opts, err := NewReportOptions(rootFlags)
	if err != nil {
		return fmt.Errorf("invalid report options: %w", err)
	}

	var report *ebmlib.ValidationReport
	var validationErr error

	if target == "-" {
		// Validate from stdin
		if flags.fileType == "" {
			return fmt.Errorf("--type is required when reading from stdin (use: epub or pdf)")
		}

		report, validationErr = validateFromStdin(ctx, flags.fileType)
	} else {
		// Validate from file
		report, validationErr = validateFromFile(ctx, target)
	}

	// Handle validation errors
	if validationErr != nil {
		return fmt.Errorf("validation failed: %w", validationErr)
	}

	// Write the report
	if err := WriteReport(report, opts); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Exit with non-zero if file is invalid (but don't return error to avoid double printing)
	if !report.IsValid {
		osExit(1)
	}

	return nil
}

func validateFromFile(ctx context.Context, filePath string) (*ebmlib.ValidationReport, error) {
	op := operations.NewValidateOperation(ctx)
	return op.Execute(filePath)
}

func validateFromStdin(ctx context.Context, fileType string) (*ebmlib.ValidationReport, error) {
	// Normalize file type
	fileType = strings.ToLower(fileType)
	if fileType != "epub" && fileType != "pdf" {
		return nil, fmt.Errorf("invalid file type: %s (expected: epub or pdf)", fileType)
	}

	// Read all data from stdin
	reader := bufio.NewReader(os.Stdin)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data received from stdin")
	}

	// Create a temporary file for validation
	// (ebook-mechanic-lib expects a file path)
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("ebm-*.%s", fileType))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Validate the temporary file
	op := operations.NewValidateOperation(ctx)
	report, err := op.Execute(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	// Update the file path in the report to show it came from stdin
	report.FilePath = "<stdin>"

	return report, nil
}
