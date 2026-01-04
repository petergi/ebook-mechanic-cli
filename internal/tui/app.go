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
	StateProgress
	StateReport
)

// App is the main TUI application coordinator
type App struct {
	state         AppState
	menuModel     models.MenuModel
	browserModel  models.BrowserModel
	progressModel models.ProgressModel
	reportModel   models.ReportModel
	ctx           context.Context
	cancel        context.CancelFunc
	selectedFile  string
}

// NewApp creates a new TUI application
func NewApp() App {
	ctx, cancel := context.WithCancel(context.Background())

	return App{
		state:     StateMenu,
		menuModel: models.NewMenuModel(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Init initializes the application
func (a App) Init() tea.Cmd {
	return a.menuModel.Init()
}

// Update handles messages and updates the application state
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.state {
	case StateMenu:
		return a.updateMenu(msg)
	case StateBrowser:
		return a.updateBrowser(msg)
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
		switch msg.Action {
		case "validate", "repair":
			// Show file browser
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBrowserModel(cwd)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "batch":
			// Show directory browser for batch
			cwd, _ := os.Getwd()
			a.browserModel = models.NewBatchBrowserModel(cwd)
			a.state = StateBrowser
			return a, a.browserModel.Init()

		case "quit":
			return a, tea.Quit
		}

	default:
		var m tea.Model
		m, cmd = a.menuModel.Update(msg)
		a.menuModel = m.(models.MenuModel)
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
		menuAction := a.menuModel.SelectedAction()

		switch menuAction {
		case "validate":
			return a.startValidation(msg.Path)
		case "repair":
			return a.startRepair(msg.Path)
		case "batch":
			return a.startBatch(msg.Path)
		}
	case models.DirectorySelectMsg:
		menuAction := a.menuModel.SelectedAction()
		if menuAction == "batch" {
			return a.startBatch(msg.Path)
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
			a.reportModel = models.NewReportModel(result)
			a.state = StateReport
			return a, a.reportModel.Init()

		case *ebmlib.RepairResult:
			a.reportModel = models.NewRepairReportModel(result)
			a.state = StateReport
			return a, a.reportModel.Init()

		case operations.BatchResult:
			var m tea.Model
			m, cmd = a.progressModel.Update(msg)
			a.progressModel = m.(models.ProgressModel)
			return a, cmd
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

func (a App) startBatch(path string) (tea.Model, tea.Cmd) {
	files, err := collectBatchFiles(path)
	if err != nil {
		a.progressModel = models.NewProgressModel("Batch Processing", path, 0)
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

	a.progressModel = models.NewProgressModel("Batch Processing", path, len(files))
	a.state = StateProgress

	return a, tea.Batch(
		a.progressModel.Init(),
		func() tea.Msg {
			start := time.Now()
			batch := operations.NewBatchProcessor(operations.DefaultBatchConfig())
			results := batch.Execute(files, operations.OperationValidate)
			return models.OperationDoneMsg{
				Result: operations.AggregateResults(results, time.Since(start)),
			}
		},
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
	a.progressModel = models.NewProgressModel("Validating", filePath, 1)
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
	a.progressModel = models.NewProgressModel("Repairing", filePath, 1)
	a.state = StateProgress

	// Start repair in background
	return a, tea.Batch(
		a.progressModel.Init(),
		func() tea.Msg {
			repairer := operations.NewRepairOperation(a.ctx)
			result, err := repairer.Execute(filePath)

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
