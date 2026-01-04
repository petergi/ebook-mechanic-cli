package operations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// RepairOperation handles ebook repair
type RepairOperation struct {
	ctx context.Context
}

// NewRepairOperation creates a new repair operation
func NewRepairOperation(ctx context.Context) *RepairOperation {
	return &RepairOperation{ctx: ctx}
}

// Preview generates a repair preview for the given file
func (r *RepairOperation) Preview(filePath string) (*ebmlib.RepairPreview, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ebmlib.PreviewEPUBRepairWithContext(r.ctx, filePath)
	case ".pdf":
		return ebmlib.PreviewPDFRepairWithContext(r.ctx, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// Execute performs repair on the given file
func (r *RepairOperation) Execute(filePath string) (*ebmlib.RepairResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ebmlib.RepairEPUBWithContext(r.ctx, filePath)
	case ".pdf":
		return ebmlib.RepairPDFWithContext(r.ctx, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// ExecuteWithPreview performs repair using a pre-generated preview
func (r *RepairOperation) ExecuteWithPreview(filePath string, preview *ebmlib.RepairPreview, outputPath string) (*ebmlib.RepairResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ebmlib.RepairEPUBWithPreviewContext(r.ctx, filePath, preview, outputPath)
	case ".pdf":
		return ebmlib.RepairPDFWithPreviewContext(r.ctx, filePath, preview, outputPath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// RepairResult contains the result of a repair operation
type RepairResult struct {
	FilePath string
	Result   *ebmlib.RepairResult
	Error    error
}
