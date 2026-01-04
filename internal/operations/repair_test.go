package operations

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/project/pkg/ebmlib"
)

func TestNewRepairOperation(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	if op == nil {
		t.Fatal("Expected non-nil operation")
	}

	if op.ctx != ctx {
		t.Error("Expected context to be set")
	}
}

func TestRepairOperation_Preview_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	_, err := op.Preview("test.txt")

	if err == nil {
		t.Error("Expected error for unsupported file type")
	}

	expectedMsg := "unsupported file type"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestRepairOperation_Preview_MissingFile(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	_, err := op.Preview("nonexistent.epub")

	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestRepairOperation_Preview_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure context expires

	op := NewRepairOperation(ctx)

	_, err := op.Preview("test.epub")

	if err == nil {
		t.Error("Expected error due to timeout or missing file")
	}
}

func TestRepairOperation_Preview_EPUB_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.epub", []byte("invalid epub content"))

	preview, err := op.Preview(testFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("EPUB file should not be rejected as unsupported")
	}
	_ = preview
}

func TestRepairOperation_Preview_PDF_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.pdf", []byte("invalid pdf content"))

	preview, err := op.Preview(testFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("PDF file should not be rejected as unsupported")
	}
	_ = preview
}

func TestRepairOperation_Preview_CaseInsensitiveExtension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	tests := []struct {
		name     string
		filename string
	}{
		{"uppercase EPUB", "test.EPUB"},
		{"mixed case EPUB", "test.ePub"},
		{"uppercase PDF", "test.PDF"},
		{"mixed case PDF", "test.Pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := createTestFile(t, tt.filename, []byte("content"))

			_, err := op.Preview(testFile)

			// Should not fail with "unsupported file type"
			if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
				t.Errorf("File '%s' should not be rejected as unsupported", tt.filename)
			}
		})
	}
}

func TestRepairOperation_Execute_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	_, err := op.Execute("test.txt")

	if err == nil {
		t.Error("Expected error for unsupported file type")
	}

	expectedMsg := "unsupported file type"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestRepairOperation_Execute_MissingFile(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	_, err := op.Execute("nonexistent.epub")

	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestRepairOperation_Execute_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure context expires

	op := NewRepairOperation(ctx)

	_, err := op.Execute("test.epub")

	if err == nil {
		t.Error("Expected error due to timeout or missing file")
	}
}

func TestRepairOperation_Execute_EPUB_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.epub", []byte("invalid epub content"))

	result, err := op.Execute(testFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("EPUB file should not be rejected as unsupported")
	}
	_ = result
}

func TestRepairOperation_Execute_PDF_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.pdf", []byte("invalid pdf content"))

	result, err := op.Execute(testFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("PDF file should not be rejected as unsupported")
	}
	_ = result
}

func TestRepairOperation_Execute_CaseInsensitiveExtension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	tests := []struct {
		name     string
		filename string
	}{
		{"uppercase EPUB", "test.EPUB"},
		{"mixed case EPUB", "test.ePub"},
		{"uppercase PDF", "test.PDF"},
		{"mixed case PDF", "test.Pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := createTestFile(t, tt.filename, []byte("content"))

			_, err := op.Execute(testFile)

			// Should not fail with "unsupported file type"
			if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
				t.Errorf("File '%s' should not be rejected as unsupported", tt.filename)
			}
		})
	}
}

func TestRepairOperation_ExecuteWithPreview_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	preview := &ebmlib.RepairPreview{}
	_, err := op.ExecuteWithPreview("test.txt", preview, "output.txt")

	if err == nil {
		t.Error("Expected error for unsupported file type")
	}

	expectedMsg := "unsupported file type"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestRepairOperation_ExecuteWithPreview_EPUB_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.epub", []byte("invalid epub content"))
	outputFile := filepath.Join(t.TempDir(), "output.epub")

	preview := &ebmlib.RepairPreview{}
	result, err := op.ExecuteWithPreview(testFile, preview, outputFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("EPUB file should not be rejected as unsupported")
	}
	_ = result
}

func TestRepairOperation_ExecuteWithPreview_PDF_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	testFile := createTestFile(t, "test.pdf", []byte("invalid pdf content"))
	outputFile := filepath.Join(t.TempDir(), "output.pdf")

	preview := &ebmlib.RepairPreview{}
	result, err := op.ExecuteWithPreview(testFile, preview, outputFile)

	// Should not fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("PDF file should not be rejected as unsupported")
	}
	_ = result
}

func TestRepairOperation_ExecuteWithPreview_CaseInsensitiveExtension(t *testing.T) {
	ctx := context.Background()
	op := NewRepairOperation(ctx)

	tests := []struct {
		name     string
		filename string
		output   string
	}{
		{"uppercase EPUB", "test.EPUB", "output.EPUB"},
		{"mixed case EPUB", "test.ePub", "output.ePub"},
		{"uppercase PDF", "test.PDF", "output.PDF"},
		{"mixed case PDF", "test.Pdf", "output.Pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := createTestFile(t, tt.filename, []byte("content"))
			outputFile := filepath.Join(t.TempDir(), tt.output)

			preview := &ebmlib.RepairPreview{}
			_, err := op.ExecuteWithPreview(testFile, preview, outputFile)

			// Should not fail with "unsupported file type"
			if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
				t.Errorf("File '%s' should not be rejected as unsupported", tt.filename)
			}
		})
	}
}

func TestRepairResult(t *testing.T) {
	// Test that RepairResult type exists and can be created
	result := RepairResult{
		FilePath: "test.epub",
		Result:   nil,
		Error:    nil,
	}

	if result.FilePath != "test.epub" {
		t.Errorf("Expected FilePath 'test.epub', got '%s'", result.FilePath)
	}
}

func TestRepairResult_WithError(t *testing.T) {
	result := RepairResult{
		FilePath: "test.epub",
		Result:   nil,
		Error:    context.DeadlineExceeded,
	}

	if result.Error != context.DeadlineExceeded {
		t.Error("Expected error to be set")
	}

	if result.Result != nil {
		t.Error("Expected Result to be nil when error is set")
	}
}

func TestRepairResult_WithResult(t *testing.T) {
	repairResult := &ebmlib.RepairResult{
		Success: true,
	}

	result := RepairResult{
		FilePath: "test.epub",
		Result:   repairResult,
		Error:    nil,
	}

	if result.Result == nil {
		t.Error("Expected Result to be set")
	}

	if result.Error != nil {
		t.Error("Expected Error to be nil when Result is set")
	}

	if result.Result != repairResult {
		t.Error("Expected Result to match the provided repair result")
	}
}
