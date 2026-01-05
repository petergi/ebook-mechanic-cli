package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
		wantErr  bool
	}{
		{"text", FormatText, false},
		{"TEXT", FormatText, false},
		{"json", FormatJSON, false},
		{"markdown", FormatMarkdown, false},
		{"md", FormatMarkdown, false},
		{"invalid", FormatText, true},
	}

	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestTextFormatter_FormatValidation(t *testing.T) {
	f := &TextFormatter{ColorEnabled: false}
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  false,
		Errors: []ebmlib.ValidationError{
			{Code: "ERR1", Message: "Error 1", Severity: ebmlib.SeverityError},
		},
		Warnings: []ebmlib.ValidationError{
			{Code: "WARN1", Message: "Warning 1", Severity: ebmlib.SeverityWarning},
		},
		Info: []ebmlib.ValidationError{
			{Code: "INFO1", Message: "Info 1", Severity: ebmlib.SeverityInfo},
		},
	}

	output := f.FormatValidation(report)

	if !strings.Contains(output, "Validation Report") {
		t.Error("Output missing title")
	}
	if !strings.Contains(output, "test.epub") {
		t.Error("Output missing filename")
	}
	if !strings.Contains(output, "[ERR1]") {
		t.Error("Output missing error code")
	}
}

func TestJSONFormatter_FormatValidation(t *testing.T) {
	f := &JSONFormatter{}
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  true,
	}

	output := f.FormatValidation(report)

	var parsed ebmlib.ValidationReport
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed.FilePath != "test.epub" {
		t.Errorf("Expected filepath 'test.epub', got '%s'", parsed.FilePath)
	}
}

func TestMarkdownFormatter_FormatValidation(t *testing.T) {
	f := &MarkdownFormatter{}
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  false,
		Errors: []ebmlib.ValidationError{
			{Code: "ERR1", Message: "Error 1"},
		},
	}

	output := f.FormatValidation(report)

	if !strings.Contains(output, "# Validation Report") {
		t.Error("Output missing header")
	}
	if !strings.Contains(output, "`test.epub`") {
		t.Error("Output missing code block for filename")
	}
}

func TestTextFormatter_FormatBatchValidation(t *testing.T) {
	f := &TextFormatter{ColorEnabled: false}
	result := &operations.BatchResult{
		Total:    10,
		Valid:    []operations.Result{{FilePath: "valid.epub"}},
		Invalid:  []operations.Result{{FilePath: "invalid.epub"}},
		Errored:  []operations.Result{{FilePath: "error.epub", Error: assertError("io error")}},
		Duration: time.Second,
	}

	output := f.FormatBatchValidation(result, false)

	if !strings.Contains(output, "Batch Validation Report") {
		t.Error("Output missing title")
	}
	if !strings.Contains(output, "Valid: 1") {
		t.Error("Output missing valid count")
	}
	if !strings.Contains(output, "Invalid: 1") {
		t.Error("Output missing invalid count")
	}
	if !strings.Contains(output, "System Errors: 1") {
		t.Error("Output missing error count")
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   OutputFormat
		expected string
	}{
		{FormatText, "*cli.TextFormatter"},
		{FormatJSON, "*cli.JSONFormatter"},
		{FormatMarkdown, "*cli.MarkdownFormatter"},
	}

	for _, tt := range tests {
		got := NewFormatter(tt.format, true)
		typeName := fmt.Sprintf("%T", got)
		if typeName != tt.expected {
			t.Errorf("NewFormatter(%v) = %s, want %s", tt.format, typeName, tt.expected)
		}
	}
}

func TestTextFormatter_FormatRepair(t *testing.T) {
	f := &TextFormatter{ColorEnabled: false}
	result := &ebmlib.RepairResult{
		Success:    true,
		BackupPath: "backup.epub",
		ActionsApplied: []ebmlib.RepairAction{
			{Description: "Action 1"},
		},
	}

	output := f.FormatRepair(result, nil)
	if !strings.Contains(output, "Repair Report") {
		t.Error("Output missing title")
	}
	if !strings.Contains(output, "backup.epub") {
		t.Error("Output missing backup path")
	}
}

func TestTextFormatter_FormatBatchRepair(t *testing.T) {
	f := &TextFormatter{ColorEnabled: false}
	result := &operations.BatchResult{
		Valid:   []operations.Result{{FilePath: "valid.epub"}},
		Invalid: []operations.Result{{FilePath: "invalid.epub", Repair: &ebmlib.RepairResult{Success: false}}},
	}

	output := f.FormatBatchRepair(result, false)
	if !strings.Contains(output, "Batch Repair Report") {
		t.Error("Output missing title")
	}
}

func TestJSONFormatter_FormatRepair(t *testing.T) {
	f := &JSONFormatter{}
	result := &ebmlib.RepairResult{Success: true}
	output := f.FormatRepair(result, nil)
	if !strings.Contains(output, `"success": true`) {
		t.Error("JSON output missing success field")
	}
}

func TestMarkdownFormatter_FormatRepair(t *testing.T) {
	f := &MarkdownFormatter{}
	result := &ebmlib.RepairResult{Success: true}
	output := f.FormatRepair(result, nil)
	if !strings.Contains(output, "# Repair Report") {
		t.Error("Markdown output missing header")
	}
}

func TestWriteOutput(t *testing.T) {
	var b strings.Builder
	err := WriteOutput(&b, "test content")
	if err != nil {
		t.Errorf("WriteOutput failed: %v", err)
	}
	if b.String() != "test content" {
		t.Errorf("Expected 'test content', got %q", b.String())
	}
}

func TestTextFormatter_ColorEnabled(t *testing.T) {
	f := &TextFormatter{ColorEnabled: true}

	s := f.header("Title")
	if s == "" {
		t.Error("Empty color header")
	}

	s = f.success("Success")
	if s == "" {
		t.Error("Empty color success")
	}

	s = f.error("Error")
	if s == "" {
		t.Error("Empty color error")
	}

	s = f.muted("Muted")
	if s == "" {
		t.Error("Empty color muted")
	}
}

func assertError(msg string) error {
	return &simpleError{msg}
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string { return e.msg }

func TestJSONFormatter_FormatBatchValidation(t *testing.T) {
	f := &JSONFormatter{}
	result := &operations.BatchResult{
		Total: 1,
		Valid: []operations.Result{{FilePath: "v.epub"}},
	}
	output := f.FormatBatchValidation(result, false)
	if !strings.Contains(output, `"total": 1`) {
		t.Error("JSON output missing total")
	}
}

func TestJSONFormatter_FormatBatchRepair(t *testing.T) {
	f := &JSONFormatter{}
	result := &operations.BatchResult{
		Total: 1,
		Valid: []operations.Result{{FilePath: "v.epub"}},
	}
	output := f.FormatBatchRepair(result, false)
	if !strings.Contains(output, `"total": 1`) {
		t.Error("JSON output missing total")
	}
}

func TestMarkdownFormatter_FormatBatchValidation(t *testing.T) {
	f := &MarkdownFormatter{}
	result := &operations.BatchResult{
		Total: 1,
		Valid: []operations.Result{{FilePath: "v.epub"}},
	}
	output := f.FormatBatchValidation(result, false)
	if !strings.Contains(output, "# Batch Validation Report") {
		t.Error("Markdown output missing header")
	}
}

func TestMarkdownFormatter_FormatBatchRepair(t *testing.T) {
	f := &MarkdownFormatter{}
	result := &operations.BatchResult{
		Total: 1,
		Valid: []operations.Result{{FilePath: "v.epub"}},
	}
	output := f.FormatBatchRepair(result, false)
	if !strings.Contains(output, "# Batch Repair Report") {
		t.Error("Markdown output missing header")
	}
}
