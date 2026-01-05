package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderTitle(t *testing.T) {
	result := RenderTitle("Test Title")

	if result == "" {
		t.Error("Expected non-empty rendered title")
	}

	if !strings.Contains(result, "Test Title") {
		t.Error("Expected rendered title to contain original text")
	}
}

func TestRenderSubtitle(t *testing.T) {
	result := RenderSubtitle("Test Subtitle")

	if result == "" {
		t.Error("Expected non-empty rendered subtitle")
	}

	if !strings.Contains(result, "Test Subtitle") {
		t.Error("Expected rendered subtitle to contain original text")
	}
}

func TestRenderError(t *testing.T) {
	result := RenderError("Error message")

	if result == "" {
		t.Error("Expected non-empty rendered error")
	}

	if !strings.Contains(result, "Error message") {
		t.Error("Expected rendered error to contain message text")
	}

	if !strings.Contains(result, IconCross) {
		t.Error("Expected rendered error to contain error icon")
	}
}

func TestRenderWarning(t *testing.T) {
	result := RenderWarning("Warning message")

	if result == "" {
		t.Error("Expected non-empty rendered warning")
	}

	if !strings.Contains(result, "Warning message") {
		t.Error("Expected rendered warning to contain message text")
	}

	if !strings.Contains(result, IconWarning) {
		t.Error("Expected rendered warning to contain warning icon")
	}
}

func TestRenderSuccess(t *testing.T) {
	result := RenderSuccess("Success message")

	if result == "" {
		t.Error("Expected non-empty rendered success")
	}

	if !strings.Contains(result, "Success message") {
		t.Error("Expected rendered success to contain message text")
	}

	if !strings.Contains(result, IconCheck) {
		t.Error("Expected rendered success to contain check icon")
	}
}

func TestRenderInfo(t *testing.T) {
	result := RenderInfo("Info message")

	if result == "" {
		t.Error("Expected non-empty rendered info")
	}

	if !strings.Contains(result, "Info message") {
		t.Error("Expected rendered info to contain message text")
	}

	if !strings.Contains(result, IconInfo) {
		t.Error("Expected rendered info to contain info icon")
	}
}

func TestRenderKeyBinding(t *testing.T) {
	result := RenderKeyBinding("ctrl+c", "quit")

	if result == "" {
		t.Error("Expected non-empty rendered key binding")
	}

	if !strings.Contains(result, "ctrl+c") {
		t.Error("Expected rendered key binding to contain key")
	}

	if !strings.Contains(result, "quit") {
		t.Error("Expected rendered key binding to contain description")
	}
}

func TestRenderProgressBar_Empty(t *testing.T) {
	result := RenderProgressBar(0, 10, 50)

	if result == "" {
		t.Error("Expected non-empty progress bar for 0/10")
	}
}

func TestRenderProgressBar_Partial(t *testing.T) {
	result := RenderProgressBar(5, 10, 50)

	if result == "" {
		t.Error("Expected non-empty progress bar for 5/10")
	}
}

func TestRenderProgressBar_Full(t *testing.T) {
	result := RenderProgressBar(10, 10, 50)

	if result == "" {
		t.Error("Expected non-empty progress bar for 10/10")
	}
}

func TestRenderProgressBar_ZeroTotal(t *testing.T) {
	result := RenderProgressBar(0, 0, 50)

	if result == "" {
		t.Error("Expected non-empty progress bar for 0/0 (edge case)")
	}
}

func TestRenderProgressBar_OverflowCurrent(t *testing.T) {
	// Test when current > total (should handle gracefully)
	result := RenderProgressBar(15, 10, 50)

	if result == "" {
		t.Error("Expected non-empty progress bar even when current > total")
	}
}

func TestRenderProgressBar_VariousWidths(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 10},
		{"medium", 50},
		{"wide", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderProgressBar(5, 10, tt.width)

			if result == "" {
				t.Errorf("Expected non-empty progress bar for width %d", tt.width)
			}
		})
	}
}

func TestRenderTable_Simple(t *testing.T) {
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"Item 1", "10"},
		{"Item 2", "20"},
	}

	result := RenderTable(headers, rows)
	plain := ansi.Strip(result)

	if result == "" {
		t.Error("Expected non-empty table")
	}

	if !strings.Contains(plain, "Name") {
		t.Error("Expected table to contain header 'Name'")
	}

	if !strings.Contains(plain, "Value") {
		t.Error("Expected table to contain header 'Value'")
	}

	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) < len(rows)+1 {
		t.Errorf("Expected table to include %d rows plus header, got %d lines", len(rows), len(lines))
	}
}

func TestRenderTable_EmptyRows(t *testing.T) {
	headers := []string{"Column1", "Column2"}
	rows := [][]string{}

	result := RenderTable(headers, rows)
	plain := ansi.Strip(result)

	if result == "" {
		t.Error("Expected non-empty table even with no rows")
	}

	if !strings.Contains(plain, "Column1") {
		t.Error("Expected table to contain header even with no rows")
	}
}

func TestRenderTable_UnevenRows(t *testing.T) {
	headers := []string{"A", "B", "C"}
	rows := [][]string{
		{"1", "2"},           // Missing column
		{"3", "4", "5"},      // Complete
		{"6", "7", "8", "9"}, // Extra column (should be ignored)
	}

	result := RenderTable(headers, rows)

	if result == "" {
		t.Error("Expected non-empty table with uneven rows")
	}
}

func TestRenderTable_WideContent(t *testing.T) {
	headers := []string{"Short", "Long"}
	rows := [][]string{
		{"A", "This is a very long cell content that should adjust column width"},
	}

	result := RenderTable(headers, rows)

	if result == "" {
		t.Error("Expected non-empty table with wide content")
	}
}

func TestRenderTable_MultipleRows(t *testing.T) {
	headers := []string{"Type", "Count", "Status"}
	rows := [][]string{
		{"Errors", "5", "Failed"},
		{"Warnings", "3", "Warning"},
		{"Info", "10", "Info"},
	}

	result := RenderTable(headers, rows)
	plain := ansi.Strip(result)

	if result == "" {
		t.Error("Expected non-empty table with multiple rows")
	}

	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) < len(rows)+1 {
		t.Errorf("Expected table to include %d rows plus header, got %d lines", len(rows), len(lines))
	}
}

func TestAdaptToTerminal_Narrow(t *testing.T) {
	// Store original values
	originalDialogWidth := DialogBoxStyle.GetWidth()
	originalDocPadding := DocStyle.GetPaddingLeft()

	// Adapt to narrow terminal
	AdaptToTerminal(50, 24)

	// Dialog box should be adjusted
	newDialogWidth := DialogBoxStyle.GetWidth()
	if newDialogWidth >= 50 {
		t.Error("Expected dialog box width to be reduced for narrow terminal")
	}

	// Doc style should be adjusted
	newDocPadding := DocStyle.GetPaddingLeft()
	if newDocPadding >= originalDocPadding {
		t.Error("Expected doc padding to be reduced for narrow terminal")
	}

	// Reset styles
	DialogBoxStyle = DialogBoxStyle.Width(originalDialogWidth)
	DocStyle = DocStyle.Padding(1, 2)
}

func TestAdaptToTerminal_MidWidth(t *testing.T) {
	originalDialogWidth := DialogBoxStyle.GetWidth()
	originalDocPadding := DocStyle.GetPaddingLeft()

	// Width triggers dialog adjustment but keeps doc padding
	AdaptToTerminal(65, 24)

	dialogWidth := DialogBoxStyle.GetWidth()
	if dialogWidth >= 65 {
		t.Error("Expected dialog box width to be reduced for mid-width terminal")
	}

	docPadding := DocStyle.GetPaddingLeft()
	if docPadding != originalDocPadding {
		t.Error("Expected doc padding to remain unchanged for mid-width terminal")
	}

	DialogBoxStyle = DialogBoxStyle.Width(originalDialogWidth)
	DocStyle = DocStyle.Padding(1, 2)
}

func TestAdaptToTerminal_VeryNarrow(t *testing.T) {
	// Adapt to very narrow terminal
	AdaptToTerminal(40, 24)

	// Both adjustments should apply
	dialogWidth := DialogBoxStyle.GetWidth()
	if dialogWidth >= 40 {
		t.Error("Expected dialog box width to be adjusted for very narrow terminal")
	}

	docPadding := DocStyle.GetPaddingLeft()
	if docPadding > 1 {
		t.Error("Expected doc padding to be minimal for very narrow terminal")
	}

	// Reset styles
	DialogBoxStyle = DialogBoxStyle.Width(60)
	DocStyle = DocStyle.Padding(1, 2)
}

func TestAdaptToTerminal_Wide(t *testing.T) {
	// Store original values
	originalDialogWidth := 60
	DialogBoxStyle = DialogBoxStyle.Width(originalDialogWidth)

	// Adapt to wide terminal
	AdaptToTerminal(120, 40)

	// Dialog box width should not change for wide terminals
	// (function only makes things smaller, not larger)
	dialogWidth := DialogBoxStyle.GetWidth()
	if dialogWidth > originalDialogWidth {
		t.Error("Expected dialog box width to not increase for wide terminal")
	}
}

func TestIcons(t *testing.T) {
	// Test that icon constants are defined and non-empty
	icons := map[string]string{
		"IconCheck":   IconCheck,
		"IconCross":   IconCross,
		"IconWarning": IconWarning,
		"IconInfo":    IconInfo,
		"IconArrow":   IconArrow,
		"IconBullet":  IconBullet,
		"IconSpinner": IconSpinner,
	}

	for name, icon := range icons {
		if icon == "" {
			t.Errorf("Expected %s to be non-empty", name)
		}
	}
}

func TestIconSpinner_AnimationFrames(t *testing.T) {
	if len(IconSpinner) < 4 {
		t.Error("Expected IconSpinner to have multiple animation frames")
	}
}

func TestColorDefinitions(t *testing.T) {
	// Test that color variables are initialized
	// We can't easily test the actual colors, but we can verify they exist
	colors := []lipgloss.AdaptiveColor{
		ColorError,
		ColorWarning,
		ColorSuccess,
		ColorInfo,
		ColorPrimary,
		ColorMuted,
		ColorBg,
		ColorFg,
	}

	for i, color := range colors {
		// Verify both light and dark values are set (non-empty strings)
		if color.Light == "" || color.Dark == "" {
			t.Errorf("Expected color %d to have both light and dark values", i)
		}
	}
}

func TestStyleDefinitions(t *testing.T) {
	// Test that style variables are initialized
	// We can't easily test the exact styling, but we can verify rendering works
	styles := map[string]func(string) string{
		"ErrorStyle":    func(s string) string { return ErrorStyle.Render(s) },
		"WarningStyle":  func(s string) string { return WarningStyle.Render(s) },
		"SuccessStyle":  func(s string) string { return SuccessStyle.Render(s) },
		"InfoStyle":     func(s string) string { return InfoStyle.Render(s) },
		"MutedStyle":    func(s string) string { return MutedStyle.Render(s) },
		"TitleStyle":    func(s string) string { return TitleStyle.Render(s) },
		"SubtitleStyle": func(s string) string { return SubtitleStyle.Render(s) },
		"ListItemStyle": func(s string) string { return ListItemStyle.Render(s) },
		"KeyStyle":      func(s string) string { return KeyStyle.Render(s) },
		"DescStyle":     func(s string) string { return DescStyle.Render(s) },
		"HelpStyle":     func(s string) string { return HelpStyle.Render(s) },
	}

	for name, renderFunc := range styles {
		result := renderFunc("test")
		if result == "" {
			t.Errorf("Expected %s to render non-empty string", name)
		}
	}
}

func TestBorderStyles(t *testing.T) {
	// Test border styles render without panicking
	text := "test content"

	borderResult := BorderStyle.Render(text)
	if borderResult == "" {
		t.Error("Expected BorderStyle to render non-empty string")
	}

	focusedResult := FocusedBorderStyle.Render(text)
	if focusedResult == "" {
		t.Error("Expected FocusedBorderStyle to render non-empty string")
	}

	dialogResult := DialogBoxStyle.Render(text)
	if dialogResult == "" {
		t.Error("Expected DialogBoxStyle to render non-empty string")
	}
}

func TestTableStyles(t *testing.T) {
	// Test table styles render without panicking
	text := "cell"

	headerResult := TableHeaderStyle.Render(text)
	if headerResult == "" {
		t.Error("Expected TableHeaderStyle to render non-empty string")
	}

	rowResult := TableRowStyle.Render(text)
	if rowResult == "" {
		t.Error("Expected TableRowStyle to render non-empty string")
	}

	cellResult := TableCellStyle.Render(text)
	if cellResult == "" {
		t.Error("Expected TableCellStyle to render non-empty string")
	}
}

func TestStatusStyles(t *testing.T) {
	// Test status styles render without panicking
	text := "status"

	validResult := ValidStyle.Render(text)
	if validResult == "" {
		t.Error("Expected ValidStyle to render non-empty string")
	}

	invalidResult := InvalidStyle.Render(text)
	if invalidResult == "" {
		t.Error("Expected InvalidStyle to render non-empty string")
	}

	processingResult := ProcessingStyle.Render(text)
	if processingResult == "" {
		t.Error("Expected ProcessingStyle to render non-empty string")
	}

	pendingResult := PendingStyle.Render(text)
	if pendingResult == "" {
		t.Error("Expected PendingStyle to render non-empty string")
	}
}

func TestProgressBarStyles(t *testing.T) {
	// Test progress bar styles render without panicking
	text := "█"

	filledResult := ProgressBarStyle.Render(text)
	if filledResult == "" {
		t.Error("Expected ProgressBarStyle to render non-empty string")
	}

	emptyResult := ProgressBarEmptyStyle.Render(text)
	if emptyResult == "" {
		t.Error("Expected ProgressBarEmptyStyle to render non-empty string")
	}
}

func TestRenderHyperlink(t *testing.T) {
	url := "https://example.com"
	text := "Example"
	got := RenderHyperlink(url, text)

	if !strings.Contains(got, url) {
		t.Errorf("Hyperlink does not contain url: %s", got)
	}
	if !strings.Contains(got, text) {
		t.Errorf("Hyperlink does not contain text: %s", got)
	}
}

func TestRenderStyledHyperlink(t *testing.T) {
	url := "https://example.com"
	text := "Example"
	got := RenderStyledHyperlink(url, text)

	if got == "" {
		t.Error("Styled hyperlink is empty")
	}
}
