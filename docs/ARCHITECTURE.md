
# Architecture

## Overview

ebook-mechanic-cli is a dual-mode application providing both a Terminal User Interface (TUI) and a Command Line Interface (CLI) for validating and repairing EPUB and PDF ebook files. It is built on top of the `ebook-mechanic-lib` library, utilizing Bubbletea/Lipgloss for the TUI and Cobra for the CLI.

## Design Principles

1. **Dual-Mode Experience**: Offer a rich interactive TUI for exploration and a scriptable CLI for automation, selectable at runtime.
2. **Separation of Concerns**: Maintain clear boundaries between UI logic (TUI/CLI), business operations, and the underlying library.
3. **Testability**: Achieve high test coverage through comprehensive unit and integration tests for both modes.
4. **Extensibility**: Design components to be easily extended for new features and operations.
5. **Performance**: Ensure responsiveness and efficiency, especially during batch operations, using concurrent processing.

## High-Level Architecture

---

```text
┌─────────────────────────────────────────────────────────────┐
│                         User                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Entry Point (main.go)                    │
│           (Detects mode based on arguments)                 │
└────────────┬───────────────────────────────┬────────────────┘
             │                               │
             ▼                               ▼
┌───────────────────────────┐   ┌─────────────────────────────┐
│   TUI Layer (Bubbletea)   │   │      CLI Layer (Cobra)      │
│  ┌──────┐ ┌──────┐ ┌────┐ │   │  ┌──────┐ ┌──────┐ ┌─────┐  │
│  │ Menu │ │ File │ │Prog│ │   │  │ Root │ │ Val  │ │ Rep │  │
│  └──────┘ └──────┘ └────┘ │   │  └──────┘ └──────┘ └─────┐  │
│                           │   │  ┌──────┐ ┌──────┐ ┌─────┐  │
│  Styling: Lipgloss        │   │  │Batch │ │Output│ │Flags│  │
└────────────┬──────────────┘   │  └──────┘ └──────┘ └─────┘  │
             │                  └────────────┬────────────────┘
             │                               │
             ▼                               ▼
┌─────────────────────────────────────────────────────────────┐
│                  Operations Layer                           │
│  ┌────────────────────────────────────────────────┐         │
│  │  Validate  │  Repair  │  Batch  │  Discovery   │         │
│  └────────────────────────────────────────────────┘         │
│                                                             │
│  - Orchestrates library calls                               │
│  - Manages concurrent worker pools                          │
│  - Handles errors, context, and progress reporting          │
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

- **menu.go**: Main menu with multi-select capability options.
- **browser.go**: Interactive file browser supporting single and multi-file selection.
- **progress.go**: Progress indicator with real-time updates and spinner.
- **report.go**: Styled report viewer with tabs for filtering (Valid, Invalid, Errored).

#### Application (`internal/tui/app.go`)

- Coordinates state transitions and handles global messages (like batch progress streaming).

### 2. CLI Layer (`internal/cli/`)

The CLI layer uses Cobra for command processing:

- **root.go**: Root command and global flags. Handles "bypass mode" (running file/dir directly).
- **validate.go**: `validate` command implementation.
- **repair.go**: `repair` command implementation.
- **batch.go**: `batch` subcommands for bulk processing.
- **output.go**: Formatters for Text, JSON, and Markdown output.
- **reporter.go**: Logic for generating and writing reports.

### 3. Operations Layer (`internal/operations/`)

Shared business logic bridging TUI/CLI and the library:

- **batch.go**:
  - Worker pool implementation for concurrent processing.
  - Result aggregation and categorization (Valid, Invalid, Errored).
  - Context-aware cancellation and progress channel management.
- **validate.go** & **repair.go**: Single file operation wrappers.
- **File Discovery**: Robust logic for finding files (recursive, max depth, glob ignores).

### 4. Configuration (`internal/config/`)

- Application configuration (defaults, user preferences).

## Data Flow

### Batch Processing Flow (CLI/TUI)

---

```text
User initiates batch (CLI args or TUI selection)
       ↓
Find Files (operations.FindFiles)
       ↓
Initialize Worker Pool (BatchProcessor)
       ↓
Spawn Workers (goroutines) ← Tasks (channel)
       ↓
Stream Progress updates → CLI ProgressBar / TUI Progress Model
       ↓
Collect Results (Valid/Invalid/Errored)
       ↓
Generate Report (Formatters/ReportModel)
```

---

## Concurrency Model

- **Worker Pools**: Operations use a configurable worker pool to process files concurrently, maximizing throughput while respecting system limits.
- **Channels**: Communication between workers, the coordinator, and the UI (CLI/TUI) happens via buffered channels to prevent blocking.
- **Context Propagation**: `context.Context` is used throughout to handle cancellation (e.g., Ctrl+C) gracefully across all layers.

## Error Handling

- **Categorization**: Errors are categorized into Validation Failures (invalid content) and System Errors (IO, permissions).
- **Graceful Failure**: In batch mode, individual file errors do not stop the entire process (unless configured to).
- **Reporting**: Detailed error messages are provided in reports, distinguished by severity.

## Performance

- **Streaming**: Large file processing avoids loading entire contents into memory where possible.
- **Concurrency**: CPU-bound tasks (validation) are parallelized.
- **Rendering**: TUI rendering is optimized to only update changed parts; CLI output is buffered.
