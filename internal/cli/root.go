package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// osExit is a copy of os.Exit that can be mocked in tests
var osExit = os.Exit

// RootFlags contains global flags shared across all commands
type RootFlags struct {
	Format      string
	Output      string
	Verbose     bool
	Color       bool
	MinSeverity string
	Severities  []string
	MaxErrors   int
}

// NewRootCmd creates the root command for the CLI
func NewRootCmd() *cobra.Command {
	flags := &RootFlags{}

	cmd := &cobra.Command{
		Use:   "ebm",
		Short: "Validate and repair EPUB and PDF files",
		Long: `ebm - Ebook Mechanic CLI

Validate and repair EPUB and PDF files with comprehensive reporting.

  - Run without arguments to launch the interactive TUI.
  - Run with a file or directory path to quick-validate (e.g. 'ebm mybook.epub').
  - Use subcommands ('repair', 'batch') for specific operations.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  # Interactive TUI mode
  ebm

  # CLI validation
  ebm validate book.epub
  ebm validate book.epub --format json --output report.json

  # CLI repair
  ebm repair book.epub --in-place --backup
  ebm repair book.epub --output fixed.epub

  # Batch operations
  ebm batch validate ./books --jobs 8
  ebm batch repair ./library --in-place --backup`,
	}

	// Global flags available to all commands
	cmd.PersistentFlags().StringVarP(&flags.Format, "format", "f", "text", "Output format: text, json, markdown")
	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "", "Write report to file instead of stdout")
	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&flags.Color, "color", true, "Enable colorized output")
	cmd.PersistentFlags().StringVar(&flags.MinSeverity, "min-severity", "", "Minimum severity to include (info, warning, error)")
	cmd.PersistentFlags().StringSliceVar(&flags.Severities, "severity", nil, "Include only specific severities (repeatable)")
	cmd.PersistentFlags().IntVar(&flags.MaxErrors, "max-errors", 0, "Limit number of errors per report (0 = unlimited)")

	// Default run behavior: if args are provided, try to validate them
	cmd.Args = cobra.ArbitraryArgs
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Should have been handled by main.go (TUI mode), but if we get here:
			return c.Help()
		}

		// Check if first arg is a file or directory
		target := args[0]
		info, err := os.Stat(target)
		if err != nil {
			// If not a file/dir, and looks like a flag or unknown command, show help/error
			return fmt.Errorf("unknown command or file: %q\n\nRun 'ebm --help' for usage", target)
		}

		// If it's a directory, run batch validate
		if info.IsDir() {
			batchFlags := &batchFlags{
				// Set reasonable defaults for implicit batch mode
				recursive:       true,
				continueOnError: true,
				maxDepth:        -1, // -1 means unlimited depth
				// We can't easily access default values from flags here without parsing again or duplicating defaults
				// but 0 int / false bool is usually fine, or we set specifics:
				jobs: 0, // 0 will use runtime.NumCPU() in logic if we update it, or we set it here
			}
			// We need to ensure jobs is set to something valid if the struct defaults aren't used
			// The runBatchValidate function expects flags to be populated.
			// Let's rely on the fact that we can reuse the logic but we need to initialize flags manually
			// or better: just invoke the validate/batch command explicitly?
			// Cobra doesn't easily let us "redirect" to another command object's RunE without parsing its flags.

			// Simpler approach: Call the logic function directly
			// We need to handle the "jobs" default logic which is inside the flag definition usually.
			// Let's update runBatchValidate to handle 0 jobs = NumCPU
			return runBatchValidate(c.Context(), target, batchFlags, flags)
		}

		// If it's a file, run single validate
		validateFlags := &validateFlags{}
		return runValidate(c.Context(), target, validateFlags, flags)
	}

	// Add subcommands
	cmd.AddCommand(newValidateCmd(flags))
	cmd.AddCommand(newRepairCmd(flags))
	cmd.AddCommand(newBatchCmd(flags))
	cmd.AddCommand(NewCompletionCmd(cmd))

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	return cmd
}

// Execute runs the CLI command
func Execute() error {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
