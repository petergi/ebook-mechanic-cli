package models

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

// MenuOption represents a selectable option in the menu
type MenuOption struct {
	Label       string
	Description string
	Action      string // Action identifier (e.g., "validate", "repair", "batch", "quit")
}

// MenuModel represents the main menu state
type MenuModel struct {
	options  []MenuOption
	selected int
	width    int
	height   int
	quitting bool
}

// NewMenuModel creates a new menu model with the given options
func NewMenuModel() MenuModel {
	return MenuModel{
		options: []MenuOption{
			{
				Label:       "Validate EPUB/PDF",
				Description: "Check ebook files for errors and issues",
				Action:      "validate",
			},
			{
				Label:       "Repair EPUB/PDF",
				Description: "Automatically fix common ebook problems",
				Action:      "repair",
			},
			{
				Label:       "Batch Process",
				Description: "Validate or repair multiple files at once",
				Action:      "batch",
			},
			{
				Label:       "Quit",
				Description: "Exit the application",
				Action:      "quit",
			},
		},
		selected: 0,
		width:    80,
		height:   24,
	}
}

// Init initializes the model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state
//
// This is where the navigation logic lives. You'll implement:
// - Arrow key navigation (up/down to move selection)
// - Enter key to select an option
// - ESC or 'q' to quit
// - Wrapping behavior (should up from top go to bottom? vice versa?)
//
// TODO: Implement the navigation logic
// Consider:
// - Should navigation wrap around (up from index 0 goes to last item)?
// - Should 'j'/'k' vim keys be supported in addition to arrows?
// - How should enter key be handled (return a custom message)?
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.selected = len(m.options) - 1 // Wrap to bottom
			}

		case "down", "j":
			m.selected++
			if m.selected >= len(m.options) {
				m.selected = 0 // Wrap to top
			}

		case "enter":
			// User selected an option, return message with the action
			action := m.SelectedAction()
			if action == "quit" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, func() tea.Msg {
				return MenuSelectMsg{Action: action}
			}

		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		default:
			// Ignore other keys
		}
	}

	return m, nil
}

// View renders the menu
func (m MenuModel) View() string {
	if m.quitting {
		return styles.RenderInfo("Goodbye!") + "\n"
	}

	// Title
	title := styles.RenderTitle("📚 Ebook Mechanic")
	subtitle := styles.RenderSubtitle("Choose an operation")

	// Render menu options
	var options string
	for i, opt := range m.options {
		cursor := " "
		if i == m.selected {
			cursor = styles.IconArrow
			options += styles.SelectedListItemStyle.Render(cursor+" "+opt.Label) + "\n"
			options += styles.MutedStyle.Render("  "+opt.Description) + "\n\n"
		} else {
			options += styles.ListItemStyle.Render(cursor+" "+opt.Label) + "\n"
			options += styles.MutedStyle.Render("  "+opt.Description) + "\n\n"
		}
	}

	// Help text
	help := styles.HelpStyle.Render(
		styles.RenderKeyBinding("↑/↓", "navigate") + "  " +
			styles.RenderKeyBinding("enter", "select") + "  " +
			styles.RenderKeyBinding("q", "quit"),
	)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		options,
		help,
	)

	// Center in terminal
	return styles.DocStyle.
		Width(m.width).
		Height(m.height).
		Render(content)
}

// SelectedAction returns the action of the currently selected option
func (m MenuModel) SelectedAction() string {
	if m.selected >= 0 && m.selected < len(m.options) {
		return m.options[m.selected].Action
	}
	return ""
}

// MenuSelectMsg is sent when a menu option is selected
type MenuSelectMsg struct {
	Action string
}
