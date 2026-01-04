package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/models"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

func TestNewApp_InitialState(t *testing.T) {
	app := NewApp()

	if app.state != StateMenu {
		t.Errorf("expected initial state to be StateMenu, got %v", app.state)
	}

	if app.ctx == nil {
		t.Fatal("expected context to be initialized")
	}

	if app.cancel == nil {
		t.Fatal("expected cancel function to be initialized")
	}
}

func TestAppInit(t *testing.T) {
	app := NewApp()

	if cmd := app.Init(); cmd != nil {
		t.Error("expected Init to return nil command")
	}
}

func TestAppView_UnknownState(t *testing.T) {
	app := NewApp()
	app.state = AppState(99)

	if app.View() != "Unknown state" {
		t.Error("expected unknown state view")
	}
}

func TestAppUpdate_UnknownState(t *testing.T) {
	app := NewApp()
	app.state = AppState(99)

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := model.(App)

	if updated.state != AppState(99) {
		t.Errorf("expected state to remain unknown, got %v", updated.state)
	}

	if cmd != nil {
		t.Error("expected no command for unknown state")
	}
}

func TestAppView_AllStates(t *testing.T) {
	app := NewApp()

	menuView := app.View()
	if !strings.Contains(menuView, "Ebook Mechanic") {
		t.Error("expected menu view to include title")
	}

	tempDir := t.TempDir()
	app.browserModel = models.NewBrowserModel(tempDir, 80, 24)
	app.state = StateBrowser
	if !strings.Contains(app.View(), "File Browser") {
		t.Error("expected browser view to include title")
	}

	app.progressModel = models.NewProgressModel("Validating", "file.epub", 1, 80, 24)
	app.state = StateProgress
	if !strings.Contains(app.View(), "Validating") {
		t.Error("expected progress view to include operation name")
	}

	app.reportModel = models.NewReportModel(&ebmlib.ValidationReport{FilePath: "file.epub"}, 80, 24)
	app.state = StateReport
	if !strings.Contains(app.View(), "Validation Report") {
		t.Error("expected report view to include title")
	}
}

func TestAppUpdateMenu_SelectActions(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"validate", "validate"},
		{"repair", "repair"},
		{"batch", "batch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			model, cmd := app.Update(models.MenuSelectMsg{Action: tt.action})
			updated := model.(App)

			if updated.state != StateBrowser {
				t.Errorf("expected state to be StateBrowser, got %v", updated.state)
			}

			if cmd != nil {
				t.Error("expected no command from browser init")
			}
		})
	}
}

func TestAppUpdateMenu_Quit(t *testing.T) {
	app := NewApp()
	_, cmd := app.Update(models.MenuSelectMsg{Action: "quit"})

	if cmd == nil {
		t.Fatal("expected quit command")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestAppUpdateMenu_DelegatesToMenuModel(t *testing.T) {
	app := NewApp()
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := model.(App)

	if updated.state != StateMenu {
		t.Errorf("expected state to remain StateMenu, got %v", updated.state)
	}

	if cmd != nil {
		t.Error("expected no command from menu update")
	}
}

func TestAppUpdateBrowser_BackToMenu(t *testing.T) {
	app := NewApp()
	app.state = StateBrowser

	model, _ := app.Update(models.BackToMenuMsg{})
	updated := model.(App)

	if updated.state != StateMenu {
		t.Errorf("expected state to be StateMenu, got %v", updated.state)
	}
}

func TestAppUpdateBrowser_DelegatesToBrowserModel(t *testing.T) {
	app := NewApp()
	app.state = StateBrowser
	app.browserModel = models.NewBrowserModel(t.TempDir(), 80, 24)

	model, cmd := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := model.(App)

	if updated.state != StateBrowser {
		t.Errorf("expected state to remain StateBrowser, got %v", updated.state)
	}

	if cmd != nil {
		t.Error("expected no command from browser update")
	}
}

func TestAppUpdateBrowser_StartValidation(t *testing.T) {
	app := NewApp()
	app.state = StateBrowser

	model, cmd := app.Update(models.FileSelectMsg{Path: "book.txt"})
	updated := model.(App)

	if updated.state != StateProgress {
		t.Errorf("expected state to be StateProgress, got %v", updated.state)
	}

	doneMsg := extractOperationDoneMsg(t, cmd)
	report, ok := doneMsg.Result.(*ebmlib.ValidationReport)
	if !ok {
		t.Fatalf("expected ValidationReport, got %T", doneMsg.Result)
	}

	if report.IsValid {
		t.Error("expected report to be invalid for unsupported file type")
	}

	if report.FilePath != "book.txt" {
		t.Errorf("expected report filepath to be book.txt, got %s", report.FilePath)
	}

	if len(report.Errors) == 0 {
		t.Fatal("expected errors in validation report")
	}

	if report.Errors[0].Code != "SYSTEM_ERROR" {
		t.Errorf("expected SYSTEM_ERROR code, got %s", report.Errors[0].Code)
	}
}

func TestAppUpdateBrowser_StartRepair(t *testing.T) {
	app := NewApp()
	app.state = StateBrowser

	menu := models.NewMenuModel()
	model, _ := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app.menuModel = model.(models.MenuModel)

	model, cmd := app.Update(models.FileSelectMsg{Path: "book.txt"})
	updated := model.(App)

	if updated.state != StateProgress {
		t.Errorf("expected state to be StateProgress, got %v", updated.state)
	}

	doneMsg := extractOperationDoneMsg(t, cmd)
	result, ok := doneMsg.Result.(*ebmlib.RepairResult)
	if !ok {
		t.Fatalf("expected RepairResult, got %T", doneMsg.Result)
	}

	if result.Success {
		t.Error("expected repair result to be unsuccessful for unsupported file type")
	}

	if result.Error == nil {
		t.Fatal("expected repair result to contain error")
	}
}

func TestAppUpdateBrowser_StartBatch(t *testing.T) {
	app := NewApp()
	app.state = StateBrowser

	menu := models.NewMenuModel()
	model, _ := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model, _ = model.(models.MenuModel).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	app.menuModel = model.(models.MenuModel)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "book.epub")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	model, cmd := app.Update(models.FileSelectMsg{Path: filePath})
	updated := model.(App)

	if updated.state != StateProgress {
		t.Errorf("expected state to be StateProgress, got %v", updated.state)
	}

	doneMsg := extractOperationDoneMsg(t, cmd)
	result, ok := doneMsg.Result.(operations.BatchResult)
	if !ok {
		t.Fatalf("expected BatchResult, got %T", doneMsg.Result)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 file in batch result, got %d", result.Total)
	}
}

func TestAppUpdateProgress_ValidationDone(t *testing.T) {
	app := NewApp()
	app.state = StateProgress

	report := &ebmlib.ValidationReport{
		FilePath: "book.epub",
		IsValid:  true,
	}

	model, _ := app.Update(models.OperationDoneMsg{Result: report})
	updated := model.(App)

	if updated.state != StateReport {
		t.Errorf("expected state to be StateReport, got %v", updated.state)
	}

	if !strings.Contains(updated.View(), "Validation Report") {
		t.Error("expected report view to contain validation report title")
	}
}

func TestAppUpdateProgress_RepairDone(t *testing.T) {
	app := NewApp()
	app.state = StateProgress

	result := &ebmlib.RepairResult{Success: true}

	model, _ := app.Update(models.OperationDoneMsg{Result: result})
	updated := model.(App)

	if updated.state != StateReport {
		t.Errorf("expected state to be StateReport, got %v", updated.state)
	}

	if !strings.Contains(updated.View(), "Repair Report") {
		t.Error("expected report view to contain repair report title")
	}
}

func TestAppUpdateProgress_Cancel(t *testing.T) {
	app := NewApp()
	app.state = StateProgress

	model, _ := app.Update(models.OperationCancelMsg{})
	updated := model.(App)

	if updated.state != StateMenu {
		t.Errorf("expected state to be StateMenu, got %v", updated.state)
	}

	select {
	case <-updated.ctx.Done():
		// ok
	default:
		t.Error("expected context to be canceled")
	}
}

func TestAppUpdateProgress_BackToMenu(t *testing.T) {
	app := NewApp()
	app.state = StateProgress

	model, _ := app.Update(models.BackToMenuMsg{})
	updated := model.(App)

	if updated.state != StateMenu {
		t.Errorf("expected state to be StateMenu, got %v", updated.state)
	}
}

func TestAppUpdateProgress_DelegatesToProgressModel(t *testing.T) {
	app := NewApp()
	app.state = StateProgress
	app.progressModel = models.NewProgressModel("Validating", "file.epub", 1, 80, 24)

	model, cmd := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := model.(App)

	if updated.state != StateProgress {
		t.Errorf("expected state to remain StateProgress, got %v", updated.state)
	}

	if cmd != nil {
		t.Error("expected no command from progress update")
	}
}

func TestAppUpdateReport_BackToMenu(t *testing.T) {
	app := NewApp()
	app.state = StateReport
	app.reportModel = models.NewReportModel(&ebmlib.ValidationReport{FilePath: "book.epub"}, 80, 24)

	model, _ := app.Update(models.BackToMenuMsg{})
	updated := model.(App)

	if updated.state != StateMenu {
		t.Errorf("expected state to be StateMenu, got %v", updated.state)
	}
}

func TestAppUpdateReport_DelegatesToReportModel(t *testing.T) {
	app := NewApp()
	app.state = StateReport
	app.reportModel = models.NewReportModel(&ebmlib.ValidationReport{FilePath: "book.epub"}, 80, 24)

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	updated := model.(App)

	if updated.state != StateReport {
		t.Errorf("expected state to remain StateReport, got %v", updated.state)
	}

	if cmd != nil {
		t.Error("expected no command from report update")
	}
}
func extractOperationDoneMsg(t *testing.T, cmd tea.Cmd) models.OperationDoneMsg {
	t.Helper()

	if cmd == nil {
		t.Fatal("expected command to be non-nil")
	}

	msg := cmd()
	return extractOperationDoneFromMsg(t, msg)
}

func extractOperationDoneFromMsg(t *testing.T, msg tea.Msg) models.OperationDoneMsg {
	t.Helper()

	switch m := msg.(type) {
	case models.OperationDoneMsg:
		return m
	case tea.BatchMsg:
		for _, cmd := range m {
			if cmd == nil {
				continue
			}
			sub := cmd()
			if sub == nil {
				continue
			}
			switch subMsg := sub.(type) {
			case models.OperationDoneMsg:
				return subMsg
			case tea.BatchMsg:
				return extractOperationDoneFromMsg(t, subMsg)
			}
		}
	}

	t.Fatalf("expected OperationDoneMsg, got %T", msg)
	return models.OperationDoneMsg{}
}
