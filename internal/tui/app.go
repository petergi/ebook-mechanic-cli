package tui

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/models"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// AppState represents the current state of the application
type AppState int

const (
	StateMenu AppState = iota
	StateBrowser
	StateSettings
	StateProgress
	StateReport
	StateCalibre
)

// App is the main TUI application coordinator
type App struct {
	state              AppState
	menuModel          models.MenuModel
	browserModel       models.BrowserModel
	settingsModel      models.SettingsModel
	progressModel      models.ProgressModel
	reportModel        models.ReportModel
	calibreModel       models.CalibreModel
	ctx                context.Context
	cancel             context.CancelFunc
	selectedFile       string
	activeAction       string
	batchJobs          int
	skipValidation     bool
	noBackup           bool
	aggressive         bool
	removeSystemErrors bool // Remove books with system errors
	moveFailedRepairs  bool // Move unrepairable books to INVALID folder
	cleanupEmptyDirs   bool // Clean up empty parent directories after removal/move
	width              int
	height             int
	progressCh         <-chan operations.ProgressUpdate
}

// NewApp creates a new TUI application
func NewApp() App {
	ctx, cancel := context.WithCancel(context.Background())

	return App{
		state:              StateMenu,
		menuModel:          models.NewMenuModel(),
		ctx:                ctx,
		cancel:             cancel,
		batchJobs:          operations.DefaultBatchConfig().NumWorkers,
		skipValidation:     false,
		noBackup:           false,
		aggressive:         false,
		removeSystemErrors: false,
		moveFailedRepairs:  false,
		cleanupEmptyDirs:   true,
	}
}

// Init initializes the application
func (a App) Init() tea.Cmd {
	return a.menuModel.Init()
}

// Update handles messages and updates the application state
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size updates for all states
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		a.width = msg.Width
		a.height = msg.Height
	}

	switch a.state {
	case StateMenu:
		return a.updateMenu(msg)
	case StateBrowser:
		return a.updateBrowser(msg)
	case StateSettings:
		return a.updateSettings(msg)
	case StateProgress:
		return a.updateProgress(msg)
	case StateReport:
		return a.updateReport(msg)
	case StateCalibre:
		return a.updateCalibre(msg)
	}

	return a, nil
}

// View renders the current view based on state
func (a App) View() string {
	switch a.state {
	case StateMenu:
		return a.menuModel.View()
	case StateBrowser:
		return a.browserModel.View()
	case StateSettings:
		return a.settingsModel.View()
	case StateProgress:
		return a.progressModel.View()
	case StateReport:
		return a.reportModel.View()
	case StateCalibre:
		return a.calibreModel.View()
	}

	return "Unknown state"
}

// updateMenu handles menu state updates
func (a App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.MenuSelectMsg:
		a.activeAction = msg.Action
		switch msg.Action {
		case "validate":
			// Show file browser
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "repair", "multi-repair":
			// Show file browser
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "multi-validate":
			// Show multi-select browser
			cwd, _ := os.Getwd()
			a.browserModel = models.NewMultiSelectBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "batch-validate", "batch-repair":
			// Show directory browser for batch
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBatchBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "calibre-cleanup":
			// Show directory browser for Calibre library
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBatchBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "settings":
			a.settingsModel = models.NewSettingsModel(a.batchJobs, a.skipValidation, a.noBackup, a.aggressive, a.removeSystemErrors, a.moveFailedRepairs, a.cleanupEmptyDirs, a.width, a.height)
			a.state = StateSettings
			return a, a.settingsModel.Init()

		case "quit":
			a.cancel()
			return a, tea.Quit
		}

	default:
		var m tea.Model
		m, cmd = a.menuModel.Update(msg)
		a.menuModel = m.(models.MenuModel)
		if m.(models.MenuModel).Quitting() {
			a.cancel()
			return a, tea.Quit
		}
	}

	return a, cmd
}

// updateSettings handles settings updates
func (a App) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.SettingsSaveMsg:
		a.batchJobs = msg.Jobs
		a.skipValidation = msg.SkipValidation
		a.noBackup = msg.NoBackup
		a.aggressive = msg.Aggressive
		a.removeSystemErrors = msg.RemoveSystemErrors
		a.moveFailedRepairs = msg.MoveFailedRepairs
		a.cleanupEmptyDirs = msg.CleanupEmptyDirs
		a.state = StateMenu
		return a, nil

	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	default:
		var m tea.Model
		m, cmd = a.settingsModel.Update(msg)
		a.settingsModel = m.(models.SettingsModel)
	}

	return a, cmd
}

// updateBrowser handles browser state updates
func (a App) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.FileSelectMsg:
		// File selected, start operation
		a.selectedFile = msg.Path
		menuAction := a.activeAction

		switch menuAction {
		case "validate":
			return a.startValidation(msg.Path)
		case "repair":
			return a.startRepair(msg.Path)
		case "batch-validate":
			return a.startBatchDirectory(msg.Path, operations.OperationValidate)
		case "batch-repair":
			return a.startBatchDirectory(msg.Path, operations.OperationRepair)
		}

	case models.MultiFileSelectMsg:
		// Multiple files selected, start batch operation
		menuAction := a.activeAction
		switch menuAction {
		case "validate", "repair", "batch-validate", "batch-repair", "multi-validate", "multi-repair":
			return a.startBatchWithFiles(msg.Paths, menuAction)
		}

	case models.DirectorySelectMsg:
		menuAction := a.activeAction
		switch menuAction {
		case "batch-validate":
			return a.startBatchDirectory(msg.Path, operations.OperationValidate)
		case "batch-repair":
			return a.startBatchDirectory(msg.Path, operations.OperationRepair)
		case "calibre-cleanup":
			return a.startCalibreCleanup(msg.Path)
		}

	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	default:
		var m tea.Model
		m, cmd = a.browserModel.Update(msg)
		a.browserModel = m.(models.BrowserModel)
	}

	return a, cmd
}

// updateProgress handles progress state updates
func (a App) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.OperationDoneMsg:
		// Operation complete, show report
		switch result := msg.Result.(type) {
		case *ebmlib.ValidationReport:
			a.reportModel = models.NewReportModel(result, a.width, a.height)
			a.state = StateReport
			return a, a.reportModel.Init()

		case models.RepairOutcome:
			a.reportModel = models.NewRepairReportModelWithValidation(result.Result, result.Report, a.width, a.height)
			a.state = StateReport
			return a, a.reportModel.Init()

		case *ebmlib.RepairResult:
			a.reportModel = models.NewRepairReportModel(result, a.width, a.height)
			a.state = StateReport
			return a, a.reportModel.Init()

		case operations.BatchResult:
			var m tea.Model
			m, cmd = a.progressModel.Update(msg)
			a.progressModel = m.(models.ProgressModel)
			return a, cmd
		}

	case models.ProgressUpdateMsg:
		var m tea.Model
		m, cmd = a.progressModel.Update(msg)
		a.progressModel = m.(models.ProgressModel)
		// Continue streaming if we have a channel
		if a.progressCh != nil {
			return a, tea.Batch(cmd, batchProgressCmd(a.progressCh))
		}
		return a, cmd

	case models.ViewReportMsg:
		switch result := msg.Result.(type) {
		case operations.BatchResult:
			a.reportModel = models.NewBatchReportModel(&result, a.width, a.height)
			a.state = StateReport
			return a, a.reportModel.Init()
		}

	case models.OperationCancelMsg:
		a.cancel()
		a.state = StateMenu
		return a, nil

	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	default:
		var m tea.Model
		m, cmd = a.progressModel.Update(msg)
		a.progressModel = m.(models.ProgressModel)
	}

	return a, cmd
}

func (a App) startBatchDirectory(path string, opType operations.OperationType) (tea.Model, tea.Cmd) {
	title := "Batch Validation"
	if opType == operations.OperationRepair {
		title = "Batch Repair"
	}

	// Show progress immediately so directory scanning doesn’t feel frozen
	a.progressModel = models.NewProgressModel(title, path, 0, a.width, a.height)
	a.state = StateProgress

	progressCh := make(chan operations.ProgressUpdate)
	a.progressCh = progressCh
	doneCh := make(chan operations.BatchResult)

	go func() {
		// Emit initial scanning status
		progressCh <- operations.ProgressUpdate{Completed: 0, Total: 0, Current: "Scanning library..."}

		files, err := collectBatchFiles(path, func(count int, sample string) {
			if count%200 == 0 {
				progressCh <- operations.ProgressUpdate{Completed: count, Total: 0, Current: fmt.Sprintf("Scanning library... (%d)", count)}
			}
		})
		if err != nil {
			doneCh <- operations.BatchResult{
				Failed: []operations.Result{{FilePath: path, Error: fmt.Errorf("batch scan failed: %w", err)}},
				Total:  1,
			}
			close(progressCh)
			return
		}

		// Announce total so the progress bar activates when batch starts
		progressCh <- operations.ProgressUpdate{Completed: 0, Total: len(files), Current: fmt.Sprintf("Found %d files", len(files))}

		config := operations.DefaultBatchConfig()
		if a.noBackup {
			config.RepairMode = operations.RepairSaveModeNoBackup
		} else {
			config.RepairMode = operations.RepairSaveModeBackupOriginal
		}
		config.Aggressive = a.aggressive
		if a.batchJobs > 0 {
			config.NumWorkers = a.batchJobs
		}

		batch := operations.NewBatchProcessor(a.ctx, config)
		batchPath := path
		batchStart := time.Now()

		// Forward batch progress into our unified channel so the UI keeps streaming
		go func() {
			for update := range batch.ProgressChannel() {
				progressCh <- update
			}
		}()

		results := batch.Execute(files, opType)
		aggregated := operations.AggregateResults(results, time.Since(batchStart), opType)

		// Record the options used for this batch
		aggregated.Options = operations.BatchOptions{
			NumWorkers:         config.NumWorkers,
			SkipValidation:     a.skipValidation,
			NoBackup:           a.noBackup,
			Aggressive:         a.aggressive,
			RemoveSystemErrors: a.removeSystemErrors,
			MoveFailedRepairs:  a.moveFailedRepairs,
			CleanupEmptyDirs:   a.cleanupEmptyDirs,
		}

		// Capture intended cleanup lists (non-blocking), then perform cleanup async
		if a.removeSystemErrors && len(aggregated.Errored) > 0 {
			aggregated.RemovedFiles = make([]string, 0, len(aggregated.Errored))
			for _, r := range aggregated.Errored {
				aggregated.RemovedFiles = append(aggregated.RemovedFiles, r.FilePath)
			}
		}
		if a.moveFailedRepairs && opType == operations.OperationRepair && len(aggregated.Invalid) > 0 {
			aggregated.MovedFiles = make([]string, 0, len(aggregated.Invalid))
			for _, r := range aggregated.Invalid {
				aggregated.MovedFiles = append(aggregated.MovedFiles, r.FilePath)
			}
		}

		doneCh <- aggregated
		close(progressCh)

		// Run filesystem cleanup asynchronously so the UI/progress is not blocked
		go func(erred []operations.Result, invalid []operations.Result) {
			// Remove errored files
			if a.removeSystemErrors {
				for _, r := range erred {
					_ = os.Remove(r.FilePath)
					if a.cleanupEmptyDirs {
						removeEmptyParentDirs(filepath.Dir(r.FilePath), batchPath)
					}
				}
			}
			// Move invalid repairs
			if a.moveFailedRepairs && opType == operations.OperationRepair {
				invalidDir := filepath.Join(batchPath, "INVALID")
				_ = os.MkdirAll(invalidDir, 0755)
				for _, r := range invalid {
					dstPath := filepath.Join(invalidDir, filepath.Base(r.FilePath))
					_ = os.Rename(r.FilePath, dstPath)
					if a.cleanupEmptyDirs {
						removeEmptyParentDirs(filepath.Dir(r.FilePath), batchPath)
					}
				}
			}
		}(aggregated.Errored, aggregated.Invalid)
	}()

	return a, tea.Batch(
		a.progressModel.Init(),
		// Command to wait for result
		func() tea.Msg {
			result := <-doneCh
			return models.OperationDoneMsg{Result: result}
		},
		// Command to stream progress
		batchProgressCmd(progressCh),
	)
}

func (a App) startBatchWithFiles(files []string, action string) (tea.Model, tea.Cmd) {
	if len(files) == 0 {
		return a, nil
	}

	// Determine operation type
	var opType operations.OperationType
	title := "Batch Processing"
	switch action {
	case "validate", "multi-validate", "batch-validate":
		opType = operations.OperationValidate
		title = "Batch Validation"
	case "repair", "multi-repair", "batch-repair":
		opType = operations.OperationRepair
		title = "Batch Repair"
	default:
		opType = operations.OperationValidate
	}

	a.progressModel = models.NewProgressModel(title, fmt.Sprintf("%d files", len(files)), len(files), a.width, a.height)
	a.state = StateProgress

	// Start batch processing in a goroutine
	config := operations.DefaultBatchConfig()
	if a.noBackup {
		config.RepairMode = operations.RepairSaveModeNoBackup
	} else {
		config.RepairMode = operations.RepairSaveModeBackupOriginal
	}
	config.Aggressive = a.aggressive
	if a.batchJobs > 0 {
		config.NumWorkers = a.batchJobs
	}
	batch := operations.NewBatchProcessor(a.ctx, config)
	doneCh := make(chan operations.BatchResult)
	start := time.Now()
	batchPath := filepath.Dir(files[0]) // Get common parent directory

	go func() {
		results := batch.Execute(files, opType)
		aggregated := operations.AggregateResults(results, time.Since(start), opType)

		// Record the options used for this batch
		aggregated.Options = operations.BatchOptions{
			NumWorkers:         config.NumWorkers,
			SkipValidation:     a.skipValidation,
			NoBackup:           a.noBackup,
			Aggressive:         a.aggressive,
			RemoveSystemErrors: a.removeSystemErrors,
			MoveFailedRepairs:  a.moveFailedRepairs,
			CleanupEmptyDirs:   a.cleanupEmptyDirs,
		}

		// Capture intended cleanup lists (non-blocking), then perform cleanup async
		if a.removeSystemErrors && len(aggregated.Errored) > 0 {
			aggregated.RemovedFiles = make([]string, 0, len(aggregated.Errored))
			for _, r := range aggregated.Errored {
				aggregated.RemovedFiles = append(aggregated.RemovedFiles, r.FilePath)
			}
		}
		if a.moveFailedRepairs && opType == operations.OperationRepair && len(aggregated.Invalid) > 0 {
			aggregated.MovedFiles = make([]string, 0, len(aggregated.Invalid))
			for _, r := range aggregated.Invalid {
				aggregated.MovedFiles = append(aggregated.MovedFiles, r.FilePath)
			}
		}

		doneCh <- aggregated

		// Run filesystem cleanup asynchronously so the UI/progress is not blocked
		go func(erred []operations.Result, invalid []operations.Result) {
			// Remove errored files
			if a.removeSystemErrors {
				for _, r := range erred {
					_ = os.Remove(r.FilePath)
					if a.cleanupEmptyDirs {
						removeEmptyParentDirs(filepath.Dir(r.FilePath), batchPath)
					}
				}
			}
			// Move invalid repairs
			if a.moveFailedRepairs && opType == operations.OperationRepair {
				invalidDir := filepath.Join(batchPath, "INVALID")
				_ = os.MkdirAll(invalidDir, 0755)
				for _, r := range invalid {
					dstPath := filepath.Join(invalidDir, filepath.Base(r.FilePath))
					_ = os.Rename(r.FilePath, dstPath)
					if a.cleanupEmptyDirs {
						removeEmptyParentDirs(filepath.Dir(r.FilePath), batchPath)
					}
				}
			}
		}(aggregated.Errored, aggregated.Invalid)
	}()

	a.progressCh = batch.ProgressChannel()

	return a, tea.Batch(
		a.progressModel.Init(),
		// Command to wait for result
		func() tea.Msg {
			result := <-doneCh
			return models.OperationDoneMsg{Result: result}
		},
		// Command to stream progress
		batchProgressCmd(a.progressCh),
	)
}

func collectBatchFiles(path string, onFound func(count int, sample string)) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if isEbookFile(path) {
			return []string{path}, nil
		}
		return []string{}, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(entryPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() && entryPath != path {
				return filepath.SkipDir
			}
			if !entry.IsDir() {
				return nil
			}
		}

		if entry.IsDir() {
			return nil
		}

		if isEbookFile(entryPath) {
			files = append(files, entryPath)
			if onFound != nil {
				onFound(len(files), entryPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func isEbookFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".epub" || ext == ".pdf"
}

// updateReport handles report state updates
func (a App) updateReport(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.(type) {
	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	default:
		var m tea.Model
		m, cmd = a.reportModel.Update(msg)
		a.reportModel = m.(models.ReportModel)
	}

	return a, cmd
}

// updateCalibre handles Calibre cleanup state updates
func (a App) updateCalibre(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	case models.CalibreScanMsg:
		// Start the scan in background
		return a, func() tea.Msg {
			progressCh := make(chan models.CalibreScanProgressMsg, 100)

			// Start progress forwarding in a goroutine
			go func() {
				for progress := range progressCh {
					// Send progress updates to the tea program
					// Note: This is a simplified approach; in production you might want
					// to use tea.Program.Send() for proper message passing
					_ = progress // Progress is handled within ScanCalibreLibrary
				}
			}()

			result := models.ScanCalibreLibrary(msg.LibraryPath, progressCh)
			close(progressCh)
			return models.CalibreScanCompleteMsg{Result: result}
		}

	case models.CalibreCleanupMsg:
		// Perform cleanup
		return a, func() tea.Msg {
			dirs, files, err := models.CleanupCalibreLibrary(a.calibreModel.GetScanResult())
			return models.CalibreCleanupCompleteMsg{
				CleanedDirs:  dirs,
				CleanedFiles: files,
				Error:        err,
			}
		}

	default:
		var m tea.Model
		m, cmd = a.calibreModel.Update(msg)
		a.calibreModel = m.(models.CalibreModel)
	}

	return a, cmd
}

// startCalibreCleanup starts the Calibre library cleanup process
func (a App) startCalibreCleanup(libraryPath string) (tea.Model, tea.Cmd) {
	a.calibreModel = models.NewCalibreModel(libraryPath, a.width, a.height)
	a.state = StateCalibre
	return a, a.calibreModel.Init()
}

// startValidation starts a validation operation
func (a App) startValidation(filePath string) (tea.Model, tea.Cmd) {
	a.progressModel = models.NewProgressModel("Validating", filePath, 1, a.width, a.height)
	a.state = StateProgress

	// Start validation in background
	return a, tea.Batch(
		a.progressModel.Init(),
		func() tea.Msg {
			validator := operations.NewValidateOperation(a.ctx)
			report, err := validator.Execute(filePath)

			if err != nil {
				// Create error report
				report = &ebmlib.ValidationReport{
					FilePath: filePath,
					IsValid:  false,
					Errors: []ebmlib.ValidationError{
						{
							Code:     "SYSTEM_ERROR",
							Message:  err.Error(),
							Severity: ebmlib.SeverityError,
						},
					},
				}
			}

			return models.OperationDoneMsg{Result: report}
		},
	)
}

// startRepair starts a repair operation
func (a App) startRepair(filePath string) (tea.Model, tea.Cmd) {
	a.progressModel = models.NewProgressModel("Repairing", filePath, 1, a.width, a.height)
	a.state = StateProgress

	// Start repair in background
	return a, tea.Batch(
		a.progressModel.Init(),
		func() tea.Msg {
			repairer := operations.NewRepairOperation(a.ctx).WithAggressive(a.aggressive)
			mode := operations.RepairSaveModeBackupOriginal
			if a.noBackup {
				mode = operations.RepairSaveModeNoBackup
			}
			result, outputPath, err := repairer.ExecuteWithSaveMode(filePath, mode, "")

			if err != nil {
				// Create error result
				result = &ebmlib.RepairResult{
					Success: false,
					Error:   err,
				}
				return models.OperationDoneMsg{Result: models.RepairOutcome{Result: result}}
			}

			var validationReport *ebmlib.ValidationReport
			if !a.skipValidation && result.Success && outputPath != "" {
				validateOp := operations.NewValidateOperation(a.ctx)
				validationReport, err = validateOp.Execute(outputPath)
				if err != nil {
					validationReport = &ebmlib.ValidationReport{
						FilePath: outputPath,
						IsValid:  false,
						Errors: []ebmlib.ValidationError{
							{
								Code:     "SYSTEM_ERROR",
								Message:  err.Error(),
								Severity: ebmlib.SeverityError,
							},
						},
					}
				}
			}

			return models.OperationDoneMsg{Result: models.RepairOutcome{Result: result, Report: validationReport}}
		},
	)
}

// Run starts the TUI application
func Run() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}

func batchProgressCmd(ch <-chan operations.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return nil
		}
		return models.ConvertBatchProgress(update)
	}
}

// isCalibreMetadataDirectory checks if a directory contains only Calibre metadata files
// (no ebooks, only cover.jpg, metadata.opf, .DS_Store, etc.)
func isCalibreMetadataDirectory(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	// Empty directory counts as removable
	if len(entries) == 0 {
		return true
	}

	// Check for ebook files - if any exist, not metadata-only
	hasEbook := false
	hasMetadata := false

	for _, entry := range entries {
		name := strings.ToLower(entry.Name())

		// Skip hidden files and common system files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			// Nested directory means not a simple Calibre metadata folder
			return false
		}

		ext := filepath.Ext(name)
		if ext == ".epub" || ext == ".pdf" || ext == ".mobi" || ext == ".azw" || ext == ".azw3" {
			hasEbook = true
			break
		}

		// Check for Calibre metadata files
		if name == "cover.jpg" || name == "cover.jpeg" || name == "cover.png" ||
			name == "metadata.opf" {
			hasMetadata = true
		}
	}

	// If no ebooks and has metadata files, or just empty/system files, it's metadata-only
	return !hasEbook && (hasMetadata || len(entries) == 0)
}

// removeEmptyParentDirs removes empty parent directories and Calibre metadata-only directories up to the batch root
// Bails out if a parent has more than 3 siblings to avoid scanning large shallow hierarchies
func removeEmptyParentDirs(dir string, rootPath string) {
	// Ensure both paths are absolute and clean
	rootPath = filepath.Clean(rootPath)
	dir = filepath.Clean(dir)

	// Walk up the directory tree
	for dir != rootPath && strings.HasPrefix(dir, rootPath) {
		// Check sibling count before attempting removal
		parent := filepath.Dir(dir)
		if parent != rootPath { // Don't bail at root level
			entries, err := os.ReadDir(parent)
			if err == nil && len(entries) > 3 {
				// Too many siblings, skip this entire branch to avoid performance hit
				break
			}
		}

		// First try to remove if it's a Calibre metadata-only directory
		if isCalibreMetadataDirectory(dir) {
			// Remove all files in the directory first
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						_ = os.Remove(filepath.Join(dir, entry.Name()))
					}
				}
			}
		}

		// Try to remove the directory if empty
		err := os.Remove(dir)
		if err != nil {
			// Directory not empty or other error, stop
			break
		}

		// Move to parent
		dir = parent
	}
}
