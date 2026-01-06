package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRepairModeModel(t *testing.T) {
	m := NewRepairModeModel(80, 24)
	if len(m.options) == 0 {
		t.Fatal("expected repair mode options")
	}
	if m.selected != 0 {
		t.Errorf("expected default selection 0, got %d", m.selected)
	}
}

func TestRepairModeModel_Select(t *testing.T) {
	m := NewRepairModeModel(80, 24)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}
	msg := cmd()
	selectMsg, ok := msg.(RepairModeSelectMsg)
	if !ok {
		t.Fatalf("expected RepairModeSelectMsg, got %T", msg)
	}
	if selectMsg.Mode != RepairSaveModeBackupOriginal {
		t.Errorf("expected backup-original, got %s", selectMsg.Mode)
	}
}

func TestRepairModeModel_View(t *testing.T) {
	m := NewRepairModeModel(80, 24)
	if view := m.View(); view == "" {
		t.Error("expected non-empty view")
	}
}
