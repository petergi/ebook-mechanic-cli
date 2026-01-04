package models

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/example/project/pkg/ebmlib"
)

func TestNewReportModel(t *testing.T) {
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  true,
	}

	m := NewReportModel(report)

	if m.report != report {
		t.Error("Expected report to be set")
	}

	if m.reportType != "validation" {
		t.Errorf("Expected reportType 'validation', got '%s'", m.reportType)
	}

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if m.viewportSize != 15 {
		t.Errorf("Expected viewportSize 15, got %d", m.viewportSize)
	}

	if !m.showErrors || !m.showWarnings || !m.showInfo {
		t.Error("Expected all filters to be enabled by default")
	}

	if m.selectedFilter != 0 {
		t.Errorf("Expected selectedFilter 0, got %d", m.selectedFilter)
	}
}

func TestNewRepairReportModel(t *testing.T) {
	result := &ebmlib.RepairResult{
		Success: true,
	}

	m := NewRepairReportModel(result)

	if m.repairResult != result {
		t.Error("Expected repairResult to be set")
	}

	if m.reportType != "repair" {
		t.Errorf("Expected reportType 'repair', got '%s'", m.reportType)
	}

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if !m.showErrors || !m.showWarnings || !m.showInfo {
		t.Error("Expected all filters to be enabled by default")
	}
}

func TestReportModel_Init(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestReportModel_Update_WindowSize(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updatedModel.(ReportModel)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}

	expectedViewportSize := 40 - 12
	if m.viewportSize != expectedViewportSize {
		t.Errorf("Expected viewportSize %d, got %d", expectedViewportSize, m.viewportSize)
	}
}

func TestReportModel_Update_NavigateDown(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 0

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}},
		{"j key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.viewportTop = 0
			updatedModel, _ := m.Update(tt.key)
			updated := updatedModel.(ReportModel)

			if updated.viewportTop != 1 {
				t.Errorf("Expected viewportTop to increment to 1, got %d", updated.viewportTop)
			}
		})
	}
}

func TestReportModel_Update_NavigateUp(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 5

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}},
		{"k key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.viewportTop = 5
			updatedModel, _ := m.Update(tt.key)
			updated := updatedModel.(ReportModel)

			if updated.viewportTop != 4 {
				t.Errorf("Expected viewportTop to decrement to 4, got %d", updated.viewportTop)
			}
		})
	}
}

func TestReportModel_Update_NavigateUp_AtTop(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 0

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated := updatedModel.(ReportModel)

	// Should stay at 0
	if updated.viewportTop != 0 {
		t.Errorf("Expected viewportTop to stay at 0, got %d", updated.viewportTop)
	}
}

func TestReportModel_Update_Filter_All(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 10
	m.showErrors = false
	m.showWarnings = false
	m.showInfo = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = updatedModel.(ReportModel)

	if !m.showErrors || !m.showWarnings || !m.showInfo {
		t.Error("Expected all filters to be enabled for '1' key")
	}

	if m.selectedFilter != 0 {
		t.Errorf("Expected selectedFilter 0, got %d", m.selectedFilter)
	}

	if m.viewportTop != 0 {
		t.Error("Expected viewportTop to reset to 0 when filter changes")
	}
}

func TestReportModel_Update_Filter_Errors(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 10

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = updatedModel.(ReportModel)

	if !m.showErrors {
		t.Error("Expected showErrors to be true for '2' key")
	}

	if m.showWarnings || m.showInfo {
		t.Error("Expected showWarnings and showInfo to be false for '2' key")
	}

	if m.selectedFilter != 1 {
		t.Errorf("Expected selectedFilter 1, got %d", m.selectedFilter)
	}

	if m.viewportTop != 0 {
		t.Error("Expected viewportTop to reset to 0 when filter changes")
	}
}

func TestReportModel_Update_Filter_Warnings(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = updatedModel.(ReportModel)

	if !m.showWarnings {
		t.Error("Expected showWarnings to be true for '3' key")
	}

	if m.showErrors || m.showInfo {
		t.Error("Expected showErrors and showInfo to be false for '3' key")
	}

	if m.selectedFilter != 2 {
		t.Errorf("Expected selectedFilter 2, got %d", m.selectedFilter)
	}
}

func TestReportModel_Update_Filter_Info(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = updatedModel.(ReportModel)

	if !m.showInfo {
		t.Error("Expected showInfo to be true for '4' key")
	}

	if m.showErrors || m.showWarnings {
		t.Error("Expected showErrors and showWarnings to be false for '4' key")
	}

	if m.selectedFilter != 3 {
		t.Errorf("Expected selectedFilter 3, got %d", m.selectedFilter)
	}
}

func TestReportModel_Update_BackToMenu(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"enter key", tea.KeyMsg{Type: tea.KeyEnter}},
		{"esc key", tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := m.Update(tt.key)

			if cmd == nil {
				t.Fatal("Expected command to be non-nil")
			}

			msg := cmd()
			if _, ok := msg.(BackToMenuMsg); !ok {
				t.Error("Expected BackToMenuMsg")
			}
		})
	}
}

func TestReportModel_Update_Quit(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := m.Update(tt.key)

			if cmd == nil {
				t.Fatal("Expected quit command to be non-nil")
			}
		})
	}
}

func TestReportModel_View_Validation(t *testing.T) {
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  true,
		Errors:   []ebmlib.ValidationError{},
		Warnings: []ebmlib.ValidationError{},
		Info:     []ebmlib.ValidationError{},
	}

	m := NewReportModel(report)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for validation report")
	}
}

func TestReportModel_View_ValidationWithErrors(t *testing.T) {
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  false,
		Errors: []ebmlib.ValidationError{
			{
				Code:     "TEST_ERROR",
				Message:  "Test error message",
				Severity: ebmlib.SeverityError,
			},
		},
		Warnings: []ebmlib.ValidationError{},
		Info:     []ebmlib.ValidationError{},
	}

	m := NewReportModel(report)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for validation report with errors")
	}
}

func TestReportModel_View_Repair(t *testing.T) {
	result := &ebmlib.RepairResult{
		Success: true,
	}

	m := NewRepairReportModel(result)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for repair report")
	}
}

func TestReportModel_View_RepairWithBackup(t *testing.T) {
	result := &ebmlib.RepairResult{
		Success:    true,
		BackupPath: "/path/to/backup.epub",
	}

	m := NewRepairReportModel(result)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for repair report with backup")
	}
}

func TestReportModel_View_RepairFailed(t *testing.T) {
	result := &ebmlib.RepairResult{
		Success: false,
		Error:   errors.New("test repair error"),
	}

	m := NewRepairReportModel(result)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for failed repair report")
	}
}

func TestReportModel_View_RepairWithActions(t *testing.T) {
	result := &ebmlib.RepairResult{
		Success: true,
		ActionsApplied: []ebmlib.RepairAction{
			{
				Type:        "fix_metadata",
				Description: "Fixed metadata",
			},
			{
				Type:        "fix_structure",
				Description: "Fixed structure",
			},
		},
	}

	m := NewRepairReportModel(result)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for repair report with actions")
	}
}

func TestReportModel_View_NullReport(t *testing.T) {
	m := ReportModel{
		reportType: "validation",
		report:     nil,
	}

	view := m.View()

	if view == "" {
		t.Error("Expected error message for nil report")
	}
}

func TestReportModel_View_NullRepairResult(t *testing.T) {
	m := ReportModel{
		reportType:   "repair",
		repairResult: nil,
	}

	view := m.View()

	if view == "" {
		t.Error("Expected error message for nil repair result")
	}
}

func TestReportModel_formatIssue_Error(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	issue := ebmlib.ValidationError{
		Code:     "TEST_ERROR",
		Message:  "Test error",
		Severity: ebmlib.SeverityError,
	}

	formatted := m.formatIssue(issue, "error")

	if formatted == "" {
		t.Error("Expected non-empty formatted error")
	}
}

func TestReportModel_formatIssue_Warning(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	issue := ebmlib.ValidationError{
		Code:     "TEST_WARNING",
		Message:  "Test warning",
		Severity: ebmlib.SeverityWarning,
	}

	formatted := m.formatIssue(issue, "warning")

	if formatted == "" {
		t.Error("Expected non-empty formatted warning")
	}
}

func TestReportModel_formatIssue_Info(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	issue := ebmlib.ValidationError{
		Code:     "TEST_INFO",
		Message:  "Test info",
		Severity: ebmlib.SeverityInfo,
	}

	formatted := m.formatIssue(issue, "info")

	if formatted == "" {
		t.Error("Expected non-empty formatted info")
	}
}

func TestReportModel_formatIssue_WithLocation(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)

	issue := ebmlib.ValidationError{
		Code:     "TEST_ERROR",
		Message:  "Test error",
		Severity: ebmlib.SeverityError,
		Location: &ebmlib.ErrorLocation{
			File: "content.opf",
			Line: 42,
		},
	}

	formatted := m.formatIssue(issue, "error")

	if formatted == "" {
		t.Error("Expected non-empty formatted error with location")
	}
}

func TestReportModel_ViewportScrolling(t *testing.T) {
	// Create report with many errors to test scrolling
	errors := make([]ebmlib.ValidationError, 50)
	for i := 0; i < 50; i++ {
		errors[i] = ebmlib.ValidationError{
			Code:     "ERROR",
			Message:  "Error message",
			Severity: ebmlib.SeverityError,
		}
	}

	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  false,
		Errors:   errors,
	}

	m := NewReportModel(report)
	m.viewportSize = 10

	// Test scrolling down
	for i := 0; i < 5; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updatedModel.(ReportModel)
	}

	if m.viewportTop != 5 {
		t.Errorf("Expected viewportTop 5 after scrolling, got %d", m.viewportTop)
	}

	// Test scrolling up
	for i := 0; i < 3; i++ {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = updatedModel.(ReportModel)
	}

	if m.viewportTop != 2 {
		t.Errorf("Expected viewportTop 2 after scrolling up, got %d", m.viewportTop)
	}
}

func TestReportModel_FilterChangeResetsScroll(t *testing.T) {
	report := &ebmlib.ValidationReport{FilePath: "test.epub"}
	m := NewReportModel(report)
	m.viewportTop = 10

	// Change to errors filter
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = updatedModel.(ReportModel)

	if m.viewportTop != 0 {
		t.Error("Expected viewportTop to reset to 0 when changing filter")
	}

	m.viewportTop = 5

	// Change to warnings filter
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = updatedModel.(ReportModel)

	if m.viewportTop != 0 {
		t.Error("Expected viewportTop to reset to 0 when changing filter")
	}
}

func TestReportModel_EmptyReportDisplay(t *testing.T) {
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  true,
		Errors:   []ebmlib.ValidationError{},
		Warnings: []ebmlib.ValidationError{},
		Info:     []ebmlib.ValidationError{},
	}

	m := NewReportModel(report)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for empty report")
	}
}

func TestReportModel_MultipleIssueTypes(t *testing.T) {
	report := &ebmlib.ValidationReport{
		FilePath: "test.epub",
		IsValid:  false,
		Errors: []ebmlib.ValidationError{
			{Code: "ERROR1", Message: "Error 1", Severity: ebmlib.SeverityError},
		},
		Warnings: []ebmlib.ValidationError{
			{Code: "WARN1", Message: "Warning 1", Severity: ebmlib.SeverityWarning},
		},
		Info: []ebmlib.ValidationError{
			{Code: "INFO1", Message: "Info 1", Severity: ebmlib.SeverityInfo},
		},
	}

	m := NewReportModel(report)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view with mixed issue types")
	}
}
