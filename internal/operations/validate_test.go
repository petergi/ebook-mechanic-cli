package operations

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewValidateOperation(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

	if op == nil {
		t.Fatal("Expected non-nil operation")
	}

	if op.ctx != ctx {
		t.Error("Expected context to be set")
	}
}

func TestValidateOperation_Execute_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

	_, err := op.Execute("test.txt")

	if err == nil {
		t.Error("Expected error for unsupported file type")
	}

	expectedMsg := "unsupported file type"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestValidateOperation_Execute_MissingFile(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

	_, err := op.Execute("nonexistent.epub")

	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestValidateOperation_Execute_WithTimeout(t *testing.T) {
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure context expires

	op := NewValidateOperation(ctx)

	// This should fail due to context timeout
	// Note: actual behavior depends on ebmlib implementation
	_, err := op.Execute("test.epub")

	// We expect an error (either file not found or context deadline)
	if err == nil {
		t.Error("Expected error due to timeout or missing file")
	}
}

func TestValidateResult(t *testing.T) {
	// Test that ValidateResult type exists and can be created
	result := ValidateResult{
		FilePath: "test.epub",
		Report:   nil,
		Error:    nil,
	}

	if result.FilePath != "test.epub" {
		t.Errorf("Expected FilePath 'test.epub', got '%s'", result.FilePath)
	}
}

// Helper function to create a test file
func createTestFile(t *testing.T, name string, content []byte) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return filePath
}

func TestValidateOperation_Execute_EPUB_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

	// Create a test file (even if invalid, we're testing the path routing)
	testFile := createTestFile(t, "test.epub", []byte("invalid epub content"))

	// This will fail validation, but we're testing that it reaches the EPUB validator
	report, err := op.Execute(testFile)

	// We expect either an error or a report (behavior depends on ebmlib)
	// The important thing is that it didn't fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("EPUB file should not be rejected as unsupported")
	}

	// If we got a report, it should be for the correct file
	if report != nil && report.FilePath != testFile {
		t.Errorf("Expected report for '%s', got '%s'", testFile, report.FilePath)
	}
}

func TestValidateOperation_Execute_PDF_Extension(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

	// Create a test file
	testFile := createTestFile(t, "test.pdf", []byte("invalid pdf content"))

	// This will fail validation, but we're testing that it reaches the PDF validator
	report, err := op.Execute(testFile)

	// We expect either an error or a report
	// The important thing is that it didn't fail with "unsupported file type"
	if err != nil && err.Error()[:len("unsupported file type")] == "unsupported file type" {
		t.Error("PDF file should not be rejected as unsupported")
	}

	// If we got a report, it should be for the correct file
	if report != nil && report.FilePath != testFile {
		t.Errorf("Expected report for '%s', got '%s'", testFile, report.FilePath)
	}
}

func TestValidateOperation_Execute_CaseInsensitiveExtension(t *testing.T) {
	ctx := context.Background()
	op := NewValidateOperation(ctx)

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
