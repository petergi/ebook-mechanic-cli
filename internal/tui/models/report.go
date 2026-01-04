package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// ReportModel displays validation or repair results
type ReportModel struct {
	report         *ebmlib.ValidationReport
	repairResult   *ebmlib.RepairResult
	reportType     string // "validation" or "repair"
	width          int
	height         int
	viewportTop    int
	viewportSize   int
	showErrors     bool
	showWarnings   bool
	showInfo       bool
	selectedFilter int // 0: all, 1: errors, 2: warnings, 3: info
}

// NewReportModel creates a new report model for validation results
func NewReportModel(report *ebmlib.ValidationReport, width, height int) ReportModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	return ReportModel{
		report:       report,
		reportType:   "validation",
		width:        width,
		height:       height,
		viewportSize: height - 12,
		showErrors:   true,
		showWarnings: true,
		showInfo:     true,
	}
}

// NewRepairReportModel creates a new report model for repair results
func NewRepairReportModel(result *ebmlib.RepairResult, width, height int) ReportModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	return ReportModel{
		repairResult: result,
		reportType:   "repair",
		width:        width,
		height:       height,
		viewportSize: height - 12,
		showErrors:   true,
		showWarnings: true,
		showInfo:     true,
	}
}

// Init initializes the model
func (m ReportModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state
func (m ReportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewportSize = m.height - 12
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.viewportTop > 0 {
				m.viewportTop--
			}

		case "down", "j":
			m.viewportTop++

		case "1":
			m.selectedFilter = 0 // Show all
			m.showErrors = true
			m.showWarnings = true
			m.showInfo = true
			m.viewportTop = 0

		case "2":
			m.selectedFilter = 1 // Show errors only
			m.showErrors = true
			m.showWarnings = false
			m.showInfo = false
			m.viewportTop = 0

		case "3":
			m.selectedFilter = 2 // Show warnings only
			m.showErrors = false
			m.showWarnings = true
			m.showInfo = false
			m.viewportTop = 0

		case "4":
			m.selectedFilter = 3 // Show info only
			m.showErrors = false
			m.showWarnings = false
			m.showInfo = true
			m.viewportTop = 0

		case "enter", "esc":
			// Return to menu
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the report
func (m ReportModel) View() string {
	if m.reportType == "repair" {
		return m.renderRepairReport()
	}
	return m.renderValidationReport()
}

// renderValidationReport renders a validation report
func (m ReportModel) renderValidationReport() string {
	if m.report == nil {
		return styles.RenderError("No report available")
	}

	// Title
	title := styles.RenderTitle("📊 Validation Report")

	// File path in subtle header
	pathBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Width(m.width - 8).
		Render(m.report.FilePath)

	// Overall status in highlighted box
	var statusText string
	var statusColor lipgloss.AdaptiveColor
	if m.report.IsValid {
		statusText = styles.IconCheck + "  File is valid!"
		statusColor = styles.ColorSuccess
	} else {
		statusText = styles.IconCross + "  File has errors"
		statusColor = styles.ColorError
	}

	statusBox := lipgloss.NewStyle().
		Foreground(statusColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(1, 2).
		Width(m.width - 8).
		Render(statusText)

	// Summary table in bordered box
	summaryContent := m.renderSummary()
	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(m.width - 8).
		Render(summaryContent)

	// Filter tabs
	filters := m.renderFilters()

	// Issues list in bordered scrollable box
	issuesContent := m.renderIssues()
	issuesBox := styles.BorderStyle.
		Width(m.width - 8).
		Height(m.viewportSize + 2).
		Render(issuesContent)

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(m.width - 8).
		Render(
			styles.RenderKeyBinding("1-4", "filter") + "  " +
				styles.RenderKeyBinding("↑/↓", "scroll") + "  " +
				styles.RenderKeyBinding("enter", "continue"),
		)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		pathBox,
		"",
		statusBox,
		"",
		summaryBox,
		filters,
		issuesBox,
		helpBox,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		content,
	)
}

// renderRepairReport renders a repair result report
func (m ReportModel) renderRepairReport() string {
	if m.repairResult == nil {
		return styles.RenderError("No repair result available")
	}

	title := styles.RenderTitle("🔧 Repair Report")

	// Status in highlighted box
	var statusText string
	var statusColor lipgloss.AdaptiveColor
	if m.repairResult.Success {
		statusText = styles.IconCheck + "  Repair successful!"
		statusColor = styles.ColorSuccess
		if m.repairResult.BackupPath != "" {
			statusText += "\n\n" + "Backup created at:\n" + m.repairResult.BackupPath
		}
	} else {
		statusText = styles.IconCross + "  Repair failed"
		statusColor = styles.ColorError
		if m.repairResult.Error != nil {
			statusText += "\n\n" + m.repairResult.Error.Error()
		}
	}

	statusBox := lipgloss.NewStyle().
		Foreground(statusColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(1, 2).
		Width(m.width - 8).
		Render(statusText)

	// Actions applied in a bordered box
	var actionsBox string
	if len(m.repairResult.ActionsApplied) > 0 {
		var actionsList string
		for _, action := range m.repairResult.ActionsApplied {
			actionsList += styles.IconCheck + " " + action.Description + "\n"
		}

		actionsBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.ColorPrimary).
			Padding(1, 2).
			Width(m.width - 8).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					styles.SubtitleStyle.Render("Actions Applied:"),
					"",
					actionsList,
				),
			)
	}

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(m.width - 8).
		Render(styles.RenderKeyBinding("enter", "continue"))

	// Combine all parts
	var parts []string
	parts = append(parts, title, statusBox)
	if actionsBox != "" {
		parts = append(parts, "", actionsBox)
	}
	parts = append(parts, helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderSummary renders the issue summary
func (m ReportModel) renderSummary() string {
	headers := []string{"Type", "Count"}
	rows := [][]string{
		{"Errors", fmt.Sprintf("%d", m.report.ErrorCount())},
		{"Warnings", fmt.Sprintf("%d", m.report.WarningCount())},
		{"Info", fmt.Sprintf("%d", m.report.InfoCount())},
	}

	return styles.RenderTable(headers, rows)
}

// renderFilters renders filter tabs
func (m ReportModel) renderFilters() string {
	tabs := []string{
		"1: All",
		"2: Errors",
		"3: Warnings",
		"4: Info",
	}

	var rendered []string
	for i, tab := range tabs {
		if i == m.selectedFilter {
			rendered = append(rendered, styles.SelectedListItemStyle.Render(tab))
		} else {
			rendered = append(rendered, styles.MutedStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// renderIssues renders the list of issues with scrolling
func (m ReportModel) renderIssues() string {
	var issues []string

	if m.showErrors {
		for _, err := range m.report.Errors {
			issues = append(issues, m.formatIssue(err, "error"))
		}
	}

	if m.showWarnings {
		for _, warn := range m.report.Warnings {
			issues = append(issues, m.formatIssue(warn, "warning"))
		}
	}

	if m.showInfo {
		for _, info := range m.report.Info {
			issues = append(issues, m.formatIssue(info, "info"))
		}
	}

	if len(issues) == 0 {
		return styles.MutedStyle.Render("No issues to display")
	}

	// Apply viewport scrolling
	start := m.viewportTop
	end := m.viewportTop + m.viewportSize
	if end > len(issues) {
		end = len(issues)
	}
	if start >= len(issues) {
		m.viewportTop = len(issues) - m.viewportSize
		if m.viewportTop < 0 {
			m.viewportTop = 0
		}
		start = m.viewportTop
	}

	visibleIssues := issues[start:end]

	content := strings.Join(visibleIssues, "\n")

	// Scroll indicators
	if len(issues) > m.viewportSize {
		if m.viewportTop > 0 {
			content = styles.MutedStyle.Render("↑ more ↑") + "\n" + content
		}
		if end < len(issues) {
			content = content + "\n" + styles.MutedStyle.Render("↓ more ↓")
		}
	}

	return content
}

// formatIssue formats a single issue for display
func (m ReportModel) formatIssue(issue ebmlib.ValidationError, issueType string) string {
	var icon, style string
	switch issueType {
	case "error":
		icon = styles.IconCross
		style = "error"
	case "warning":
		icon = styles.IconWarning
		style = "warning"
	case "info":
		icon = styles.IconInfo
		style = "info"
	}

	header := fmt.Sprintf("%s [%s] %s", icon, issue.Code, issue.Message)

	var location string
	if issue.Location != nil {
		location = fmt.Sprintf("  Location: %s", issue.Location.File)
		if issue.Location.Line > 0 {
			location += fmt.Sprintf(":%d", issue.Location.Line)
		}
	}

	formatted := header
	if location != "" {
		formatted += "\n" + styles.MutedStyle.Render(location)
	}

	switch style {
	case "error":
		return styles.ErrorStyle.Render(formatted)
	case "warning":
		return styles.WarningStyle.Render(formatted)
	case "info":
		return styles.InfoStyle.Render(formatted)
	}

	return formatted
}
