package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/operations"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
	"github.com/petergi/ebook-mechanic-lib/pkg/ebmlib"
)

// ReportModel displays validation or repair results
type ReportModel struct {
	report          *ebmlib.ValidationReport
	repairResult    *ebmlib.RepairResult
	repairReport    *ebmlib.ValidationReport
	batchResult     *operations.BatchResult
	reportType      string // "validation", "repair", "batch"
	width           int
	height          int
	viewportTop     int
	viewportSize    int
	showErrors      bool
	showWarnings    bool
	showInfo        bool
	selectedFilter  int    // 0: all, 1: errors, 2: warnings, 3: info
	savedReportPath string // Path where report was saved
	saveError       error  // Error from last save attempt
	saveStatusToken int
	saveStatusShow  bool
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
		viewportSize: calculateViewportSize(height, "validation"),
		showErrors:   true,
		showWarnings: true,
		showInfo:     true,
	}
}

// NewBatchReportModel creates a new report model for batch results
func NewBatchReportModel(result *operations.BatchResult, width, height int) ReportModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	selectedFilter := 0 // Default to Invalid
	if len(result.Invalid) == 0 && len(result.Errored) > 0 {
		selectedFilter = 1 // Default to Errored if only system errors
	} else if len(result.Invalid) == 0 && len(result.Errored) == 0 {
		selectedFilter = 2 // Default to Valid if all good
	}

	return ReportModel{
		batchResult:    result,
		reportType:     "batch",
		width:          width,
		height:         height,
		viewportSize:   calculateViewportSize(height, "batch"),
		selectedFilter: selectedFilter,
		showErrors:     true,
		showWarnings:   true,
		showInfo:       true,
	}
}

// NewRepairReportModel creates a new report model for repair results
func NewRepairReportModel(result *ebmlib.RepairResult, width, height int) ReportModel {
	return NewRepairReportModelWithValidation(result, nil, width, height)
}

// NewRepairReportModelWithValidation creates a new report model for repair results with optional validation.
func NewRepairReportModelWithValidation(result *ebmlib.RepairResult, report *ebmlib.ValidationReport, width, height int) ReportModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	return ReportModel{
		repairResult: result,
		repairReport: report,
		reportType:   "repair",
		width:        width,
		height:       height,
		viewportSize: calculateViewportSize(height, "repair"),
		showErrors:   true,
		showWarnings: true,
		showInfo:     true,
	}
}

func calculateViewportSize(height int, reportType string) int {
	var offset int
	switch reportType {
	case "validation":
		// Overhead: Title(2) + Path(3) + Gap(1) + Status(5) + Gap(1) + Summary(9) + Filters(1) + Border(2) + Help(5) = ~29
		offset = 32
	case "batch":
		// Overhead: Title(2) + Status(5) + Gap(1) + Summary(9) + Border(2) + Help(5) = ~24
		offset = 28
	case "repair":
		// Overhead: Title(2) + Status(5) + Border(2) + Help(5) = ~14 (plus actions content)
		// Since repair doesn't use a viewport for actions currently, this just limits max size if we were to use it.
		// For now we keep it safer.
		offset = 20
	default:
		offset = 12
	}

	size := height - offset
	if size < 5 {
		return 5 // Minimum viewport size
	}
	return size
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
		m.viewportSize = calculateViewportSize(m.height, m.reportType)
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
			if m.reportType == "batch" {
				m.selectedFilter = 0 // Invalid
				m.viewportTop = 0
			} else {
				m.selectedFilter = 0 // Show all
				m.showErrors = true
				m.showWarnings = true
				m.showInfo = true
				m.viewportTop = 0
			}

		case "2":
			if m.reportType == "batch" {
				m.selectedFilter = 1 // Errored
				m.viewportTop = 0
			} else {
				m.selectedFilter = 1 // Show errors only
				m.showErrors = true
				m.showWarnings = false
				m.showInfo = false
				m.viewportTop = 0
			}

		case "3":
			if m.reportType == "batch" {
				m.selectedFilter = 2 // Valid
				m.viewportTop = 0
			} else {
				m.selectedFilter = 2 // Show warnings only
				m.showErrors = false
				m.showWarnings = true
				m.showInfo = false
				m.viewportTop = 0
			}

		case "4":
			if m.reportType == "batch" {
				m.selectedFilter = 3 // All
				m.viewportTop = 0
			} else {
				m.selectedFilter = 3 // Show info only
				m.showErrors = false
				m.showWarnings = false
				m.showInfo = true
				m.viewportTop = 0
			}

		case "s":
			// Save report to file
			return m, func() tea.Msg {
				path, err := m.saveReport()
				if err != nil {
					return ReportSaveMsg{Error: err}
				}
				return ReportSaveMsg{Path: path}
			}

		case "enter", "esc":
			// Return to menu
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}

		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case ReportSaveMsg:
		if msg.Error != nil {
			// Store error to display in UI
			m.saveError = msg.Error
			m.savedReportPath = "" // Clear any previous success
		} else {
			m.savedReportPath = msg.Path
			m.saveError = nil // Clear any previous error
		}
		m.saveStatusToken++
		token := m.saveStatusToken
		m.saveStatusShow = true
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return ReportSaveTimeoutMsg{Token: token}
		})

	case ReportSaveTimeoutMsg:
		if msg.Token == m.saveStatusToken {
			m.saveStatusShow = false
		}
		return m, nil
	}

	return m, nil
}

// View renders the report
func (m ReportModel) View() string {
	switch m.reportType {
	case "repair":
		return m.renderRepairReport()
	case "batch":
		return m.renderBatchReport()
	default:
		return m.renderValidationReport()
	}
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
		Render(m.makeClickable(m.report.FilePath))

	// Success or error message for save operation
	var savedBox string
	if m.saveStatusShow && m.savedReportPath != "" {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorSuccess).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Report saved to: %s", styles.IconCheck, m.makeClickable(m.savedReportPath)))
	} else if m.saveStatusShow && m.saveError != nil {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorError).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Error saving report: %v", styles.IconCross, m.saveError))
	}

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
				styles.RenderKeyBinding("s", "save report") + "  " +
				styles.RenderKeyBinding("enter", "continue"),
		)

	saveStatusLine := m.renderSaveStatusLine()

	// Combine all parts
	var parts []string
	parts = append(parts, title, pathBox)
	if savedBox != "" {
		parts = append(parts, "", savedBox)
	}
	parts = append(parts, "", statusBox, "", summaryBox, filters, issuesBox)
	if saveStatusLine != "" {
		parts = append(parts, "", saveStatusLine)
	}
	parts = append(parts, helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

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
			statusText += "\n\n" + repairArtifactTitle(m.repairResult.BackupPath) + "\n" + m.makeClickable(m.repairResult.BackupPath)
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

	// Success or error message for save operation
	var savedBox string
	if m.saveStatusShow && m.savedReportPath != "" {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorSuccess).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Report saved to: %s", styles.IconCheck, m.makeClickable(m.savedReportPath)))
	} else if m.saveStatusShow && m.saveError != nil {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorError).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Error saving report: %v", styles.IconCross, m.saveError))
	}

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

	validationBox := m.renderRepairValidationBox()

	// Scrollable details view
	var detailParts []string
	detailParts = append(detailParts, statusBox)
	if validationBox != "" {
		detailParts = append(detailParts, "", validationBox)
	}
	if actionsBox != "" {
		detailParts = append(detailParts, "", actionsBox)
	}

	detailsContent := strings.Join(detailParts, "\n")
	lines := strings.Split(detailsContent, "\n")
	start := m.viewportTop
	end := m.viewportTop + m.viewportSize
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		m.viewportTop = len(lines) - m.viewportSize
		if m.viewportTop < 0 {
			m.viewportTop = 0
		}
		start = m.viewportTop
	}
	visibleLines := lines[start:end]
	detailsContent = strings.Join(visibleLines, "\n")

	if len(lines) > m.viewportSize {
		if m.viewportTop > 0 {
			detailsContent = styles.MutedStyle.Render("↑ more ↑") + "\n" + detailsContent
		}
		if end < len(lines) {
			detailsContent = detailsContent + "\n" + styles.MutedStyle.Render("↓ more ↓")
		}
	}

	detailsBox := styles.BorderStyle.
		Width(m.width - 8).
		Height(m.viewportSize + 2).
		Render(detailsContent)

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 2).
		Width(m.width - 8).
		Render(
			styles.RenderKeyBinding("s", "save report") + "  " +
				styles.RenderKeyBinding("enter", "continue"),
		)

	saveStatusLine := m.renderSaveStatusLine()

	// Combine all parts
	var parts []string
	parts = append(parts, title)
	if savedBox != "" {
		parts = append(parts, "", savedBox)
	}
	parts = append(parts, "", detailsBox)
	if saveStatusLine != "" {
		parts = append(parts, "", saveStatusLine)
	}
	parts = append(parts, "", helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		content,
	)
}

func (m ReportModel) renderBatchReport() string {
	if m.batchResult == nil {
		return styles.RenderError("No batch result available")
	}

	title := styles.RenderTitle("📦 Batch Report")

	// Status in highlighted box - operation-aware
	var statusText string
	var statusColor lipgloss.AdaptiveColor

	if m.batchResult.Operation == "repair" {
		// Repair operation status
		if m.batchResult.RepairsSucceeded == m.batchResult.RepairsAttempted && len(m.batchResult.Errored) == 0 {
			statusText = fmt.Sprintf("%s  All repairs successful! (%d/%d)", styles.IconCheck, m.batchResult.RepairsSucceeded, m.batchResult.RepairsAttempted)
			statusColor = styles.ColorSuccess
		} else {
			statusText = fmt.Sprintf("%s  %d/%d repairs succeeded, %d failed, %d system error(s)",
				styles.IconCross,
				m.batchResult.RepairsSucceeded,
				m.batchResult.RepairsAttempted,
				m.batchResult.RepairsAttempted-m.batchResult.RepairsSucceeded,
				len(m.batchResult.Errored))
			statusColor = styles.ColorError
		}
	} else {
		// Validation operation status
		if len(m.batchResult.Invalid) == 0 && len(m.batchResult.Errored) == 0 {
			statusText = styles.IconCheck + "  All files processed successfully!"
			statusColor = styles.ColorSuccess
		} else {
			statusText = fmt.Sprintf("%s  Found %d invalid and %d system error(s)", styles.IconCross, len(m.batchResult.Invalid), len(m.batchResult.Errored))
			statusColor = styles.ColorError
		}
	}

	statusBox := lipgloss.NewStyle().
		Foreground(statusColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(1, 2).
		Width(m.width - 8).
		Render(statusText)

	// Success or error message for save operation
	var savedBox string
	if m.saveStatusShow && m.savedReportPath != "" {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorSuccess).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Report saved to: %s", styles.IconCheck, m.makeClickable(m.savedReportPath)))
	} else if m.saveStatusShow && m.saveError != nil {
		savedBox = lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorError).
			Padding(0, 1).
			Width(m.width - 8).
			Render(fmt.Sprintf("%s Error saving report: %v", styles.IconCross, m.saveError))
	}

	// Summary - different for repair vs validation
	headers := []string{"Metric", "Value"}
	var rows [][]string

	if m.batchResult.Operation == "repair" {
		// Repair operation - show repair-specific metrics
		rows = [][]string{
			{"Total Files", fmt.Sprintf("%d", m.batchResult.Total)},
			{"Repairs Attempted", fmt.Sprintf("%d", m.batchResult.RepairsAttempted)},
			{"Repairs Succeeded", fmt.Sprintf("%d", m.batchResult.RepairsSucceeded)},
			{"Repairs Failed", fmt.Sprintf("%d", m.batchResult.RepairsAttempted-m.batchResult.RepairsSucceeded)},
			{"System Errors", fmt.Sprintf("%d", len(m.batchResult.Errored))},
			{"Duration", m.batchResult.Duration.Round(time.Millisecond).String()},
		}
	} else {
		// Validation operation - show validation metrics
		rows = [][]string{
			{"Total Files", fmt.Sprintf("%d", m.batchResult.Total)},
			{"Valid", fmt.Sprintf("%d", len(m.batchResult.Valid))},
			{"Invalid", fmt.Sprintf("%d", len(m.batchResult.Invalid))},
			{"System Errors", fmt.Sprintf("%d", len(m.batchResult.Errored))},
			{"Duration", m.batchResult.Duration.Round(time.Millisecond).String()},
		}
	}
	summaryContent := styles.RenderTable(headers, rows)
	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(m.width - 8).
		Render(summaryContent)

	// Filters
	filters := m.renderBatchFilters()

	// File list based on filter
	var listContent string
	var items []string

	// Filter: 0=Invalid, 1=Errored, 2=Valid, 3=All
	switch m.selectedFilter {
	case 0: // Invalid
		for _, r := range m.batchResult.Invalid {
			items = append(items, m.formatBatchItem(r, "invalid"))
		}
	case 1: // Errored
		for _, r := range m.batchResult.Errored {
			items = append(items, m.formatBatchItem(r, "errored"))
		}
	case 2: // Valid
		for _, r := range m.batchResult.Valid {
			items = append(items, m.formatBatchItem(r, "valid"))
		}
	case 3: // All
		for _, r := range m.batchResult.Errored {
			items = append(items, m.formatBatchItem(r, "errored"))
		}
		for _, r := range m.batchResult.Invalid {
			items = append(items, m.formatBatchItem(r, "invalid"))
		}
		for _, r := range m.batchResult.Valid {
			items = append(items, m.formatBatchItem(r, "valid"))
		}
	}

	if len(items) > 0 {
		listContent = strings.Join(items, "\n")
	} else {
		listContent = styles.MutedStyle.Render("No items to display")
	}

	// Apply viewport scrolling
	lines := strings.Split(listContent, "\n")
	start := m.viewportTop
	end := m.viewportTop + m.viewportSize
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		m.viewportTop = len(lines) - m.viewportSize
		if m.viewportTop < 0 {
			m.viewportTop = 0
		}
		start = m.viewportTop
	}
	visibleLines := lines[start:end]
	listContent = strings.Join(visibleLines, "\n")

	// Scroll indicators
	if len(lines) > m.viewportSize {
		if m.viewportTop > 0 {
			listContent = styles.MutedStyle.Render("↑ more ↑") + "\n" + listContent
		}
		if end < len(lines) {
			listContent = listContent + "\n" + styles.MutedStyle.Render("↓ more ↓")
		}
	}

	listBox := styles.BorderStyle.
		Width(m.width - 8).
		Height(m.viewportSize + 2).
		Render(listContent)

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
				styles.RenderKeyBinding("s", "save report") + "  " +
				styles.RenderKeyBinding("enter", "continue"),
		)

	saveStatusLine := m.renderSaveStatusLine()

	// Combine all parts
	var parts []string
	parts = append(parts, title)
	if savedBox != "" {
		parts = append(parts, "", savedBox)
	}
	parts = append(parts, "", statusBox, "", summaryBox, filters, listBox)
	if saveStatusLine != "" {
		parts = append(parts, "", saveStatusLine)
	}
	parts = append(parts, helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		content,
	)
}

func (m ReportModel) renderBatchFilters() string {
	tabs := []string{
		"1: Invalid",
		"2: Errored",
		"3: Valid",
		"4: All",
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

func (m ReportModel) formatBatchItem(r operations.Result, category string) string {
	var icon string
	var style lipgloss.Style

	switch category {
	case "valid":
		icon = styles.IconCheck
		style = styles.SuccessStyle
	case "invalid":
		icon = styles.IconCross
		style = styles.ErrorStyle
	case "errored":
		icon = "⚠"
		style = styles.ErrorStyle
	}

	msg := fmt.Sprintf("%s %s", icon, m.makeClickable(filepath.Base(r.FilePath)))

	switch category {
	case "invalid":
		if r.Report != nil && !r.Report.IsValid {
			msg += fmt.Sprintf(": %d errors", r.Report.ErrorCount())
		} else if r.Repair != nil && !r.Repair.Success {
			if r.Repair.Error != nil {
				msg += ": " + r.Repair.Error.Error()
			} else {
				msg += ": repair failed"
			}
		}
	case "errored":
		if r.Error != nil {
			msg += ": " + r.Error.Error()
		}
	case "valid":
		if r.Report != nil {
			msg += " (Valid)"
		} else if r.Repair != nil {
			msg += " (Repaired)"
		}
	}

	return style.Render(msg)
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

func (m ReportModel) makeClickable(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	fileURL := "file://" + absPath
	// OSC 8 hyperlink format: \x1b]8;;<URL>\x1b\\<TEXT>\x1b]8;;\x1b\\
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", fileURL, path)
}

func (m ReportModel) renderSaveStatusLine() string {
	if m.saveStatusShow && m.savedReportPath != "" {
		return styles.SuccessStyle.Render(fmt.Sprintf("%s Saved report: %s", styles.IconCheck, m.makeClickable(m.savedReportPath)))
	}
	if m.saveStatusShow && m.saveError != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("%s Error saving report: %v", styles.IconCross, m.saveError))
	}
	return ""
}

func (m ReportModel) renderRepairValidationBox() string {
	if m.repairReport == nil {
		return ""
	}

	status := "Invalid"
	statusStyle := styles.ErrorStyle
	if m.repairReport.IsValid {
		status = "Valid"
		statusStyle = styles.SuccessStyle
	}

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Status", statusStyle.Render(status)},
		{"Errors", fmt.Sprintf("%d", m.repairReport.ErrorCount())},
		{"Warnings", fmt.Sprintf("%d", m.repairReport.WarningCount())},
		{"Info", fmt.Sprintf("%d", m.repairReport.InfoCount())},
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.SubtitleStyle.Render("Post-repair Validation"),
		"",
		styles.RenderTable(headers, rows),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(m.width - 8).
		Render(content)
}

func repairArtifactTitle(path string) string {
	if isRepairedPath(path) {
		return "Repaired file created at:"
	}
	return "Backup created at:"
}

func repairArtifactLabel(path string) string {
	if isRepairedPath(path) {
		return "Repaired File"
	}
	return "Backup"
}

func isRepairedPath(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return strings.HasSuffix(name, "_repaired")
}

func (m ReportModel) saveReport() (string, error) {
	// Create reports directory
	reportDir := "reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", err
	}

	// Generate filename based on timestamp
	timestamp := time.Now().Format("20060102-150405")
	var filename string
	var content string

	switch m.reportType {
	case "repair":
		filename = fmt.Sprintf("repair-%s.txt", timestamp)
		// Basic text formatting for repair
		var b strings.Builder
		b.WriteString("=== Repair Report ===\n\n")
		if m.repairResult.Success {
			b.WriteString("Status: Success\n")
		} else {
			b.WriteString("Status: Failed\n")
			if m.repairResult.Error != nil {
				b.WriteString(fmt.Sprintf("Error: %v\n", m.repairResult.Error))
			}
		}
		if m.repairResult.BackupPath != "" {
			b.WriteString(fmt.Sprintf("%s: %s\n", repairArtifactLabel(m.repairResult.BackupPath), m.repairResult.BackupPath))
		}
		b.WriteString("\nActions Applied:\n")
		for _, action := range m.repairResult.ActionsApplied {
			b.WriteString(fmt.Sprintf("- %s\n", action.Description))
		}
		content = b.String()
	case "batch":
		filename = fmt.Sprintf("batch-%s.txt", timestamp)
		var b strings.Builder

		if m.batchResult.Operation == "repair" {
			b.WriteString("=== Batch Repair Report ===\n\n")
			b.WriteString(fmt.Sprintf("Total Files: %d\n", m.batchResult.Total))
			b.WriteString(fmt.Sprintf("Repairs Attempted: %d\n", m.batchResult.RepairsAttempted))
			b.WriteString(fmt.Sprintf("Repairs Succeeded: %d\n", m.batchResult.RepairsSucceeded))
			b.WriteString(fmt.Sprintf("Repairs Failed: %d\n", m.batchResult.RepairsAttempted-m.batchResult.RepairsSucceeded))
			b.WriteString(fmt.Sprintf("System Errors: %d\n", len(m.batchResult.Errored)))
			b.WriteString(fmt.Sprintf("Duration: %v\n\n", m.batchResult.Duration))
		} else {
			b.WriteString("=== Batch Validation Report ===\n\n")
			b.WriteString(fmt.Sprintf("Total Files: %d\n", m.batchResult.Total))
			b.WriteString(fmt.Sprintf("Valid: %d\n", len(m.batchResult.Valid)))
			b.WriteString(fmt.Sprintf("Invalid: %d\n", len(m.batchResult.Invalid)))
			b.WriteString(fmt.Sprintf("System Errors: %d\n", len(m.batchResult.Errored)))
			b.WriteString(fmt.Sprintf("Duration: %v\n\n", m.batchResult.Duration))
		}

		if len(m.batchResult.Errored) > 0 {
			b.WriteString("System Errors:\n")
			for _, r := range m.batchResult.Errored {
				b.WriteString(fmt.Sprintf("- %s: %v\n", filepath.Base(r.FilePath), r.Error))
			}
			b.WriteString("\n")
		}

		if len(m.batchResult.Invalid) > 0 {
			b.WriteString("Invalid Files:\n")
			for _, r := range m.batchResult.Invalid {
				b.WriteString(fmt.Sprintf("- %s: ", filepath.Base(r.FilePath)))
				if r.Report != nil && !r.Report.IsValid {
					b.WriteString(fmt.Sprintf("%d errors\n", r.Report.ErrorCount()))
				} else if r.Repair != nil && !r.Repair.Success {
					b.WriteString("repair failed\n")
				} else {
					b.WriteString("unknown error\n")
				}
			}
			b.WriteString("\n")
		}

		if len(m.batchResult.Valid) > 0 {
			b.WriteString("Valid Files:\n")
			for _, r := range m.batchResult.Valid {
				b.WriteString(fmt.Sprintf("- %s\n", filepath.Base(r.FilePath)))
			}
		}
		content = b.String()
	default:
		filename = fmt.Sprintf("validation-%s.txt", timestamp)
		// We can reuse the CLI TextFormatter logic if we want,
		// but for now simple text output
		var b strings.Builder
		b.WriteString("=== Validation Report ===\n\n")
		b.WriteString(fmt.Sprintf("File: %s\n", m.report.FilePath))
		b.WriteString(fmt.Sprintf("Status: %v\n\n", m.report.IsValid))
		b.WriteString("Summary:\n")
		b.WriteString(fmt.Sprintf("- Errors: %d\n", m.report.ErrorCount()))
		b.WriteString(fmt.Sprintf("- Warnings: %d\n", m.report.WarningCount()))
		b.WriteString(fmt.Sprintf("- Info: %d\n\n", m.report.InfoCount()))

		if len(m.report.Errors) > 0 {
			b.WriteString("Errors:\n")
			for _, err := range m.report.Errors {
				b.WriteString(fmt.Sprintf("- [%s] %s\n", err.Code, err.Message))
			}
		}
		content = b.String()
	}

	path := filepath.Join(reportDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}

	return absPath, nil
}

// ReportSaveMsg is sent when a report is saved
type ReportSaveMsg struct {
	Path  string
	Error error
}

// ReportSaveTimeoutMsg hides save status after a delay.
type ReportSaveTimeoutMsg struct {
	Token int
}
