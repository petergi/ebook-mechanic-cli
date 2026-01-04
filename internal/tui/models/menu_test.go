package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMenuModel(t *testing.T) {
	m := NewMenuModel()

	if len(m.options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(m.options))
	}

	if m.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", m.selected)
	}

	if m.options[0].Action != "validate" {
		t.Errorf("Expected first option to be 'validate', got '%s'", m.options[0].Action)
	}

	if m.options[3].Action != "quit" {
		t.Errorf("Expected last option to be 'quit', got '%s'", m.options[3].Action)
	}
}

func TestMenuModel_Update_NavigateDown(t *testing.T) {
	m := NewMenuModel()

	// Navigate down
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(MenuModel)

	if m.selected != 1 {
		t.Errorf("Expected selected to be 1, got %d", m.selected)
	}
}

func TestMenuModel_Update_NavigateUp(t *testing.T) {
	m := NewMenuModel()
	m.selected = 1

	// Navigate up
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(MenuModel)

	if m.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", m.selected)
	}
}

func TestMenuModel_Update_WrappingDown(t *testing.T) {
	m := NewMenuModel()
	m.selected = 3 // Last option

	// Navigate down (should wrap to 0)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(MenuModel)

	if m.selected != 0 {
		t.Errorf("Expected selected to wrap to 0, got %d", m.selected)
	}
}

func TestMenuModel_Update_WrappingUp(t *testing.T) {
	m := NewMenuModel()
	m.selected = 0 // First option

	// Navigate up (should wrap to last)
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(MenuModel)

	if m.selected != len(m.options)-1 {
		t.Errorf("Expected selected to wrap to %d, got %d", len(m.options)-1, m.selected)
	}
}

func TestMenuModel_Update_VimKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		initial  int
		expected int
	}{
		{"j moves down", "j", 0, 1},
		{"k moves up", "k", 1, 0},
		{"j wraps at bottom", "j", 3, 0},
		{"k wraps at top", "k", 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMenuModel()
			m.selected = tt.initial

			updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tt.key[0])}})
			m = updatedModel.(MenuModel)

			if m.selected != tt.expected {
				t.Errorf("Expected selected to be %d, got %d", tt.expected, m.selected)
			}
		})
	}
}

func TestMenuModel_Update_SelectOption(t *testing.T) {
	m := NewMenuModel()
	m.selected = 1 // Repair option

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Execute the command to get the message
	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	selectMsg, ok := msg.(MenuSelectMsg)
	if !ok {
		t.Fatal("Expected MenuSelectMsg")
	}

	if selectMsg.Action != "repair" {
		t.Errorf("Expected action 'repair', got '%s'", selectMsg.Action)
	}
}

func TestMenuModel_Update_QuitOption(t *testing.T) {
	m := NewMenuModel()
	m.selected = 3 // Quit option

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(MenuModel)

	if !m.quitting {
		t.Error("Expected quitting to be true")
	}

	// Should return tea.Quit command
	if cmd == nil {
		t.Fatal("Expected quit command to be non-nil")
	}
}

func TestMenuModel_Update_QuitKey(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"esc key", tea.KeyMsg{Type: tea.KeyEsc}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMenuModel()
			updatedModel, cmd := m.Update(tt.key)
			m = updatedModel.(MenuModel)

			if !m.quitting {
				t.Error("Expected quitting to be true")
			}

			if cmd == nil {
				t.Fatal("Expected quit command to be non-nil")
			}
		})
	}
}

func TestMenuModel_Update_WindowSize(t *testing.T) {
	m := NewMenuModel()

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updatedModel.(MenuModel)

	if m.width != 100 {
		t.Errorf("Expected width to be 100, got %d", m.width)
	}

	if m.height != 30 {
		t.Errorf("Expected height to be 30, got %d", m.height)
	}
}

func TestMenuModel_SelectedAction(t *testing.T) {
	m := NewMenuModel()

	tests := []struct {
		selected int
		expected string
	}{
		{0, "validate"},
		{1, "repair"},
		{2, "batch"},
		{3, "quit"},
	}

	for _, tt := range tests {
		m.selected = tt.selected
		action := m.SelectedAction()

		if action != tt.expected {
			t.Errorf("Expected action '%s', got '%s'", tt.expected, action)
		}
	}
}

func TestMenuModel_SelectedAction_OutOfBounds(t *testing.T) {
	m := NewMenuModel()
	m.selected = -1

	action := m.SelectedAction()
	if action != "" {
		t.Errorf("Expected empty action for out of bounds, got '%s'", action)
	}

	m.selected = 999
	action = m.SelectedAction()
	if action != "" {
		t.Errorf("Expected empty action for out of bounds, got '%s'", action)
	}
}

func TestMenuModel_View(t *testing.T) {
	m := NewMenuModel()

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// View should contain option labels
	// Note: We're not checking exact rendering as it depends on lipgloss,
	// but we can check that the view is non-empty as a basic sanity check
}

func TestMenuModel_View_Quitting(t *testing.T) {
	m := NewMenuModel()
	m.quitting = true

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view when quitting")
	}
}

func TestMenuModel_Init(t *testing.T) {
	m := NewMenuModel()
	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}
