package operations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/example/project/pkg/ebmlib"
)

// ValidateOperation handles ebook validation
type ValidateOperation struct {
	ctx context.Context
}

// NewValidateOperation creates a new validation operation
func NewValidateOperation(ctx context.Context) *ValidateOperation {
	return &ValidateOperation{ctx: ctx}
}

// Execute performs validation on the given file
func (v *ValidateOperation) Execute(filePath string) (*ebmlib.ValidationReport, error) {
	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ebmlib.ValidateEPUBWithContext(v.ctx, filePath)
	case ".pdf":
		return ebmlib.ValidatePDFWithContext(v.ctx, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// ValidateResult contains the result of a validation operation
type ValidateResult struct {
	FilePath string
	Report   *ebmlib.ValidationReport
	Error    error
}
