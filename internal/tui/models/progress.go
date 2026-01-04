package models

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

// ProgressModel displays progress for ongoing operations
type ProgressModel struct {
	operation   string // "Validating", "Repairing", "Batch Processing"
	filePath    string
	current     int
	total       int
	currentFile string
	startTime   time.Time
	width       int
	height      int
	spinner     int
	done        bool
	result      interface{} // Holds the final result when done
}

// NewProgressModel creates a new progress model
func NewProgressModel(operation string, filePath string, total int) ProgressModel {
	return ProgressModel{
		operation: operation,
		filePath:  filePath,
		total:     total,
		startTime: time.Now(),
		width:     80,
		height:    24,
	}
}

// Init initializes the model
func (m ProgressModel) Init() tea.Cmd {
	return m.tick()
}

// Update handles messages and updates the model state
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "enter", "esc":
				// Return to menu
				return m, func() tea.Msg {
					return BackToMenuMsg{}
				}
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		} else {
			switch msg.String() {
			case "ctrl+c":
				// Cancel operation
				return m, func() tea.Msg {
					return OperationCancelMsg{}
				}
			}
		}

	case TickMsg:
		if !m.done {
			m.spinner = (m.spinner + 1) % len(styles.IconSpinner)
			return m, m.tick()
		}

	case ProgressUpdateMsg:
		m.current = msg.Current
		m.currentFile = msg.CurrentFile
		return m, nil

	case OperationDoneMsg:
		m.done = true
		m.result = msg.Result
		return m, nil
	}

	return m, nil
}

// View renders the progress display
func (m ProgressModel) View() string {
	if m.done {
		return m.renderDone()
	}

	// Title
	title := styles.RenderTitle(m.operation)

	// Spinner
	spinnerChar := string(styles.IconSpinner[m.spinner])
	spinnerDisplay := styles.InfoStyle.Render(spinnerChar + " Processing...")

	// Current file (for single file operations)
	var fileDisplay string
	if m.total == 1 && m.filePath != "" {
		fileDisplay = styles.MutedStyle.Render("File: " + m.filePath)
	} else if m.currentFile != "" {
		fileDisplay = styles.MutedStyle.Render("Current: " + m.currentFile)
	}

	// Progress bar (for batch operations)
	var progressDisplay string
	if m.total > 1 {
		percentage := 0
		if m.total > 0 {
			percentage = (m.current * 100) / m.total
		}

		progressBar := styles.RenderProgressBar(m.current, m.total, 50)
		progressText := fmt.Sprintf("%d / %d (%d%%)", m.current, m.total, percentage)

		progressDisplay = lipgloss.JoinVertical(
			lipgloss.Left,
			progressBar,
			styles.MutedStyle.Render(progressText),
		)
	}

	// Elapsed time
	elapsed := time.Since(m.startTime)
	timeDisplay := styles.MutedStyle.Render(
		fmt.Sprintf("Elapsed: %s", elapsed.Round(time.Second)),
	)

	// Help text
	help := styles.HelpStyle.Render(
		styles.RenderKeyBinding("ctrl+c", "cancel"),
	)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		spinnerDisplay,
		fileDisplay,
		"",
		progressDisplay,
		"",
		timeDisplay,
		"",
		help,
	)

	return styles.DocStyle.
		Width(m.width).
		Height(m.height).
		Render(content)
}

// renderDone renders the completion screen
func (m ProgressModel) renderDone() string {
	elapsed := time.Since(m.startTime)

	title := styles.RenderSuccess(m.operation + " Complete!")
	timeDisplay := styles.MutedStyle.Render(
		fmt.Sprintf("Completed in: %s", elapsed.Round(time.Millisecond)),
	)

	help := styles.HelpStyle.Render(
		styles.RenderKeyBinding("enter", "continue"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		timeDisplay,
		"",
		help,
	)

	return styles.DocStyle.
		Width(m.width).
		Height(m.height).
		Render(content)
}

// tick returns a command that triggers a spinner animation update
func (m ProgressModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// TickMsg is sent periodically to update the spinner
type TickMsg time.Time

// ProgressUpdateMsg updates the progress display
type ProgressUpdateMsg struct {
	Current     int
	Total       int
	CurrentFile string
}

// OperationDoneMsg signals that the operation is complete
type OperationDoneMsg struct {
	Result interface{}
}

// OperationCancelMsg signals that the user wants to cancel
type OperationCancelMsg struct{}

// ConvertBatchProgress converts batch progress to progress message
func ConvertBatchProgress(update operations.ProgressUpdate) ProgressUpdateMsg {
	return ProgressUpdateMsg{
		Current:     update.Completed,
		Total:       update.Total,
		CurrentFile: update.Current,
	}
}
