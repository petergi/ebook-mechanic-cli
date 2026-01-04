# ADR 0001: Use Bubbletea for TUI Framework

## Status

Accepted

## Context

The ebook-mechanic-cli requires a Terminal User Interface (TUI) to provide an interactive, user-friendly experience for validating and repairing ebook files. We need a framework that:

- Provides robust terminal interaction handling
- Supports complex UI state management
- Offers good performance for responsive UIs
- Has active maintenance and community support
- Works cross-platform (Linux, macOS, Windows)
- Integrates well with modern Go practices

### Considered Alternatives

1. **tview** (rivo/tview)
   - Pros: Rich widget library, table/list views
   - Cons: More opinionated, heavier framework, harder to test

2. **termui** (gizak/termui)
   - Pros: Dashboard-style layouts, charts
   - Cons: Less active maintenance, focused on dashboards not interactive UIs

3. **Bubbletea** (charmbracelet/bubbletea)
   - Pros: Clean MVU architecture, highly testable, active ecosystem, composable
   - Cons: More code for complex widgets (but mitigated by bubbles library)

4. **Custom TUI** with termbox-go
   - Pros: Full control, lightweight
   - Cons: Significant development effort, need to handle all edge cases

## Decision

We will use **Bubbletea** as the TUI framework for ebook-mechanic-cli.

## Rationale

### Architecture Alignment

Bubbletea's Model-View-Update (MVU/Elm) architecture provides:
- **Clear separation of concerns**: State, update logic, and rendering are distinct
- **Testability**: Pure functions for updates and views make testing straightforward
- **Predictability**: Unidirectional data flow eliminates state synchronization issues

This aligns with our goal of 95%+ test coverage, as each component can be tested in isolation.

### Ecosystem Integration

- **Lipgloss**: Seamless integration for styling and layouts (same maintainers)
- **Bubbles**: Reusable components (progress bars, spinners, text inputs, lists)
- **Active Community**: Regular updates, extensive examples, responsive maintainers

### Developer Experience

- **Type Safety**: Strong typing throughout the framework
- **Composability**: Models can be composed and reused
- **Message Passing**: Clean concurrency model via messages and commands
- **Documentation**: Excellent tutorials and examples

### Performance

- **Efficient Rendering**: Only re-renders when state changes
- **Non-Blocking**: Command pattern for async operations
- **Low Overhead**: Minimal runtime overhead compared to alternatives

### Cross-Platform Support

- **Windows/macOS/Linux**: Works consistently across platforms
- **Terminal Compatibility**: Handles various terminal emulators gracefully

## Consequences

### Positive

1. **Clean Architecture**: MVU pattern enforces good practices
2. **Easy Testing**: Pure functions and message passing simplify test writing
3. **Extensibility**: Adding new views/models is straightforward
4. **Future-Proof**: Active development and strong community
5. **Integration**: Works seamlessly with Lipgloss for styling

### Negative

1. **Learning Curve**: MVU pattern may be unfamiliar to some developers
2. **Boilerplate**: More code required compared to widget-heavy frameworks
3. **Widget Library**: Need to build or use bubbles for common components

### Mitigation

- Provide clear examples and documentation of MVU patterns in codebase
- Use bubbles library for common UI components (progress, spinners, lists)
- Create reusable model templates for new features

## Implementation Notes

### Model Structure

Each screen/feature will have its own model:
- `menu.go`: Main menu
- `browser.go`: File browser
- `progress.go`: Progress indicator
- `report.go`: Report viewer
- `repair.go`: Repair preview

### Message Types

Define clear message types for:
- User input (key presses)
- Navigation events
- Operation results
- Error conditions

### Testing Strategy

- **Unit Tests**: Test update functions with various messages
- **View Tests**: Verify rendered output (snapshot testing)
- **Integration Tests**: Test message flow between models

## References

- [Bubbletea GitHub](https://github.com/charmbracelet/bubbletea)
- [Bubbletea Tutorials](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Elm Architecture](https://guide.elm-lang.org/architecture/)
