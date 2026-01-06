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
	StateRepairMode
	StateProgress
	StateReport
)

// App is the main TUI application coordinator
type App struct {
	state         AppState
	menuModel     models.MenuModel
	browserModel  models.BrowserModel
	repairModel   models.RepairModeModel
	progressModel models.ProgressModel
	reportModel   models.ReportModel
	ctx           context.Context
	cancel        context.CancelFunc
	selectedFile  string
	activeAction  string
	repairMode    operations.RepairSaveMode
	width         int
	height        int
	progressCh    <-chan operations.ProgressUpdate
}

// NewApp creates a new TUI application
func NewApp() App {
	ctx, cancel := context.WithCancel(context.Background())

	return App{
		state:      StateMenu,
		menuModel:  models.NewMenuModel(),
		ctx:        ctx,
		cancel:     cancel,
		repairMode: operations.RepairSaveModeBackupOriginal,
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
	case StateRepairMode:
		return a.updateRepairMode(msg)
	case StateProgress:
		return a.updateProgress(msg)
	case StateReport:
		return a.updateReport(msg)
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
	case StateRepairMode:
		return a.repairModel.View()
	case StateProgress:
		return a.progressModel.View()
	case StateReport:
		return a.reportModel.View()
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

		case "repair", "multi-repair", "batch-repair":
			a.repairModel = models.NewRepairModeModel(a.width, a.height)
			a.state = StateRepairMode
			return a, a.repairModel.Init()

		case "multi-validate":
			// Show multi-select browser
			cwd, _ := os.Getwd()
			a.browserModel = models.NewMultiSelectBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "batch-validate":
			// Show directory browser for batch
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBatchBrowserModel(cwd, a.width, a.height)
			a.state = StateBrowser
			return a, a.browserModel.Init()

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

// updateRepairMode handles repair save-mode selection
func (a App) updateRepairMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case models.RepairModeSelectMsg:
		a.repairMode = operations.RepairSaveMode(msg.Mode)
		cwd, _ := os.Getwd()
		switch a.activeAction {
		case "repair":
			a.browserModel = models.NewBrowserModel(cwd, a.width, a.height)
		case "multi-repair":
			a.browserModel = models.NewMultiSelectBrowserModel(cwd, a.width, a.height)
		case "batch-repair":
			a.browserModel = models.NewBatchBrowserModel(cwd, a.width, a.height)
		default:
			a.state = StateMenu
			return a, nil
		}
		a.state = StateBrowser
		return a, a.browserModel.Init()

	case models.BackToMenuMsg:
		a.state = StateMenu
		return a, nil

	default:
		var m tea.Model
		m, cmd = a.repairModel.Update(msg)
		a.repairModel = m.(models.RepairModeModel)
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
			return a.startRepair(msg.Path, a.repairMode)
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
	files, err := collectBatchFiles(path)
	if err != nil {
		title := "Batch Validation"
		if opType == operations.OperationRepair {
			title = "Batch Repair"
		}
		a.progressModel = models.NewProgressModel(title, path, 0, a.width, a.height)
		a.state = StateProgress
		return a, tea.Batch(
			a.progressModel.Init(),
			func() tea.Msg {
				return models.OperationDoneMsg{
					Result: operations.BatchResult{
						Failed: []operations.Result{
							{FilePath: path, Error: fmt.Errorf("batch scan failed: %w", err)},
						},
						Total: 1,
					},
				}
			},
		)
	}

	title := "Batch Validation"
	if opType == operations.OperationRepair {
		title = "Batch Repair"
	}
	a.progressModel = models.NewProgressModel(title, path, len(files), a.width, a.height)
	a.state = StateProgress

	// Start batch processing in a goroutine to not block the UI
	// and send the result when done.
	config := operations.DefaultBatchConfig()
	config.RepairMode = a.repairMode
	batch := operations.NewBatchProcessor(a.ctx, config)
	doneCh := make(chan operations.BatchResult)
	start := time.Now()

	go func() {
		results := batch.Execute(files, opType)
		doneCh <- operations.AggregateResults(results, time.Since(start), opType)
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
	config.RepairMode = a.repairMode
	batch := operations.NewBatchProcessor(a.ctx, config)
	doneCh := make(chan operations.BatchResult)
	start := time.Now()

	go func() {
		results := batch.Execute(files, opType)
		doneCh <- operations.AggregateResults(results, time.Since(start), opType)
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

func collectBatchFiles(path string) ([]string, error) {
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
func (a App) startRepair(filePath string, mode operations.RepairSaveMode) (tea.Model, tea.Cmd) {
	a.progressModel = models.NewProgressModel("Repairing", filePath, 1, a.width, a.height)
	a.state = StateProgress

	// Start repair in background
	return a, tea.Batch(
		a.progressModel.Init(),
		func() tea.Msg {
			repairer := operations.NewRepairOperation(a.ctx)
			result, _, err := repairer.ExecuteWithSaveMode(filePath, mode, "")

			if err != nil {
				// Create error result
				result = &ebmlib.RepairResult{
					Success: false,
					Error:   err,
				}
			}

			return models.OperationDoneMsg{Result: result}
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
