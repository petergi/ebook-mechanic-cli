package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected Severity
		wantErr  bool
	}{
		{"info", SeverityInfo, false},
		{"warning", SeverityWarning, false},
		{"error", SeverityError, false},
		{"INFO", SeverityInfo, false},
		{"invalid", SeverityInfo, true},
	}

	for _, tt := range tests {
		got, err := ParseSeverity(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseSeverity(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSeverityFilter_FilterReport(t *testing.T) {
	report := &ebmlib.ValidationReport{
		Errors: []ebmlib.ValidationError{
			{Message: "Error 1", Severity: ebmlib.SeverityError},
			{Message: "Error 2", Severity: ebmlib.SeverityError},
		},
		Warnings: []ebmlib.ValidationError{
			{Message: "Warning 1", Severity: ebmlib.SeverityWarning},
		},
		Info: []ebmlib.ValidationError{
			{Message: "Info 1", Severity: ebmlib.SeverityInfo},
		},
	}

	tests := []struct {
		name        string
		minSeverity string
		maxErrors   int
		wantErrors  int
		wantWarns   int
		wantInfo    int
	}{
		{"No filter", "", 0, 2, 1, 1},
		{"Min Warning", "warning", 0, 2, 1, 0},
		{"Min Error", "error", 0, 2, 0, 0},
		{"Max Errors 1", "", 1, 1, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewSeverityFilter(tt.minSeverity, nil, tt.maxErrors)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}

			filtered := filter.FilterReport(report)

			if len(filtered.Errors) != tt.wantErrors {
				t.Errorf("Want %d errors, got %d", tt.wantErrors, len(filtered.Errors))
			}
			if len(filtered.Warnings) != tt.wantWarns {
				t.Errorf("Want %d warnings, got %d", tt.wantWarns, len(filtered.Warnings))
			}
			if len(filtered.Info) != tt.wantInfo {
				t.Errorf("Want %d info, got %d", tt.wantInfo, len(filtered.Info))
			}
		})
	}
}

func TestNewReportOptions(t *testing.T) {
	flags := &RootFlags{
		Format: "json",
		Color:  true,
	}

	opts, err := NewReportOptions(flags)
	if err != nil {
		t.Fatalf("NewReportOptions failed: %v", err)
	}

	if opts.Format != FormatJSON {
		t.Errorf("Expected FormatJSON, got %v", opts.Format)
	}

	if opts.ColorEnabled {
		t.Error("Expected ColorEnabled false for JSON format")
	}
}

func TestNewReportOptions_Severities(t *testing.T) {
	flags := &RootFlags{
		Format:      "text",
		MinSeverity: "error",
		Severities:  []string{"warning"},
		MaxErrors:   10,
	}
	opts, err := NewReportOptions(flags)
	if err != nil {
		t.Fatalf("NewReportOptions failed: %v", err)
	}
	if opts.Filter.MaxErrors != 10 {
		t.Errorf("Expected MaxErrors 10, got %d", opts.Filter.MaxErrors)
	}
}

func TestWriteReport_Markdown(t *testing.T) {
	flags := &RootFlags{Format: "markdown"}
	opts, _ := NewReportOptions(flags)
	report := &ebmlib.ValidationReport{FilePath: "test.epub", IsValid: false}
	_ = WriteReport(report, opts)
}

func TestWriteRepairReport_WithError(t *testing.T) {
	flags := &RootFlags{Format: "text"}
	opts, _ := NewReportOptions(flags)
	result := &ebmlib.RepairResult{Success: false, Error: assertError("fail")}
	_ = WriteRepairReport(result, nil, opts)
}

func TestWriteReport_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.txt")

	flags := &RootFlags{
		Output: outputPath,
		Format: "text",
	}
	opts, _ := NewReportOptions(flags)

	report := &ebmlib.ValidationReport{FilePath: "test.epub", IsValid: true}
	err := WriteReport(report, opts)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}
}

func TestWriteReport_Stdout(t *testing.T) {
	flags := &RootFlags{Format: "text"}
	opts, _ := NewReportOptions(flags)
	report := &ebmlib.ValidationReport{FilePath: "test.epub", IsValid: true}
	err := WriteReport(report, opts)
	if err != nil {
		t.Errorf("WriteReport failed: %v", err)
	}
}

func TestWriteRepairReport(t *testing.T) {
	flags := &RootFlags{Format: "text"}
	opts, _ := NewReportOptions(flags)
	result := &ebmlib.RepairResult{Success: true}
	err := WriteRepairReport(result, nil, opts)
	if err != nil {
		t.Errorf("WriteRepairReport failed: %v", err)
	}
}

func TestWriteReport_Nil(t *testing.T) {
	err := WriteReport(nil, nil)
	if err == nil {
		t.Error("Expected error for nil report")
	}
}

func TestWriteRepairReport_Nil(t *testing.T) {
	err := WriteRepairReport(nil, nil, nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestWriteBatchValidationReport_Nil(t *testing.T) {
	err := WriteBatchValidationReport(nil, nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestWriteBatchRepairReport_Nil(t *testing.T) {
	err := WriteBatchRepairReport(nil, nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestWriteBatchValidationReport_Error(t *testing.T) {
	flags := &RootFlags{Output: "/invalid/path", Format: "text"}
	opts, err := NewReportOptions(flags)
	if err != nil {
		t.Fatalf("NewReportOptions failed: %v", err)
	}
	result := &operations.BatchResult{}
	err = WriteBatchValidationReport(result, opts)
	if err == nil {
		t.Error("Expected error for invalid output path")
	}
}

func TestWriteBatchRepairReport_Error(t *testing.T) {
	flags := &RootFlags{Output: "/invalid/path", Format: "text"}
	opts, err := NewReportOptions(flags)
	if err != nil {
		t.Fatalf("NewReportOptions failed: %v", err)
	}
	result := &operations.BatchResult{}
	err = WriteBatchRepairReport(result, opts)
	if err == nil {
		t.Error("Expected error for invalid output path")
	}
}

func TestWriteBatchValidationReport(t *testing.T) {
	flags := &RootFlags{Format: "text"}
	opts, _ := NewReportOptions(flags)
	result := &operations.BatchResult{Total: 0}
	err := WriteBatchValidationReport(result, opts)
	if err != nil {
		t.Errorf("WriteBatchValidationReport failed: %v", err)
	}
}

func TestWriteBatchRepairReport(t *testing.T) {
	flags := &RootFlags{Format: "text"}
	opts, _ := NewReportOptions(flags)
	result := &operations.BatchResult{Total: 0}
	err := WriteBatchRepairReport(result, opts)
	if err != nil {
		t.Errorf("WriteBatchRepairReport failed: %v", err)
	}
}
