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
				Description: "Check a single ebook file for errors",
				Action:      "validate",
			},
			{
				Label:       "Repair EPUB/PDF",
				Description: "Automatically fix a single ebook file",
				Action:      "repair",
			},
			{
				Label:       "Validate Multiple Files",
				Description: "Select multiple files for validation",
				Action:      "multi-validate",
			},
			{
				Label:       "Repair Multiple Files",
				Description: "Select multiple files for repair",
				Action:      "multi-repair",
			},
			{
				Label:       "Batch Process Directory",
				Description: "Validate or repair all files in a folder",
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

	// Title with decorative separator
	title := styles.RenderTitle("📚 Ebook Mechanic")
	subtitle := styles.RenderSubtitle("Choose an operation")
	separator := styles.MutedStyle.Render(lipgloss.PlaceHorizontal(50, lipgloss.Left, "─────────────────────────────────────────────────"))

	// Render menu options in a bordered box
	var options string
	for i, opt := range m.options {
		cursor := "  "
		if i == m.selected {
			cursor = styles.IconArrow + " "
			options += styles.SelectedListItemStyle.Render(cursor+opt.Label) + "\n"
			options += styles.MutedStyle.Render("  "+opt.Description) + "\n"
		} else {
			options += styles.ListItemStyle.Render(cursor+opt.Label) + "\n"
			options += styles.MutedStyle.Render("  "+opt.Description) + "\n"
		}
		if i < len(m.options)-1 {
			options += "\n"
		}
	}

	// Wrap options in a bordered box
	menuBox := styles.BorderStyle.
		Width(50).
		Render(options)

	// Help text in a subtle box
	helpText := styles.RenderKeyBinding("↑/↓", "navigate") + "  " +
		styles.RenderKeyBinding("enter", "select") + "  " +
		styles.RenderKeyBinding("q", "quit")

	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(50).
		Render(helpText)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		separator,
		"",
		menuBox,
		"",
		helpBox,
	)

	// Center in terminal
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
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

// Quitting returns true if the user wants to quit
func (m MenuModel) Quitting() bool {
	return m.quitting
}
