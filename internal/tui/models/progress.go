package models

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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
	progress    progress.Model
}

// NewProgressModel creates a new progress model
func NewProgressModel(operation string, filePath string, total int, width, height int) ProgressModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width-20), // Adjust width to fit
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		operation: operation,
		filePath:  filePath,
		total:     total,
		startTime: time.Now(),
		width:     width,
		height:    height,
		progress:  p,
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
		m.progress.Width = msg.Width - 20
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "enter":
				// View results
				return m, func() tea.Msg {
					return ViewReportMsg{Result: m.result}
				}
			case "esc":
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
		if m.total > 0 {
			cmd := m.progress.SetPercent(float64(m.current) / float64(m.total))
			return m, cmd
		}
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case OperationDoneMsg:
		m.done = true
		m.result = msg.Result
		if m.total > 0 {
			cmd := m.progress.SetPercent(1.0)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

// View renders the progress display
func (m ProgressModel) View() string {
	if m.done {
		return m.renderDone()
	}

	// Title with animated spinner
	spinnerChar := string(styles.IconSpinner[m.spinner])
	title := styles.RenderTitle(spinnerChar + "  " + m.operation)

	// Status box with operation-specific information
	statusText := "Processing..."
	var operationIcon string

	// Show operation-specific status
	switch m.operation {
	case "Batch Repair":
		operationIcon = "🔧"
		if m.currentFile != "" {
			statusText = fmt.Sprintf("%s Attempting repair: %s", operationIcon, m.currentFile)
		} else {
			statusText = fmt.Sprintf("%s Repairing ebooks...", operationIcon)
		}
	case "Batch Validation":
		operationIcon = "🔍"
		if m.currentFile != "" {
			statusText = fmt.Sprintf("%s Validating: %s", operationIcon, m.currentFile)
		} else {
			statusText = fmt.Sprintf("%s Validating ebooks...", operationIcon)
		}
	default:
		if m.currentFile != "" {
			statusText = "Current: " + m.currentFile
		} else if m.total == 1 && m.filePath != "" {
			statusText = "File: " + m.filePath
		}
	}

	statusBox := lipgloss.NewStyle().
		Foreground(styles.ColorInfo).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorInfo).
		Padding(1, 2).
		Width(m.width - 4).
		Render(statusText)

	// Progress bar (for batch operations) in a box
	var progressBox string
	if m.total > 1 {
		percentage := 0.0
		if m.total > 0 {
			percentage = float64(m.current) / float64(m.total) * 100
		}

		progressText := fmt.Sprintf("Completed: %d / %d (%.0f%%)", m.current, m.total, percentage)

		progressContent := lipgloss.JoinVertical(
			lipgloss.Left,
			progressText,
			"",
			m.progress.View(),
		)

		progressBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.ColorPrimary).
			Padding(1, 2).
			Width(m.width - 4).
			Render(progressContent)
	}

	// Time info in a subtle status bar
	elapsed := time.Since(m.startTime)
	timeBar := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Width(m.width - 4).
		Render("Elapsed: " + elapsed.Round(time.Second).String())

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(m.width - 4).
		Render(styles.RenderKeyBinding("ctrl+c", "cancel operation"))

	// Combine all parts
	var parts []string
	parts = append(parts, title, "", statusBox)
	if progressBox != "" {
		parts = append(parts, "", progressBox)
	}
	parts = append(parts, "", timeBar, helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderDone renders the completion screen
func (m ProgressModel) renderDone() string {
	elapsed := time.Since(m.startTime)

	title := styles.RenderTitle(styles.IconCheck + "  Operation Complete")

	// Success summary in a highlighted box
	summaryText := m.operation + " completed successfully!"
	if m.total > 1 {
		summaryText = fmt.Sprintf("%s completed successfully!\n\nProcessed: %d files", m.operation, m.total)
	}

	summaryBox := lipgloss.NewStyle().
		Foreground(styles.ColorSuccess).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSuccess).
		Padding(1, 2).
		Width(60).
		Render(summaryText)

	// Time info
	timeBar := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Width(60).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				"Completed in: ",
				lipgloss.NewStyle().Foreground(styles.ColorInfo).Render(elapsed.Round(time.Millisecond).String()),
			),
		)

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(60).
		Render(styles.RenderKeyBinding("enter", "continue to results"))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		summaryBox,
		"",
		timeBar,
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

// ViewReportMsg signals request to view the report
type ViewReportMsg struct {
	Result interface{}
}

// ConvertBatchProgress converts batch progress to progress message
func ConvertBatchProgress(update operations.ProgressUpdate) ProgressUpdateMsg {
	return ProgressUpdateMsg{
		Current:     update.Completed,
		Total:       update.Total,
		CurrentFile: update.Current,
	}
}
