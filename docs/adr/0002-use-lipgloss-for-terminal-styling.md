# ADR 0002: Use Lipgloss for Terminal Styling

## Status

Accepted

## Context

The ebook-mechanic-cli TUI needs a robust styling solution for:
- Consistent visual design across all components
- Semantic coloring (errors=red, warnings=yellow, success=green)
- Responsive layouts that adapt to terminal size
- Accessibility (support for color-disabled terminals)
- Maintainable style definitions

### Considered Alternatives

1. **ANSI Escape Codes** (manual)
   - Pros: No dependencies, full control
   - Cons: Error-prone, hard to maintain, no layout assistance

2. **Lipgloss** (charmbracelet/lipgloss)
   - Pros: Declarative styling, layout utilities, responsive, integrates with Bubbletea
   - Cons: Additional dependency

3. **termenv** (muesli/termenv)
   - Pros: Cross-platform terminal detection, color adaptation
   - Cons: Lower-level, no layout features (note: Lipgloss uses termenv internally)

4. **Color** (fatih/color)
   - Pros: Simple color API
   - Cons: No layout capabilities, less declarative

## Decision

We will use **Lipgloss** for all terminal styling and layout in ebook-mechanic-cli.

## Rationale

### Declarative Styling

Lipgloss provides a CSS-like API for terminal styling:
```go
errorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("9")).
    Bold(true).
    Border(lipgloss.RoundedBorder()).
    Padding(1)
```

This is more maintainable and readable than manual ANSI codes.

### Layout Capabilities

Lipgloss includes powerful layout primitives:
- **JoinHorizontal/JoinVertical**: Combine elements
- **Place**: Position elements with alignment
- **Width/Height**: Responsive sizing
- **Borders and Padding**: Box model for spacing

These are essential for building complex TUI layouts.

### Semantic Design System

We can define a complete design system:
```go
var (
    ErrorStyle   = baseStyle.Foreground(lipgloss.Color("9"))
    WarningStyle = baseStyle.Foreground(lipgloss.Color("11"))
    SuccessStyle = baseStyle.Foreground(lipgloss.Color("10"))
    InfoStyle    = baseStyle.Foreground(lipgloss.Color("12"))
)
```

This ensures consistency across the application.

### Terminal Adaptation

- **Color Support Detection**: Automatically adapts to terminal capabilities
- **Color Profiles**: Supports TrueColor, ANSI256, ANSI16
- **Graceful Degradation**: Falls back to basic colors when needed

### Integration with Bubbletea

- Same maintainers and design philosophy
- Designed to work together seamlessly
- Shared patterns and best practices
- Many examples in the wild

### Performance

- **Efficient Rendering**: Minimizes ANSI code generation
- **Style Caching**: Styles are reusable and cached
- **Low Overhead**: Minimal performance impact

## Consequences

### Positive

1. **Maintainability**: Centralized style definitions in `theme.go`
2. **Consistency**: Semantic colors ensure uniform UX
3. **Responsive**: Layouts adapt to terminal size changes
4. **Accessibility**: Support for non-color terminals
5. **Productivity**: Fast UI development with layout utilities

### Negative

1. **Dependency**: Adds external dependency (mitigated by being well-maintained)
2. **Bundle Size**: Small increase in binary size (acceptable trade-off)

### Mitigation

- Keep Lipgloss updated to benefit from performance improvements
- Use style caching to avoid recreation
- Provide fallback for environments where Lipgloss has issues

## Implementation Notes

### Style Organization

Create centralized style definitions in `internal/tui/styles/theme.go`:

```go
package styles

import "github.com/charmbracelet/lipgloss"

// Semantic Colors
var (
    ColorError   = lipgloss.Color("9")   // Red
    ColorWarning = lipgloss.Color("11")  // Yellow
    ColorSuccess = lipgloss.Color("10")  // Green
    ColorInfo    = lipgloss.Color("12")  // Blue
    ColorPrimary = lipgloss.Color("14")  // Cyan
    ColorMuted   = lipgloss.Color("8")   // Gray
)

// Base Styles
var (
    BaseStyle = lipgloss.NewStyle().
        Padding(0, 1)

    ErrorStyle = BaseStyle.
        Foreground(ColorError).
        Bold(true)

    WarningStyle = BaseStyle.
        Foreground(ColorWarning)

    SuccessStyle = BaseStyle.
        Foreground(ColorSuccess)

    InfoStyle = BaseStyle.
        Foreground(ColorInfo)
)

// Component Styles
var (
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        MarginBottom(1)

    BorderStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorPrimary).
        Padding(1, 2)

    ListItemStyle = lipgloss.NewStyle().
        PaddingLeft(2)

    SelectedListItemStyle = ListItemStyle.
        Foreground(ColorPrimary).
        Bold(true)
)
```

### Responsive Layouts

Use width/height utilities for terminal size adaptation:

```go
func (m model) View() string {
    width := m.terminalWidth
    height := m.terminalHeight

    content := lipgloss.NewStyle().
        Width(width - 4).
        Height(height - 4).
        Render(m.content)

    return BorderStyle.Render(content)
}
```

### Accessibility

Support both colored and non-colored output:

```go
func NewTheme(colorEnabled bool) *Theme {
    if !colorEnabled {
        // Return monochrome theme
        return &Theme{
            Error: lipgloss.NewStyle().Bold(true),
            Warning: lipgloss.NewStyle(),
            // ...
        }
    }
    // Return colored theme
}
```

## Testing Strategy

### Style Tests

Verify styles are applied correctly:

```go
func TestErrorStyle(t *testing.T) {
    rendered := ErrorStyle.Render("Error message")
    assert.Contains(t, rendered, "Error message")
    // Verify ANSI codes for red color
}
```

### Snapshot Tests

Compare rendered output against expected:

```go
func TestReportView(t *testing.T) {
    model := NewReportModel(sampleReport)
    view := model.View()
    golden.Assert(t, view, "report_view.golden")
}
```

## References

- [Lipgloss GitHub](https://github.com/charmbracelet/lipgloss)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Style Guide Tutorial](https://github.com/charmbracelet/lipgloss#style-definitions)
