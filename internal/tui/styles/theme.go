package styles

import "github.com/charmbracelet/lipgloss"

// Semantic Colors
var (
	ColorError   = lipgloss.AdaptiveColor{Light: "#E06C75", Dark: "#E06C75"} // Red
	ColorWarning = lipgloss.AdaptiveColor{Light: "#E5C07B", Dark: "#E5C07B"} // Yellow
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#98C379", Dark: "#98C379"} // Green
	ColorInfo    = lipgloss.AdaptiveColor{Light: "#61AFEF", Dark: "#61AFEF"} // Blue
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#56B6C2", Dark: "#56B6C2"} // Cyan
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#5C6370", Dark: "#5C6370"} // Gray
	ColorBg      = lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#282C34"} // Background
	ColorFg      = lipgloss.AdaptiveColor{Light: "#282C34", Dark: "#ABB2BF"} // Foreground
)

// Base Styles
var (
	// BaseStyle is the foundation for all text styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// ErrorStyle for error messages and critical issues
	ErrorStyle = BaseStyle.
			Foreground(ColorError).
			Bold(true)

	// WarningStyle for warnings and non-critical issues
	WarningStyle = BaseStyle.
			Foreground(ColorWarning)

	// SuccessStyle for success messages
	SuccessStyle = BaseStyle.
			Foreground(ColorSuccess).
			Bold(true)

	// InfoStyle for informational messages
	InfoStyle = BaseStyle.
			Foreground(ColorInfo)

	// MutedStyle for less important text
	MutedStyle = BaseStyle.
			Foreground(ColorMuted)
)

// Component Styles
var (
	// TitleStyle for main headings
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1).
			Padding(0, 1)

	// SubtitleStyle for secondary headings
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			MarginBottom(1).
			Padding(0, 1)

	// BorderStyle for bordered boxes
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// FocusedBorderStyle for focused/selected bordered boxes
	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(ColorSuccess).
				Padding(1, 2)

	// ListItemStyle for list items
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// SelectedListItemStyle for selected list items
	SelectedListItemStyle = ListItemStyle.
				Foreground(ColorPrimary).
				Bold(true)

	// KeyStyle for keyboard shortcut keys
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// DescStyle for descriptions
	DescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// HelpStyle for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 0)
)

// Layout Styles
var (
	// DocStyle for the main document container
	DocStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// DialogBoxStyle for modal dialogs
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(60)

	// ProgressBarStyle for progress bars
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	// ProgressBarEmptyStyle for empty progress bar sections
	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
)

// Table Styles
var (
	// TableHeaderStyle for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorMuted)

	// TableRowStyle for table rows
	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// TableCellStyle for table cells
	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// Status Styles
var (
	// ValidStyle for valid status
	ValidStyle = SuccessStyle

	// InvalidStyle for invalid status
	InvalidStyle = ErrorStyle

	// ProcessingStyle for processing status
	ProcessingStyle = InfoStyle

	// PendingStyle for pending status
	PendingStyle = MutedStyle
)

// Icon strings (using Unicode symbols)
const (
	IconCheck   = "✓"
	IconCross   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconArrow   = "→"
	IconBullet  = "•"
	IconSpinner = "⣾⣽⣻⢿⡿⣟⣯⣷" // Animation frames for spinner
)

// Helper functions

// RenderTitle renders a styled title
func RenderTitle(text string) string {
	return TitleStyle.Render(text)
}

// RenderSubtitle renders a styled subtitle
func RenderSubtitle(text string) string {
	return SubtitleStyle.Render(text)
}

// RenderError renders an error message with icon
func RenderError(text string) string {
	return ErrorStyle.Render(IconCross + " " + text)
}

// RenderWarning renders a warning message with icon
func RenderWarning(text string) string {
	return WarningStyle.Render(IconWarning + " " + text)
}

// RenderSuccess renders a success message with icon
func RenderSuccess(text string) string {
	return SuccessStyle.Render(IconCheck + " " + text)
}

// RenderInfo renders an info message with icon
func RenderInfo(text string) string {
	return InfoStyle.Render(IconInfo + " " + text)
}

// RenderKeyBinding renders a keyboard shortcut
func RenderKeyBinding(key, desc string) string {
	return KeyStyle.Render(key) + " " + DescStyle.Render(desc)
}

// RenderProgressBar renders a simple progress bar
func RenderProgressBar(current, total int, width int) string {
	if total == 0 {
		return ProgressBarEmptyStyle.Render(lipgloss.PlaceHorizontal(width, lipgloss.Left, ""))
	}

	percentage := float64(current) / float64(total)
	filled := int(float64(width) * percentage)
	empty := width - filled

	bar := ProgressBarStyle.Render(lipgloss.PlaceHorizontal(filled, lipgloss.Left, "")) +
		ProgressBarEmptyStyle.Render(lipgloss.PlaceHorizontal(empty, lipgloss.Left, ""))

	return bar
}

// RenderTable renders a simple table with headers and rows
func RenderTable(headers []string, rows [][]string) string {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = lipgloss.Width(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				w := lipgloss.Width(cell)
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	// Render header
	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells,
			TableHeaderStyle.Width(widths[i]).Render(h))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)

	// Render rows
	var rowStrs []string
	for _, row := range rows {
		var cells []string
		for i, cell := range row {
			if i < len(widths) {
				cells = append(cells,
					TableCellStyle.Width(widths[i]).Render(cell))
			}
		}
		rowStrs = append(rowStrs, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	// Join all parts
	parts := append([]string{header}, rowStrs...)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// AdaptToTerminal adjusts styles based on terminal width and height
func AdaptToTerminal(width, height int) {
	// Adjust dialog box width if terminal is narrow
	if width < 70 {
		DialogBoxStyle = DialogBoxStyle.Width(width - 10)
	}

	// Adjust document padding for narrow terminals
	if width < 60 {
		DocStyle = DocStyle.Padding(0, 1)
	}
}
