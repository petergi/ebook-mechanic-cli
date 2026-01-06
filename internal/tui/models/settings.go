package models

import (
	"fmt"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

const (
	settingsMinJobs = 1
	settingsMaxJobs = 64
)

// SettingsModel manages TUI settings.
type SettingsModel struct {
	jobs           int
	skipValidation bool
	selected       int
	width          int
	height         int
}

// SettingsSaveMsg is sent when settings are saved.
type SettingsSaveMsg struct {
	Jobs           int
	SkipValidation bool
}

// NewSettingsModel creates a new settings model.
func NewSettingsModel(jobs int, skipValidation bool, width, height int) SettingsModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}
	if jobs < settingsMinJobs {
		jobs = settingsMinJobs
	}

	return SettingsModel{
		jobs:           jobs,
		skipValidation: skipValidation,
		selected:       0,
		width:          width,
		height:         height,
	}
}

// Init initializes the model.
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.selected = 2
			}
		case "down", "j":
			m.selected++
			if m.selected > 2 {
				m.selected = 0
			}
		case "-":
			if m.selected == 0 {
				m.jobs--
				if m.jobs < settingsMinJobs {
					m.jobs = settingsMinJobs
				}
			}
		case "+":
			if m.selected == 0 {
				m.jobs++
				if m.jobs > settingsMaxJobs {
					m.jobs = settingsMaxJobs
				}
			}
		case " ":
			if m.selected == 1 {
				m.skipValidation = !m.skipValidation
			}
		case "enter":
			switch m.selected {
			case 1:
				m.skipValidation = !m.skipValidation
			case 2:
				return m, func() tea.Msg {
					return SettingsSaveMsg{Jobs: m.jobs, SkipValidation: m.skipValidation}
				}
			}
		case "esc", "q":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}
	}

	return m, nil
}

// View renders the settings.
func (m SettingsModel) View() string {
	title := styles.RenderTitle("⚙ Settings")
	subtitle := styles.RenderSubtitle("Adjust batch jobs and validation")

	jobsLabel := fmt.Sprintf("Batch jobs: %d", m.jobs)
	validationLabel := fmt.Sprintf("Skip post-repair validation: %v", m.skipValidation)
	doneLabel := "Done"

	items := []string{jobsLabel, validationLabel, doneLabel}
	var rendered string
	for i, item := range items {
		cursor := "  "
		if i == m.selected {
			cursor = styles.IconArrow + " "
			rendered += styles.SelectedListItemStyle.Render(cursor+item) + "\n"
		} else {
			rendered += styles.ListItemStyle.Render(cursor+item) + "\n"
		}
		if i < len(items)-1 {
			rendered += "\n"
		}
	}

	note := styles.MutedStyle.Render("Tip: Use +/- only on the Batch jobs row.")

	settingsBox := styles.BorderStyle.
		Width(70).
		Render(rendered + "\n" + note)

	helpText := styles.RenderKeyBinding("↑/↓", "navigate") + "  " +
		styles.RenderKeyBinding("+/-", "change jobs") + "  " +
		styles.RenderKeyBinding("space", "toggle") + "  " +
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
		settingsBox,
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
