package models

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
)

func TestNewProgressModel(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	if m.operation != "Validating" {
		t.Errorf("Expected operation 'Validating', got '%s'", m.operation)
	}

	if m.filePath != "test.epub" {
		t.Errorf("Expected filePath 'test.epub', got '%s'", m.filePath)
	}

	if m.total != 1 {
		t.Errorf("Expected total 1, got %d", m.total)
	}

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}

	if m.done {
		t.Error("Expected done to be false initially")
	}

	if m.current != 0 {
		t.Errorf("Expected current to be 0, got %d", m.current)
	}

	// Verify start time was set (should be very recent)
	if time.Since(m.startTime) > time.Second {
		t.Error("Expected startTime to be recent")
	}
}

func TestProgressModel_Init(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("Expected Init to return non-nil command")
	}

	// Execute the command to verify it returns a TickMsg
	msg := cmd()
	if _, ok := msg.(TickMsg); !ok {
		t.Error("Expected Init command to return TickMsg")
	}
}

func TestProgressModel_Update_WindowSize(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updatedModel.(ProgressModel)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestProgressModel_Update_TickMsg(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	initialSpinner := m.spinner

	updatedModel, cmd := m.Update(TickMsg(time.Now()))
	m = updatedModel.(ProgressModel)

	// Spinner should have advanced
	if m.spinner == initialSpinner {
		t.Error("Expected spinner to advance")
	}

	// Should return another tick command
	if cmd == nil {
		t.Error("Expected tick command to continue animation")
	}
}

func TestProgressModel_Update_TickMsg_WhenDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	m.done = true

	updatedModel, cmd := m.Update(TickMsg(time.Now()))
	m = updatedModel.(ProgressModel)

	// Should not return tick command when done
	if cmd != nil {
		t.Error("Expected no tick command when done")
	}
}

func TestProgressModel_Update_ProgressUpdate(t *testing.T) {
	m := NewProgressModel("Batch Processing", "", 10, 80, 24)

	msg := ProgressUpdateMsg{
		Current:     5,
		Total:       10,
		CurrentFile: "file5.epub",
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(ProgressModel)

	if m.current != 5 {
		t.Errorf("Expected current 5, got %d", m.current)
	}

	if m.currentFile != "file5.epub" {
		t.Errorf("Expected currentFile 'file5.epub', got '%s'", m.currentFile)
	}
}

func TestProgressModel_Update_OperationDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	testResult := "test result"
	msg := OperationDoneMsg{Result: testResult}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(ProgressModel)

	if !m.done {
		t.Error("Expected done to be true")
	}

	if m.result != testResult {
		t.Errorf("Expected result to be set to '%v'", testResult)
	}
}

func TestProgressModel_Update_CancelKey_WhenNotDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	if _, ok := msg.(OperationCancelMsg); !ok {
		t.Error("Expected OperationCancelMsg")
	}
}

func TestProgressModel_Update_EnterKey_WhenDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	m.done = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	if _, ok := msg.(BackToMenuMsg); !ok {
		t.Error("Expected BackToMenuMsg")
	}
}

func TestProgressModel_Update_EscKey_WhenDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	m.done = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	if _, ok := msg.(BackToMenuMsg); !ok {
		t.Error("Expected BackToMenuMsg")
	}
}

func TestProgressModel_Update_QuitKey_WhenDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	m.done = true

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

			// The command should be tea.Quit
			// We can't easily test the exact command, but we verified it's non-nil
		})
	}
}

func TestProgressModel_Update_IgnoreKeysWhenNotDone(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}},
		{"esc", tea.KeyMsg{Type: tea.KeyEsc}},
		{"q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, _ := m.Update(tt.key)
			updated := updatedModel.(ProgressModel)

			// Model should not change state (shouldn't quit or go to menu)
			if updated.done {
				t.Error("Model should not be done")
			}
		})
	}
}

func TestProgressModel_View_InProgress(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// View should contain operation name
	// Note: We can't check exact content due to lipgloss rendering,
	// but we verified it's non-empty
}

func TestProgressModel_View_Done(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)
	m.done = true

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view when done")
	}
}

func TestProgressModel_View_BatchProgress(t *testing.T) {
	m := NewProgressModel("Batch Processing", "", 10, 80, 24)
	m.current = 5

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for batch progress")
	}

	// When total > 1, should show progress bar
	// We can't easily check the exact rendering, but we verified the view exists
}

func TestProgressModel_View_SingleFile(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view for single file")
	}

	// When total == 1, should show single file display
	// We can't easily check the exact rendering, but we verified the view exists
}

func TestProgressModel_View_CurrentFile(t *testing.T) {
	m := NewProgressModel("Batch Processing", "", 10, 80, 24)
	m.current = 5
	m.currentFile = "current.epub"

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view with current file")
	}
}

func TestProgressModel_SpinnerAnimation(t *testing.T) {
	m := NewProgressModel("Validating", "test.epub", 1, 80, 24)

	// Simulate multiple ticks to verify spinner cycles
	for i := 0; i < 10; i++ {
		updatedModel, _ := m.Update(TickMsg(time.Now()))
		m = updatedModel.(ProgressModel)
	}

	// Spinner should have cycled (we can't predict exact value,
	// but it should have changed from initial 0)
	// This test mainly verifies no panic occurs during animation
}

func TestConvertBatchProgress(t *testing.T) {
	batchUpdate := operations.ProgressUpdate{
		Completed: 5,
		Total:     10,
		Current:   "file5.epub",
	}

	msg := ConvertBatchProgress(batchUpdate)

	if msg.Current != 5 {
		t.Errorf("Expected Current 5, got %d", msg.Current)
	}

	if msg.Total != 10 {
		t.Errorf("Expected Total 10, got %d", msg.Total)
	}

	if msg.CurrentFile != "file5.epub" {
		t.Errorf("Expected CurrentFile 'file5.epub', got '%s'", msg.CurrentFile)
	}
}

func TestProgressUpdateMsg(t *testing.T) {
	// Test that ProgressUpdateMsg type exists and can be created
	msg := ProgressUpdateMsg{
		Current:     3,
		Total:       10,
		CurrentFile: "test.epub",
	}

	if msg.Current != 3 {
		t.Errorf("Expected Current 3, got %d", msg.Current)
	}

	if msg.Total != 10 {
		t.Errorf("Expected Total 10, got %d", msg.Total)
	}

	if msg.CurrentFile != "test.epub" {
		t.Errorf("Expected CurrentFile 'test.epub', got '%s'", msg.CurrentFile)
	}
}

func TestOperationDoneMsg(t *testing.T) {
	// Test that OperationDoneMsg type exists and can be created
	result := "test result"
	msg := OperationDoneMsg{Result: result}

	if msg.Result != result {
		t.Errorf("Expected Result '%v', got '%v'", result, msg.Result)
	}
}

func TestOperationCancelMsg(t *testing.T) {
	// Test that OperationCancelMsg type exists and can be created
	msg := OperationCancelMsg{}

	// Type should exist and be instantiable
	_ = msg
}

func TestTickMsg(t *testing.T) {
	// Test that TickMsg type exists and can be created
	now := time.Now()
	msg := TickMsg(now)

	// Convert back to time.Time
	timeVal := time.Time(msg)

	if !timeVal.Equal(now) {
		t.Error("Expected TickMsg to preserve time value")
	}
}

func TestProgressModel_MultipleUpdates(t *testing.T) {
	m := NewProgressModel("Batch Processing", "", 10, 80, 24)

	// Simulate a series of progress updates
	updates := []ProgressUpdateMsg{
		{Current: 1, Total: 10, CurrentFile: "file1.epub"},
		{Current: 2, Total: 10, CurrentFile: "file2.epub"},
		{Current: 5, Total: 10, CurrentFile: "file5.epub"},
		{Current: 10, Total: 10, CurrentFile: "file10.epub"},
	}

	for _, update := range updates {
		var updatedModel tea.Model
		updatedModel, _ = m.Update(update)
		m = updatedModel.(ProgressModel)
	}

	// Should end at final update
	if m.current != 10 {
		t.Errorf("Expected current 10, got %d", m.current)
	}

	if m.currentFile != "file10.epub" {
		t.Errorf("Expected currentFile 'file10.epub', got '%s'", m.currentFile)
	}
}

func TestProgressModel_ZeroTotalHandling(t *testing.T) {
	m := NewProgressModel("Batch Processing", "", 0, 80, 24)
	m.current = 0

	// Should not panic with zero total
	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view even with zero total")
	}
}
