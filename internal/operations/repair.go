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
	ctx        context.Context
	aggressive bool
}

// RepairSaveMode controls how repaired files are saved.
type RepairSaveMode string

const (
	RepairSaveModeBackupOriginal RepairSaveMode = "backup-original"
	RepairSaveModeNoBackup       RepairSaveMode = "no-backup"
)

// NewRepairOperation creates a new repair operation
func NewRepairOperation(ctx context.Context) *RepairOperation {
	return &RepairOperation{ctx: ctx}
}

// WithAggressive enables aggressive repair mode for the operation.
func (r *RepairOperation) WithAggressive(enabled bool) *RepairOperation {
	r.aggressive = enabled
	return r
}

// Preview generates a repair preview for the given file
func (r *RepairOperation) Preview(filePath string) (*ebmlib.RepairPreview, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ebmlib.PreviewEPUBRepairWithOptions(r.ctx, filePath, ebmlib.RepairOptions{Aggressive: r.aggressive})
	case ".pdf":
		return ebmlib.PreviewPDFRepairWithContext(r.ctx, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (expected .epub or .pdf)", ext)
	}
}

// Execute performs repair on the given file
func (r *RepairOperation) Execute(filePath string) (*ebmlib.RepairResult, error) {
	result, _, err := r.ExecuteWithSaveMode(filePath, RepairSaveModeBackupOriginal, "")
	return result, err
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
	case RepairSaveModeBackupOriginal:
		result, outputPath, err := r.executeInPlaceWithBackup(filePath, backupDir)
		return result, outputPath, err

	case RepairSaveModeNoBackup:
		if backupDir != "" {
			return nil, "", fmt.Errorf("backup dir is not supported with no-backup mode")
		}
		result, outputPath, err := r.executeInPlaceNoBackup(filePath)
		return result, outputPath, err

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

func (r *RepairOperation) executeInPlaceWithBackup(filePath, backupDir string) (*ebmlib.RepairResult, string, error) {
	preview, err := r.Preview(filePath)
	if err != nil {
		return nil, "", err
	}
	if len(preview.Actions) == 0 {
		return &ebmlib.RepairResult{Success: true, ActionsApplied: []ebmlib.RepairAction{}}, filePath, nil
	}

	tmpPath, err := tempRepairPath(filePath)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	result, err := r.ExecuteWithPreview(filePath, preview, tmpPath)
	if err != nil {
		return nil, "", err
	}
	if !result.Success {
		result.BackupPath = ""
		return result, filePath, nil
	}

	backupPath := withSuffix(filePath, "_original", backupDir)
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	if err := copyFile(filePath, backupPath); err != nil {
		return nil, "", fmt.Errorf("failed to create backup: %w", err)
	}

	if err := replaceFile(tmpPath, filePath); err != nil {
		return nil, "", err
	}

	result.BackupPath = backupPath
	return result, filePath, nil
}

func (r *RepairOperation) executeInPlaceNoBackup(filePath string) (*ebmlib.RepairResult, string, error) {
	preview, err := r.Preview(filePath)
	if err != nil {
		return nil, "", err
	}
	if len(preview.Actions) == 0 {
		return &ebmlib.RepairResult{Success: true, ActionsApplied: []ebmlib.RepairAction{}}, filePath, nil
	}

	tmpPath, err := tempRepairPath(filePath)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	result, err := r.ExecuteWithPreview(filePath, preview, tmpPath)
	if err != nil {
		return nil, "", err
	}
	if !result.Success {
		result.BackupPath = ""
		return result, filePath, nil
	}

	if err := replaceFile(tmpPath, filePath); err != nil {
		return nil, "", err
	}

	result.BackupPath = ""
	return result, filePath, nil
}

func tempRepairPath(filePath string) (string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	pattern := fmt.Sprintf("%s.repairing-*%s", name, ext)

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(tmpPath); err != nil {
		return "", err
	}

	return tmpPath, nil
}

func replaceFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to replace file: %w", err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
