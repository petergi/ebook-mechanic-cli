
# Architecture

## Overview

ebook-mechanic-cli is a Terminal User Interface (TUI) application for validating and repairing EPUB and PDF ebook files. It provides an interactive, visually appealing interface built on top of the `ebook-mechanic-lib` library, using Bubbletea for TUI framework and Lipgloss for styling.

## Design Principles

1. **User-Centric Interface**: Provide an intuitive, visually appealing TUI that makes ebook validation and repair accessible to all users
2. **Separation of Concerns**: Maintain clear boundaries between UI logic, business operations, and the underlying library
3. **Testability**: Achieve 95%+ test coverage through comprehensive unit, integration, and TUI component tests
4. **Extensibility**: Design components to be easily extended for new features and operations
5. **Performance**: Ensure responsive UI even during long-running operations using concurrent processing

## High-Level Architecture

---

```text
┌─────────────────────────────────────────────────────────────┐
│                         User                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    TUI Layer (Bubbletea)                    │
│  ┌────────────┐  ┌────────────┐  ┌───────────┐              │
│  │    Menu    │  │   Browser  │  │  Progress │              │
│  │   Model    │  │    Model   │  │   Model   │              │
│  └────────────┘  └────────────┘  └───────────┘              │
│  ┌────────────┐  ┌────────────┐                             │
│  │   Report   │  │   Repair   │                             │
│  │   Model    │  │   Model    │                             │
│  └────────────┘  └────────────┘                             │
│                                                             │
│  Styling: Lipgloss (colors, borders, layout)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Operations Layer                           │
│  ┌────────────────────────────────────────────────┐         │
│  │  Validate  │  Repair  │  Batch  │  Report      │         │
│  └────────────────────────────────────────────────┘         │
│                                                             │
│  - Orchestrates library calls                               │
│  - Manages concurrent operations                            │
│  - Handles errors and state                                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              ebook-mechanic-lib (External)                  │
│  ┌────────────────────────────────────────────────┐         │
│  │  ValidateEPUB  │  ValidatePDF  │  Repair       │         │
│  └────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Architecture

### 1. TUI Layer (`internal/tui/`)

The TUI layer implements the Bubbletea Model-View-Update (MVU) pattern:

#### Models (`internal/tui/models/`)

Each model represents a distinct screen or interaction mode:

- **menu.go**: Main menu for selecting operations (Validate, Repair, Batch)
- **browser.go**: Interactive file/directory browser with filtering
- **progress.go**: Progress indicator for long-running operations
- **report.go**: Styled validation/repair report viewer
- **repair.go**: Interactive repair preview with confirmation

**Model Responsibilities**:

- Maintain component state
- Handle user input (key presses, events)
- Update state based on messages
- Render view using Lipgloss styles

#### Styles (`internal/tui/styles/`)

- **theme.go**: Centralized Lipgloss style definitions
  - Color scheme (semantic colors: error=red, warning=yellow, success=green, info=blue)
  - Layout styles (borders, padding, margins)
  - Component-specific styles (tables, lists, boxes)
  - Responsive sizing utilities

#### Application (`internal/tui/app.go`)

- Main Bubbletea program coordinator
- Manages state transitions between models
- Handles global messages and commands
- Orchestrates model lifecycle

### 2. Operations Layer (`internal/operations/`)

The operations layer bridges the TUI and the library, providing:

#### validate.go

- Single file validation
- Context-aware operation (cancellation support)
- Progress reporting via channels
- Error handling and reporting

#### repair.go

- Single file repair
- Preview generation
- Interactive confirmation handling
- Backup management

#### batch.go

- Multi-file processing with worker pools
- Concurrent validation/repair
- Progress aggregation
- Error collection and reporting

**Design Pattern**: Each operation returns a result channel and an error channel, allowing the TUI to reactively update progress.

### 3. Configuration (`internal/config/`)

- **config.go**: Application configuration management
  - Default output formats
  - Concurrent worker limits
  - UI preferences (colors, navigation style)
  - File filters and exclusions

### 4. Utilities (`pkg/utils/`)

- **fileutils.go**: File system operations (walk, filter, glob)
- **formatters.go**: Output formatting helpers

## Data Flow

### Validation Flow

---

```text
User selects file → Browser Model
       ↓
File path → Menu Model → Validate Operation
       ↓
ValidateEPUB/PDF (ebook-mechanic-lib)
       ↓
Progress updates → Progress Model
       ↓
ValidationReport → Report Model → Styled display
```

---

### Repair Flow

---

```text
User selects file → Browser Model
       ↓
File path → Menu Model → Repair Operation
       ↓
PreviewRepair (ebook-mechanic-lib)
       ↓
RepairPreview → Repair Model (user confirms)
       ↓
User confirms → Apply Repair
       ↓
RepairResult → Report Model → Success message
```

---

### Batch Processing Flow

---

```text
User selects directory → Browser Model
       ↓
Directory path → Menu Model → Batch Operation
       ↓
Discover files → Worker Pool (concurrent processing)
       ↓
For each file: Validate/Repair → Progress updates
       ↓
Aggregate results → Report Model → Summary display
```

---

## State Management

The application uses Bubbletea's message-passing architecture:

### Message Types

- **Navigation Messages**: Menu selection, screen transitions
- **File Messages**: File selected, directory selected
- **Operation Messages**: Start validation, start repair, batch process
- **Progress Messages**: Progress update, operation complete
- **Error Messages**: Operation error, user error

### State Transitions

---

```text
Menu → Browser → Menu → Operation → Progress → Report → Menu
  ↑                                                        ↓
  └────────────────────────────────────────────────────────┘
```

---

## Error Handling

### Layers

1. **Library Layer**: Errors from ebook-mechanic-lib
2. **Operations Layer**: Wraps library errors with context
3. **TUI Layer**: Displays user-friendly error messages

### Strategy

- **Graceful Degradation**: Continue batch processing on individual file errors
- **User Feedback**: Clear error messages with actionable guidance
- **Error Recovery**: Allow retry without restarting application
- **Logging**: Structured error logging for debugging

## Testing Strategy

### Unit Tests (Target: 95%+ coverage)

- **TUI Models**: Test state transitions, message handling, view rendering
- **Operations**: Test validation, repair, batch processing logic
- **Utilities**: Test file operations, formatters

### Integration Tests

- **TUI Integration**: Test complete user flows (file selection → operation → report)
- **Library Integration**: Test operations with ebook-mechanic-lib
- **Batch Processing**: Test concurrent processing with multiple files

### TUI Component Tests

- **Snapshot Tests**: Verify rendered output matches expected format
- **Interaction Tests**: Simulate key presses and verify state changes
- **Style Tests**: Verify Lipgloss styles are applied correctly

### Test Organization

---

```text
tests/
├── tui/
│   ├── models_test.go      # Model unit tests
│   ├── styles_test.go      # Style rendering tests
│   └── integration_test.go # TUI flow tests
├── operations/
│   ├── validate_test.go    # Validation tests
│   ├── repair_test.go      # Repair tests
│   └── batch_test.go       # Batch processing tests
└── integration/
    └── e2e_test.go         # End-to-end tests
```

---

## Concurrency Model

### Batch Processing

- **Worker Pool**: Configurable number of workers (default: CPU count)
- **Task Queue**: Buffered channel for file processing tasks
- **Result Aggregation**: Concurrent-safe result collection
- **Progress Tracking**: Atomic counters for progress updates

### UI Responsiveness

- **Non-Blocking Operations**: All file operations run in goroutines
- **Message Passing**: Operations communicate via Bubbletea messages
- **Cancellation**: Context-based cancellation for long operations

## Performance Considerations

### File Processing

- **Streaming**: Use io.Reader for large files
- **Memory Management**: Process files individually, avoid loading all into memory
- **Progress Feedback**: Regular progress updates (every 100ms minimum)

### UI Rendering

- **Efficient Rendering**: Only re-render changed components
- **Terminal Size Adaptation**: Responsive layouts using Lipgloss
- **Lazy Loading**: Load file lists on-demand for large directories

## Security Considerations

1. **Path Traversal**: Validate all file paths before operations
2. **Resource Limits**: Limit concurrent operations to prevent DoS
3. **Input Validation**: Validate user input before processing
4. **Sensitive Data**: Don't log file contents or paths with sensitive information

## Extensibility

### Adding New Operations

1. Create operation function in `internal/operations/`
2. Add model in `internal/tui/models/` if needed
3. Wire up in `internal/tui/app.go`
4. Add tests for new functionality

### Adding New File Types

1. Extend operations to call appropriate library functions
2. Update file browser filters
3. Add file type detection logic
4. Update report formatting

### Customizing UI

1. Modify styles in `internal/tui/styles/theme.go`
2. Override in configuration
3. Support custom themes via config files

## Dependencies

### Core Dependencies

- **charmbracelet/bubbletea**: TUI framework (MVU architecture)
- **charmbracelet/lipgloss**: Terminal styling and layout
- **ebook-mechanic-lib**: Core validation and repair functionality

### Supporting Libraries

- **charmbracelet/bubbles**: Reusable TUI components (progress bars, spinners, lists)
- Standard library: context, sync, io, os, path/filepath

## Build and Deployment

### Build Targets

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### Distribution

- Single binary with no external dependencies (except ebook-mechanic-lib)
- No configuration files required (sensible defaults)
- Optional config file support for customization

## Future Considerations

1. **Plugin System**: Support for custom validators/repairers
2. **Remote File Support**: HTTP/S3 file access
3. **Watch Mode**: Continuous directory monitoring
4. **Export Formats**: Additional report formats (HTML, XML)
5. **Theming**: User-configurable color schemes and styles
