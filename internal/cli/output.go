package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// OutputFormat represents the type of output format
type OutputFormat int

const (
	FormatText OutputFormat = iota
	FormatJSON
	FormatMarkdown
)

// ParseFormat converts a string to OutputFormat
func ParseFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(s) {
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return FormatText, fmt.Errorf("invalid format: %s (valid: text, json, markdown)", s)
	}
}

// Formatter defines the interface for output formatters
type Formatter interface {
	FormatValidation(report *ebmlib.ValidationReport) string
	FormatRepair(result *ebmlib.RepairResult, report *ebmlib.ValidationReport) string
	FormatBatchValidation(result *operations.BatchResult, summaryOnly bool) string
	FormatBatchRepair(result *operations.BatchResult, summaryOnly bool) string
}

// NewFormatter creates a formatter based on format and options
func NewFormatter(format OutputFormat, colorEnabled bool) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatMarkdown:
		return &MarkdownFormatter{}
	default:
		return &TextFormatter{ColorEnabled: colorEnabled}
	}
}

// TextFormatter formats output as human-readable text
type TextFormatter struct {
	ColorEnabled bool
}

func (f *TextFormatter) FormatValidation(report *ebmlib.ValidationReport) string {
	var b strings.Builder

	// Header
	b.WriteString(f.header("Validation Report"))
	b.WriteString("\n")
	b.WriteString(f.field("File", report.FilePath))
	b.WriteString("\n")

	// Overall status
	if report.IsValid {
		b.WriteString(f.success("✓ File is valid!"))
	} else {
		b.WriteString(f.error("✗ File has errors"))
	}
	b.WriteString("\n\n")

	// Summary
	b.WriteString(f.subheader("Summary"))
	b.WriteString(f.field("Errors", fmt.Sprintf("%d", report.ErrorCount())))
	b.WriteString(f.field("Warnings", fmt.Sprintf("%d", report.WarningCount())))
	b.WriteString(f.field("Info", fmt.Sprintf("%d", report.InfoCount())))
	b.WriteString("\n")

	// Issues
	if len(report.Errors) > 0 {
		b.WriteString(f.subheader("Errors"))
		for _, err := range report.Errors {
			b.WriteString(f.formatIssue(err, "error"))
		}
		b.WriteString("\n")
	}

	if len(report.Warnings) > 0 {
		b.WriteString(f.subheader("Warnings"))
		for _, warn := range report.Warnings {
			b.WriteString(f.formatIssue(warn, "warning"))
		}
		b.WriteString("\n")
	}

	if len(report.Info) > 0 {
		b.WriteString(f.subheader("Information"))
		for _, info := range report.Info {
			b.WriteString(f.formatIssue(info, "info"))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *TextFormatter) FormatRepair(result *ebmlib.RepairResult, report *ebmlib.ValidationReport) string {
	var b strings.Builder

	// Header
	b.WriteString(f.header("Repair Report"))
	b.WriteString("\n")

	// Status
	if result.Success {
		b.WriteString(f.success("✓ Repair successful!"))
	} else {
		b.WriteString(f.error("✗ Repair failed"))
		if result.Error != nil {
			b.WriteString("\n")
			b.WriteString(f.error(result.Error.Error()))
		}
	}
	b.WriteString("\n\n")

	// Backup path
	if result.BackupPath != "" {
		b.WriteString(f.field("Backup", result.BackupPath))
		b.WriteString("\n")
	}

	// Actions applied
	if len(result.ActionsApplied) > 0 {
		b.WriteString(f.subheader("Actions Applied"))
		for _, action := range result.ActionsApplied {
			b.WriteString(f.success("  ✓ " + action.Description))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Include validation report if present
	if report != nil {
		b.WriteString("\n")
		b.WriteString(f.FormatValidation(report))
	}

	return b.String()
}

func (f *TextFormatter) formatIssue(issue ebmlib.ValidationError, issueType string) string {
	var icon, prefix string
	if f.ColorEnabled {
		switch issueType {
		case "error":
			icon = color.RedString("✗")
			prefix = color.RedString("[%s]", issue.Code)
		case "warning":
			icon = color.YellowString("⚠")
			prefix = color.YellowString("[%s]", issue.Code)
		case "info":
			icon = color.CyanString("ℹ")
			prefix = color.CyanString("[%s]", issue.Code)
		}
	} else {
		icon = "•"
		prefix = fmt.Sprintf("[%s]", issue.Code)
	}

	line := fmt.Sprintf("  %s %s %s\n", icon, prefix, issue.Message)

	if issue.Location != nil {
		location := fmt.Sprintf("      Location: %s", issue.Location.File)
		if issue.Location.Line > 0 {
			location += fmt.Sprintf(":%d", issue.Location.Line)
		}
		line += f.muted(location) + "\n"
	}

	return line
}

func (f *TextFormatter) header(s string) string {
	if f.ColorEnabled {
		return color.New(color.Bold, color.FgCyan).Sprintf("═══ %s ═══\n", s)
	}
	return fmt.Sprintf("=== %s ===\n", s)
}

func (f *TextFormatter) subheader(s string) string {
	if f.ColorEnabled {
		return color.New(color.Bold).Sprintf("%s:\n", s)
	}
	return fmt.Sprintf("%s:\n", s)
}

func (f *TextFormatter) field(key, value string) string {
	if f.ColorEnabled {
		return fmt.Sprintf("  %s: %s\n", color.CyanString(key), value)
	}
	return fmt.Sprintf("  %s: %s\n", key, value)
}

func (f *TextFormatter) success(s string) string {
	if f.ColorEnabled {
		return color.GreenString(s)
	}
	return s
}

func (f *TextFormatter) error(s string) string {
	if f.ColorEnabled {
		return color.RedString(s)
	}
	return s
}

func (f *TextFormatter) muted(s string) string {
	if f.ColorEnabled {
		return color.New(color.Faint).Sprint(s)
	}
	return s
}

func (f *TextFormatter) FormatBatchValidation(result *operations.BatchResult, summaryOnly bool) string {
	var b strings.Builder

	// Header
	b.WriteString(f.header("Batch Validation Report"))
	b.WriteString("\n")

	// Summary
	b.WriteString(f.subheader("Summary"))
	b.WriteString(f.field("Total Files", fmt.Sprintf("%d", result.Total)))
	b.WriteString(f.field("Valid", fmt.Sprintf("%d", len(result.Valid))))
	b.WriteString(f.field("Invalid", fmt.Sprintf("%d", len(result.Invalid))))
	if len(result.Errored) > 0 {
		b.WriteString(f.field("System Errors", fmt.Sprintf("%d", len(result.Errored))))
	}
	b.WriteString(f.field("Duration", result.Duration.Round(time.Millisecond).String()))
	b.WriteString("\n")

	// Overall status
	if len(result.Invalid) == 0 && len(result.Errored) == 0 {
		b.WriteString(f.success("✓ All files are valid!"))
	} else {
		b.WriteString(f.error(fmt.Sprintf("✗ Found %d invalid file(s)", len(result.Invalid))))
		if len(result.Errored) > 0 {
			b.WriteString("\n")
			b.WriteString(f.error(fmt.Sprintf("⚠ Encountered %d system error(s)", len(result.Errored))))
		}
	}
	b.WriteString("\n\n")

	// Problematic files (Invalid or Errored)
	if !summaryOnly && (len(result.Invalid) > 0 || len(result.Errored) > 0) {
		b.WriteString(f.subheader("Issues Found"))

		// List system errors first
		for _, r := range result.Errored {
			fileName := filepath.Base(r.FilePath)
			b.WriteString(f.error(fmt.Sprintf("  ⚠ %s: System Error: %s\n", fileName, r.Error)))
		}

		// List invalid files
		for _, r := range result.Invalid {
			fileName := filepath.Base(r.FilePath)
			if r.Report != nil && !r.Report.IsValid {
				b.WriteString(f.error(fmt.Sprintf("  ✗ %s: %d errors\n", fileName, r.Report.ErrorCount())))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *TextFormatter) FormatBatchRepair(result *operations.BatchResult, summaryOnly bool) string {
	var b strings.Builder

	// Header
	b.WriteString(f.header("Batch Repair Report"))
	b.WriteString("\n")

	// Summary
	b.WriteString(f.subheader("Summary"))
	b.WriteString(f.field("Total Files", fmt.Sprintf("%d", result.Total)))
	b.WriteString(f.field("Successfully Repaired", fmt.Sprintf("%d", len(result.Valid))))
	b.WriteString(f.field("Repair Failed", fmt.Sprintf("%d", len(result.Invalid))))
	if len(result.Errored) > 0 {
		b.WriteString(f.field("System Errors", fmt.Sprintf("%d", len(result.Errored))))
	}
	b.WriteString(f.field("Duration", result.Duration.Round(time.Millisecond).String()))
	b.WriteString("\n")

	// Overall status
	if len(result.Invalid) == 0 && len(result.Errored) == 0 {
		b.WriteString(f.success("✓ All files repaired successfully!"))
	} else {
		b.WriteString(f.error(fmt.Sprintf("✗ %d file(s) failed to repair", len(result.Invalid))))
		if len(result.Errored) > 0 {
			b.WriteString("\n")
			b.WriteString(f.error(fmt.Sprintf("⚠ Encountered %d system error(s)", len(result.Errored))))
		}
	}
	b.WriteString("\n\n")

	// Problematic files
	if !summaryOnly && (len(result.Invalid) > 0 || len(result.Errored) > 0) {
		b.WriteString(f.subheader("Issues Found"))

		// List system errors first
		for _, r := range result.Errored {
			fileName := filepath.Base(r.FilePath)
			b.WriteString(f.error(fmt.Sprintf("  ⚠ %s: System Error: %s\n", fileName, r.Error)))
		}

		// List failed repairs
		for _, r := range result.Invalid {
			fileName := filepath.Base(r.FilePath)
			errMsg := "unknown error"
			if r.Repair != nil && r.Repair.Error != nil {
				errMsg = r.Repair.Error.Error()
			}
			b.WriteString(f.error(fmt.Sprintf("  ✗ %s: %s\n", fileName, errMsg)))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

func (f *JSONFormatter) FormatValidation(report *ebmlib.ValidationReport) string {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal report: %s"}`, err)
	}
	return string(data)
}

func (f *JSONFormatter) FormatRepair(result *ebmlib.RepairResult, report *ebmlib.ValidationReport) string {
	output := map[string]interface{}{
		"success":         result.Success,
		"backup_path":     result.BackupPath,
		"actions_applied": result.ActionsApplied,
	}

	if result.Error != nil {
		output["error"] = result.Error.Error()
	}

	if report != nil {
		output["validation_report"] = report
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal result: %s"}`, err)
	}
	return string(data)
}

func (f *JSONFormatter) FormatBatchValidation(result *operations.BatchResult, summaryOnly bool) string {
	output := map[string]interface{}{
		"total":      result.Total,
		"successful": len(result.Successful),
		"failed":     len(result.Failed),
		"duration":   result.Duration.Milliseconds(),
	}

	if !summaryOnly {
		output["results"] = result
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal batch result: %s"}`, err)
	}
	return string(data)
}

func (f *JSONFormatter) FormatBatchRepair(result *operations.BatchResult, summaryOnly bool) string {
	output := map[string]interface{}{
		"total":      result.Total,
		"successful": len(result.Successful),
		"failed":     len(result.Failed),
		"duration":   result.Duration.Milliseconds(),
	}

	if !summaryOnly {
		output["results"] = result
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal batch result: %s"}`, err)
	}
	return string(data)
}

// MarkdownFormatter formats output as GitHub-flavored Markdown
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) FormatValidation(report *ebmlib.ValidationReport) string {
	var b strings.Builder

	// Header
	b.WriteString("# Validation Report\n\n")
	b.WriteString(fmt.Sprintf("**File:** `%s`\n\n", report.FilePath))

	// Status
	if report.IsValid {
		b.WriteString("**Status:** ✅ Valid\n\n")
	} else {
		b.WriteString("**Status:** ❌ Has Errors\n\n")
	}

	// Summary table
	b.WriteString("## Summary\n\n")
	b.WriteString("| Type | Count |\n")
	b.WriteString("|------|-------|\n")
	b.WriteString(fmt.Sprintf("| Errors | %d |\n", report.ErrorCount()))
	b.WriteString(fmt.Sprintf("| Warnings | %d |\n", report.WarningCount()))
	b.WriteString(fmt.Sprintf("| Info | %d |\n\n", report.InfoCount()))

	// Issues
	if len(report.Errors) > 0 {
		b.WriteString("## Errors\n\n")
		for _, err := range report.Errors {
			b.WriteString(f.formatIssue(err))
		}
		b.WriteString("\n")
	}

	if len(report.Warnings) > 0 {
		b.WriteString("## Warnings\n\n")
		for _, warn := range report.Warnings {
			b.WriteString(f.formatIssue(warn))
		}
		b.WriteString("\n")
	}

	if len(report.Info) > 0 {
		b.WriteString("## Information\n\n")
		for _, info := range report.Info {
			b.WriteString(f.formatIssue(info))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *MarkdownFormatter) FormatRepair(result *ebmlib.RepairResult, report *ebmlib.ValidationReport) string {
	var b strings.Builder

	// Header
	b.WriteString("# Repair Report\n\n")

	// Status
	if result.Success {
		b.WriteString("**Status:** ✅ Repair Successful\n\n")
	} else {
		b.WriteString("**Status:** ❌ Repair Failed\n\n")
		if result.Error != nil {
			b.WriteString(fmt.Sprintf("**Error:** %s\n\n", result.Error.Error()))
		}
	}

	// Backup
	if result.BackupPath != "" {
		b.WriteString(fmt.Sprintf("**Backup:** `%s`\n\n", result.BackupPath))
	}

	// Actions
	if len(result.ActionsApplied) > 0 {
		b.WriteString("## Actions Applied\n\n")
		for _, action := range result.ActionsApplied {
			b.WriteString(fmt.Sprintf("- ✅ %s\n", action.Description))
		}
		b.WriteString("\n")
	}

	// Validation report
	if report != nil {
		b.WriteString("---\n\n")
		b.WriteString(f.FormatValidation(report))
	}

	return b.String()
}

func (f *MarkdownFormatter) formatIssue(issue ebmlib.ValidationError) string {
	line := fmt.Sprintf("- **[%s]** %s", issue.Code, issue.Message)

	if issue.Location != nil {
		location := fmt.Sprintf("`%s", issue.Location.File)
		if issue.Location.Line > 0 {
			location += fmt.Sprintf(":%d", issue.Location.Line)
		}
		location += "`"
		line += fmt.Sprintf("\n  - Location: %s", location)
	}

	return line + "\n"
}

func (f *MarkdownFormatter) FormatBatchValidation(result *operations.BatchResult, summaryOnly bool) string {
	var b strings.Builder

	// Header
	b.WriteString("# Batch Validation Report\n\n")

	// Summary table
	b.WriteString("## Summary\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Total Files | %d |\n", result.Total))
	b.WriteString(fmt.Sprintf("| Successful | %d |\n", len(result.Successful)))
	b.WriteString(fmt.Sprintf("| Failed | %d |\n", len(result.Failed)))
	b.WriteString(fmt.Sprintf("| Duration | %s |\n\n", result.Duration.Round(time.Millisecond)))

	// Status
	if len(result.Failed) == 0 {
		b.WriteString("**Status:** ✅ All files valid\n\n")
	} else {
		b.WriteString(fmt.Sprintf("**Status:** ❌ %d file(s) have errors\n\n", len(result.Failed)))
	}

	// Failed files
	if !summaryOnly && len(result.Failed) > 0 {
		b.WriteString("## Failed Files\n\n")
		for _, r := range result.Failed {
			fileName := filepath.Base(r.FilePath)
			if r.Error != nil {
				b.WriteString(fmt.Sprintf("- ❌ **%s**: %s\n", fileName, r.Error))
			} else if r.Report != nil && !r.Report.IsValid {
				b.WriteString(fmt.Sprintf("- ❌ **%s**: %d errors\n", fileName, r.Report.ErrorCount()))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *MarkdownFormatter) FormatBatchRepair(result *operations.BatchResult, summaryOnly bool) string {
	var b strings.Builder

	// Header
	b.WriteString("# Batch Repair Report\n\n")

	// Summary table
	b.WriteString("## Summary\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Total Files | %d |\n", result.Total))
	b.WriteString(fmt.Sprintf("| Successful | %d |\n", len(result.Successful)))
	b.WriteString(fmt.Sprintf("| Failed | %d |\n", len(result.Failed)))
	b.WriteString(fmt.Sprintf("| Duration | %s |\n\n", result.Duration.Round(time.Millisecond)))

	// Status
	if len(result.Failed) == 0 {
		b.WriteString("**Status:** ✅ All files repaired successfully\n\n")
	} else {
		b.WriteString(fmt.Sprintf("**Status:** ❌ %d file(s) failed to repair\n\n", len(result.Failed)))
	}

	// Failed files
	if !summaryOnly && len(result.Failed) > 0 {
		b.WriteString("## Failed Files\n\n")
		for _, r := range result.Failed {
			fileName := filepath.Base(r.FilePath)
			if r.Error != nil {
				b.WriteString(fmt.Sprintf("- ❌ **%s**: %s\n", fileName, r.Error))
			} else if r.Repair != nil && !r.Repair.Success {
				errMsg := "unknown error"
				if r.Repair.Error != nil {
					errMsg = r.Repair.Error.Error()
				}
				b.WriteString(fmt.Sprintf("- ❌ **%s**: %s\n", fileName, errMsg))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// WriteOutput writes formatted output to a writer or file
func WriteOutput(w io.Writer, content string) error {
	_, err := fmt.Fprint(w, content)
	return err
}
