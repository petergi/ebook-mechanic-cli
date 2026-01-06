package operations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// RepairOperation handles ebook repair
type RepairOperation struct {
	ctx context.Context
}

// RepairSaveMode controls how repaired files are saved.
type RepairSaveMode string

const (
	RepairSaveModeBackupOriginal RepairSaveMode = "backup-original"
	RepairSaveModeRenameRepaired RepairSaveMode = "rename-repaired"
)

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
	ext, err := repairExtension(filePath)
	if err != nil {
		return nil, err
	}

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
	ext, err := repairExtension(filePath)
	if err != nil {
		return nil, err
	}

	switch ext {
	case ".epub":
		return ebmlib.RepairEPUBWithPreviewContext(r.ctx, filePath, preview, outputPath)
	case ".pdf":
		return ebmlib.RepairPDFWithPreviewContext(r.ctx, filePath, preview, outputPath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// ExecuteWithSaveMode performs a repair and applies the requested save mode.
func (r *RepairOperation) ExecuteWithSaveMode(filePath string, mode RepairSaveMode, backupDir string) (*ebmlib.RepairResult, string, error) {
	if _, err := repairExtension(filePath); err != nil {
		return nil, "", err
	}

	switch mode {
	case RepairSaveModeRenameRepaired:
		if backupDir != "" {
			return nil, "", fmt.Errorf("backup dir is not supported with rename-repaired mode")
		}
		repairedPath := withSuffix(filePath, "_repaired", "")
		if err := copyFile(filePath, repairedPath); err != nil {
			return nil, "", fmt.Errorf("failed to copy file for repair: %w", err)
		}
		result, err := r.Execute(repairedPath)
		if err != nil {
			return nil, "", err
		}
		result.BackupPath = repairedPath
		return result, repairedPath, nil

	case RepairSaveModeBackupOriginal:
		backupPath := withSuffix(filePath, "_original", backupDir)
		if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
			return nil, "", fmt.Errorf("failed to create backup directory: %w", err)
		}
		if err := copyFile(filePath, backupPath); err != nil {
			return nil, "", fmt.Errorf("failed to create backup: %w", err)
		}
		result, err := r.Execute(filePath)
		if err != nil {
			return nil, "", err
		}
		result.BackupPath = backupPath
		return result, filePath, nil

	default:
		return nil, "", fmt.Errorf("unsupported save mode: %s", mode)
	}
}

// RepairResult contains the result of a repair operation
type RepairResult struct {
	FilePath string
	Result   *ebmlib.RepairResult
	Error    error
}

func repairExtension(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub", ".pdf":
		return ext, nil
	default:
		return ext, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

func withSuffix(filePath, suffix, dirOverride string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	dir := filepath.Dir(filePath)
	if dirOverride != "" {
		dir = dirOverride
	}
	return filepath.Join(dir, fmt.Sprintf("%s%s%s", name, suffix, ext))
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
