package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

// RepairSaveMode describes how repaired files are saved.
type RepairSaveMode string

const (
	RepairSaveModeBackupOriginal RepairSaveMode = "backup-original"
	RepairSaveModeRenameRepaired RepairSaveMode = "rename-repaired"
	RepairSaveModeNoBackup       RepairSaveMode = "no-backup"
)

// RepairModeOption represents a selectable repair save-mode option.
type RepairModeOption struct {
	Label       string
	Description string
	Mode        RepairSaveMode
}

// RepairModeModel lets the user pick how repaired files are saved.
type RepairModeModel struct {
	options  []RepairModeOption
	selected int
	width    int
	height   int
}

// NewRepairModeModel creates a new repair mode model.
func NewRepairModeModel(width, height int) RepairModeModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	return RepairModeModel{
		options: []RepairModeOption{
			{
				Label:       "Backup original, keep repaired name",
				Description: "Creates bookname_original and repairs the original filename",
				Mode:        RepairSaveModeBackupOriginal,
			},
			{
				Label:       "No backup (in-place)",
				Description: "Repairs the original filename without creating backups",
				Mode:        RepairSaveModeNoBackup,
			},
			{
				Label:       "Rename repaired file",
				Description: "Creates bookname_repaired and keeps the original untouched",
				Mode:        RepairSaveModeRenameRepaired,
			},
		},
		selected: 0,
		width:    width,
		height:   height,
	}
}

// Init initializes the model.
func (m RepairModeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m RepairModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			m.selected++
			if m.selected >= len(m.options) {
				m.selected = 0
			}
		case "enter":
			option := m.options[m.selected]
			return m, func() tea.Msg {
				return RepairModeSelectMsg{Mode: option.Mode}
			}
		case "esc", "q":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}
	}

	return m, nil
}

// View renders the model.
func (m RepairModeModel) View() string {
	title := styles.RenderTitle("🛠 Repair Save Mode")
	subtitle := styles.RenderSubtitle("Choose how to save repaired files")

	var options string
	for i, option := range m.options {
		cursor := "  "
		if i == m.selected {
			cursor = styles.IconArrow + " "
			options += styles.SelectedListItemStyle.Render(cursor+option.Label) + "\n"
			options += styles.MutedStyle.Render("  "+option.Description) + "\n"
		} else {
			options += styles.ListItemStyle.Render(cursor+option.Label) + "\n"
			options += styles.MutedStyle.Render("  "+option.Description) + "\n"
		}
		if i < len(m.options)-1 {
			options += "\n"
		}
	}

	modeBox := styles.BorderStyle.
		Width(70).
		Render(options)

	helpText := styles.RenderKeyBinding("↑/↓", "navigate") + "  " +
		styles.RenderKeyBinding("enter", "select") + "  " +
		styles.RenderKeyBinding("esc", "back")

	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(70).
		Render(helpText)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		modeBox,
		"",
		helpBox,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// RepairModeSelectMsg is sent when a mode is selected.
type RepairModeSelectMsg struct {
	Mode RepairSaveMode
}
