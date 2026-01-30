package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel(4, false, false, false, false, false, true, 80, 24)
	if m.jobs != 4 {
		t.Errorf("expected jobs 4, got %d", m.jobs)
	}
	if m.skipValidation {
		t.Error("expected skipValidation false")
	}
}

func TestSettingsModel_Update_ChangeJobs(t *testing.T) {
	m := NewSettingsModel(4, false, false, false, false, false, true, 80, 24)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = updated.(SettingsModel)
	if m.jobs != 5 {
		t.Errorf("expected jobs 5, got %d", m.jobs)
	}
}

func TestSettingsModel_Update_ToggleSkipValidation(t *testing.T) {
	m := NewSettingsModel(4, false, false, false, false, false, true, 80, 24)
	m.selected = 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(SettingsModel)
	if !m.skipValidation {
		t.Error("expected skipValidation true")
	}
}

func TestSettingsModel_Update_Save(t *testing.T) {
	m := NewSettingsModel(4, false, false, false, false, false, true, 80, 24)
	m.selected = 4
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected save command")
	}
	msg := cmd()
	saveMsg, ok := msg.(SettingsSaveMsg)
	if !ok {
		t.Fatalf("expected SettingsSaveMsg, got %T", msg)
	}
	if saveMsg.Jobs != 4 {
		t.Errorf("expected jobs 4, got %d", saveMsg.Jobs)
	}
	if saveMsg.SkipValidation {
		t.Error("expected skipValidation false")
	}
	if saveMsg.NoBackup {
		t.Error("expected noBackup false")
	}
	if saveMsg.Aggressive {
		t.Error("expected aggressive false")
	}
}
