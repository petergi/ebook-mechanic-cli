package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// Severity represents issue severity levels
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// ParseSeverity converts a string to Severity
func ParseSeverity(s string) (Severity, error) {
	switch strings.ToLower(s) {
	case "info":
		return SeverityInfo, nil
	case "warning":
		return SeverityWarning, nil
	case "error":
		return SeverityError, nil
	default:
		return SeverityInfo, fmt.Errorf("invalid severity: %s (valid: info, warning, error)", s)
	}
}

// SeverityFilter filters validation issues by severity
type SeverityFilter struct {
	MinSeverity *Severity
	Severities  map[Severity]bool
	MaxErrors   int
}

// NewSeverityFilter creates a filter from flags
func NewSeverityFilter(minSeverity string, severities []string, maxErrors int) (*SeverityFilter, error) {
	filter := &SeverityFilter{
		Severities: make(map[Severity]bool),
		MaxErrors:  maxErrors,
	}

	// Parse minimum severity
	if minSeverity != "" {
		sev, err := ParseSeverity(minSeverity)
		if err != nil {
			return nil, err
		}
		filter.MinSeverity = &sev
	}

	// Parse specific severities
	for _, s := range severities {
		sev, err := ParseSeverity(s)
		if err != nil {
			return nil, err
		}
		filter.Severities[sev] = true
	}

	return filter, nil
}

// FilterReport applies severity filtering to a validation report
func (f *SeverityFilter) FilterReport(report *ebmlib.ValidationReport) *ebmlib.ValidationReport {
	if report == nil {
		return nil
	}

	filtered := &ebmlib.ValidationReport{
		FilePath: report.FilePath,
		IsValid:  report.IsValid,
	}

	// Apply filters
	errorCount := 0

	// Filter errors
	if f.shouldInclude(SeverityError) {
		for _, err := range report.Errors {
			if f.MaxErrors > 0 && errorCount >= f.MaxErrors {
				break
			}
			filtered.Errors = append(filtered.Errors, err)
			errorCount++
		}
	}

	// Filter warnings
	if f.shouldInclude(SeverityWarning) {
		for _, warn := range report.Warnings {
			filtered.Warnings = append(filtered.Warnings, warn)
		}
	}

	// Filter info
	if f.shouldInclude(SeverityInfo) {
		for _, info := range report.Info {
			filtered.Info = append(filtered.Info, info)
		}
	}

	return filtered
}

func (f *SeverityFilter) shouldInclude(sev Severity) bool {
	// If specific severities are set, only include those
	if len(f.Severities) > 0 {
		return f.Severities[sev]
	}

	// If minimum severity is set, include this and higher severities
	if f.MinSeverity != nil {
		return sev >= *f.MinSeverity
	}

	// Include all by default
	return true
}

// ReportOptions contains options for report generation
type ReportOptions struct {
	Format       OutputFormat
	Formatter    Formatter
	Filter       *SeverityFilter
	OutputPath   string
	ColorEnabled bool
	Verbose      bool
	SummaryOnly  bool
}

// NewReportOptions creates report options from flags
func NewReportOptions(flags *RootFlags) (*ReportOptions, error) {
	format, err := ParseFormat(flags.Format)
	if err != nil {
		return nil, err
	}

	filter, err := NewSeverityFilter(flags.MinSeverity, flags.Severities, flags.MaxErrors)
	if err != nil {
		return nil, err
	}

	// Disable color for non-terminal output or if explicitly disabled
	colorEnabled := flags.Color
	if flags.Output != "" || format != FormatText {
		colorEnabled = false
	}

	return &ReportOptions{
		Format:       format,
		Formatter:    NewFormatter(format, colorEnabled),
		Filter:       filter,
		OutputPath:   flags.Output,
		ColorEnabled: colorEnabled,
		Verbose:      flags.Verbose,
		SummaryOnly:  false, // Default to false, batch commands will set it
	}, nil
}

// WriteReport writes a formatted and filtered validation report
func WriteReport(report *ebmlib.ValidationReport, opts *ReportOptions) error {
	if report == nil {
		return fmt.Errorf("no report to write")
	}

	// Apply severity filtering
	filtered := opts.Filter.FilterReport(report)

	// Format the report
	content := opts.Formatter.FormatValidation(filtered)

	// Write to file or stdout
	if opts.OutputPath != "" {
		return os.WriteFile(opts.OutputPath, []byte(content), 0644)
	}

	return WriteOutput(os.Stdout, content)
}

// WriteRepairReport writes a formatted repair result
func WriteRepairReport(result *ebmlib.RepairResult, report *ebmlib.ValidationReport, opts *ReportOptions) error {
	if result == nil {
		return fmt.Errorf("no repair result to write")
	}

	// Apply severity filtering to validation report if present
	var filtered *ebmlib.ValidationReport
	if report != nil {
		filtered = opts.Filter.FilterReport(report)
	}

	// Format the repair result
	content := opts.Formatter.FormatRepair(result, filtered)

	// Write to file or stdout
	if opts.OutputPath != "" {
		return os.WriteFile(opts.OutputPath, []byte(content), 0644)
	}

	return WriteOutput(os.Stdout, content)
}

// WriteBatchValidationReport writes a formatted batch validation result
func WriteBatchValidationReport(result *operations.BatchResult, opts *ReportOptions) error {
	if result == nil {
		return fmt.Errorf("no batch result to write")
	}

	// Format the batch result
	content := opts.Formatter.FormatBatchValidation(result, opts.SummaryOnly)

	// Write to file or stdout
	if opts.OutputPath != "" {
		return os.WriteFile(opts.OutputPath, []byte(content), 0644)
	}

	return WriteOutput(os.Stdout, content)
}

// WriteBatchRepairReport writes a formatted batch repair result
func WriteBatchRepairReport(result *operations.BatchResult, opts *ReportOptions) error {
	if result == nil {
		return fmt.Errorf("no batch result to write")
	}

	// Format the batch result
	content := opts.Formatter.FormatBatchRepair(result, opts.SummaryOnly)

	// Write to file or stdout
	if opts.OutputPath != "" {
		return os.WriteFile(opts.OutputPath, []byte(content), 0644)
	}

	return WriteOutput(os.Stdout, content)
}
