# ebook-mechanic-cli Feature Parity & Enhancement Plan

**Branch**: `feature/cli-parity-and-enhancements`
**Created**: 2026-01-03
**Status**: Completed

## Executive Summary

This plan outlines the implementation of missing features to achieve feature parity with ebook-mechanic-lib CLI while adding requested TUI enhancements. The work is divided into 5 phases, each building on the previous one. This plan is now complete.

## Technology Stack

**TUI Framework** (Required):
- `github.com/charmbracelet/bubbletea` - Terminal UI framework (Elm Architecture)
- `github.com/charmbracelet/lipgloss` - Terminal styling and layout

**CLI Framework**:
- `github.com/spf13/cobra` - Command-line interface framework
- `github.com/spf13/pflag` - POSIX-compliant flag parsing
- `github.com/fatih/color` - Colored terminal output
- `github.com/schollz/progressbar/v3` - Progress bars for batch operations
- `golang.org/x/term` - Terminal detection and control

**Core Library**:
- `github.com/petergi/ebook-mechanic-lib` - EPUB/PDF validation and repair

## Gap Analysis

### Final Implemented Features

✅ **Dual-Mode Operation**:
- Interactive TUI (default) and scriptable CLI modes.
- Automatic mode detection based on arguments.
- CLI bypass mode: `ebm <path>` defaults to validation.

✅ **CLI Features**:
- Subcommands: `validate`, `repair`, `batch`, `completion`.
- Multiple output formats: Text, JSON, and Markdown (`--format`).
- File output (`--output`) and stdin support (`-`).
- Severity filtering (`--min-severity`) and error limits (`--max-errors`).
- In-place repair with atomic replace (`--in-place`) and backups (`--backup`).
- Colored/non-colored output control (`--color`).
- Verbose mode (`--verbose`).

✅ **Batch Processing Features**:
- Concurrent worker pools (`--jobs`).
- Advanced file discovery: recursive, max depth (`--max-depth`), extension filtering (`--ext`), and glob ignore patterns (`--ignore`).
- Multiple progress reporting modes (`auto`, `simple`, `none`).
- Summary-only mode (`--summary-only`).
- Detailed batch reports with `Valid`, `Invalid`, and `Errored` categories.

✅ **TUI Enhancements**:
- Multi-file selection with checkboxes.
- Clickable file paths and report links (OSC 8).
- In-TUI report saving.
- Filterable batch reports (`Invalid`, `Errored`, `Valid`, `All`).
- Enhanced progress bar using `bubbles/progress`.

## Implementation Plan

---

## Phase 1: Dual-Mode Architecture (CLI + TUI)

**Goal**: Restructure to support both CLI and TUI modes
**Duration**: 2-3 hours
**Dependencies**: None

### 1.1 Restructure Entry Point
- **File**: `cmd/ebm/main.go`
- **Changes**:
  - Detect if running in TUI mode (no args) or CLI mode (with args)
  - Add cobra command structure for CLI mode
  - Keep TUI as default when no arguments provided
- **Implementation**:
  ```go
  func main() {
      // If no args, run TUI
      if len(os.Args) == 1 {
          if err := tui.Run(); err != nil {
              fmt.Fprintf(os.Stderr, "Error: %v\n", err)
              os.Exit(1)
          }
          return
      }

      // Otherwise, run CLI
      cmd := newRootCmd()
      if err := cmd.Execute(); err != nil {
          fmt.Fprintln(os.Stderr, err)
          os.Exit(1)
      }
  }
  ```

### 1.2 Create CLI Package Structure
- **New Package**: `internal/cli/`
- **Files to Create**:
  - `internal/cli/root.go` - Root command with global flags
  - `internal/cli/validate.go` - Validate command
  - `internal/cli/repair.go` - Repair command
  - `internal/cli/batch.go` - Batch commands (validate/repair)
  - `internal/cli/flags.go` - Shared flag definitions
  - `internal/cli/output.go` - Output formatters (text, json, markdown)
  - `internal/cli/reporter.go` - Report generation and writing

### 1.3 Add Dependencies
- **File**: `go.mod`
- **Add**:
  - `github.com/spf13/cobra` for CLI framework
  - `github.com/spf13/pflag` for flag parsing
  - `github.com/fatih/color` for colored CLI output
  - `encoding/json` (standard library)
  - Consider: `github.com/jedib0t/go-pretty/v6` for table formatting

**Testing**:
- Verify TUI still works with no arguments
- Verify CLI help displays with `--help`
- Ensure smooth mode detection

---

## Phase 2: Core CLI Commands (Validate & Repair)

**Goal**: Implement validate and repair commands with output formats
**Duration**: 3-4 hours
**Dependencies**: Phase 1

### 2.1 Implement Validate Command
- **File**: `internal/cli/validate.go`
- **Features**:
  - Accept file path argument
  - Support stdin with `-` and `--type` flag
  - Output formats: text, json, markdown
  - Write to file with `--output`
  - Color control with `--color`
  - Verbose mode with `--verbose`
- **Flags**:
  ```
  --format, -f     Output format (text|json|markdown) [default: text]
  --output, -o     Write to file instead of stdout
  --verbose, -v    Enable verbose output
  --color          Enable colored output [default: true]
  --min-severity   Minimum severity to include (info|warning|error)
  --severity       Include only specific severities (repeatable)
  --max-errors     Limit number of errors per report (0 = unlimited)
  --type           File type for stdin input (epub|pdf)
  ```
- **Examples**:
  ```bash
  ebm validate book.epub
  ebm validate doc.pdf --format json
  cat book.epub | ebm validate - --type epub
  ebm validate book.epub --output report.md --format markdown
  ebm validate book.epub --min-severity error
  ```

### 2.2 Implement Repair Command
- **File**: `internal/cli/repair.go`
- **Features**:
  - Accept file path argument
  - Support output path with `--output`
  - In-place repair with `--in-place`
  - Backup creation with `--backup`
  - Custom backup directory with `--backup-dir`
  - Inherit all validation report flags
- **Flags**:
  ```
  --output, -o     Output path for repaired file
  --in-place       Repair file in place (atomic replace)
  --backup         Create backup before in-place repair
  --backup-dir     Directory for backups
  (plus all validation flags)
  ```
- **Examples**:
  ```bash
  ebm repair book.epub
  ebm repair book.epub --output fixed.epub
  ebm repair book.epub --in-place --backup
  ebm repair book.epub --in-place --backup --backup-dir ./backups
  ```

### 2.3 Implement Output Formatters
- **File**: `internal/cli/output.go`
- **Formatters**:
  - **Text**: Human-readable with colors, similar to current TUI report
  - **JSON**: Machine-readable structured output
  - **Markdown**: GitHub-flavored markdown with tables
- **Implementation**:
  ```go
  type OutputFormatter interface {
      FormatValidation(*ebmlib.ValidationReport) (string, error)
      FormatRepair(*ebmlib.RepairResult, *ebmlib.ValidationReport) (string, error)
  }

  type TextFormatter struct{ ColorEnabled bool }
  type JSONFormatter struct{}
  type MarkdownFormatter struct{}
  ```

### 2.4 Implement Severity Filtering
- **File**: `internal/cli/reporter.go`
- **Features**:
  - Filter by minimum severity level
  - Filter by specific severity types
  - Apply max errors limit
  - Respect severity hierarchy (error > warning > info)

**Testing**:
- Test each output format
- Test stdin input
- Test file output
- Test severity filtering
- Test backup creation
- Test in-place repair

---

## Phase 3: Batch Processing Enhancement

**Goal**: Implement advanced batch processing with worker pools
**Duration**: 4-5 hours
**Dependencies**: Phase 2

### 3.1 Implement Batch Validate Command
- **File**: `internal/cli/batch.go`
- **Features**:
  - Accept multiple paths (files or directories)
  - Worker pool with configurable size
  - Queue buffer size control
  - Directory traversal with depth limit
  - Extension filtering
  - Glob pattern exclusions
  - Progress reporting modes
  - Summary-only output
- **Flags**:
  ```
  --jobs, -j       Number of parallel workers [default: 4]
  --queue          Job queue buffer size [default: 64]
  --max-depth      Maximum directory depth (-1 = unlimited) [default: -1]
  --ext            File extensions to include [default: .epub,.pdf]
  --ignore         Glob patterns to ignore
  --progress       Progress output mode (auto|simple|none) [default: auto]
  --summary-only   Only print summary output
  (plus all validation flags)
  ```
- **Examples**:
  ```bash
  ebm batch validate ./books
  ebm batch validate ./library --ext .epub --jobs 8
  ebm batch validate ./books ./more-books --ignore "*/draft/*"
  ebm batch validate ./library --max-depth 2 --summary-only
  ```

### 3.2 Implement Batch Repair Command
- **File**: `internal/cli/batch.go`
- **Features**:
  - Same as batch validate
  - Add repair-specific flags (in-place, backup, backup-dir)
  - Atomic operations per file
  - Rollback on failure
- **Flags**:
  ```
  (all batch validate flags)
  --in-place       Repair files in place
  --backup         Create backups before repair
  --backup-dir     Directory for backups
  ```
- **Examples**:
  ```bash
  ebm batch repair ./books --in-place --backup
  ebm batch repair ./library --jobs 4 --backup-dir ./backups
  ```

### 3.3 Enhance Batch Operations Package
- **File**: `internal/operations/batch.go`
- **Enhancements**:
  - Add worker pool configuration
  - Add queue management
  - Add depth-limited directory walking
  - Add extension filtering
  - Add glob pattern matching for ignores
  - Add progress reporting callbacks
  - Thread-safe result aggregation
- **New Structures**:
  ```go
  type BatchConfig struct {
      Workers     int
      QueueSize   int
      MaxDepth    int
      Extensions  []string
      IgnoreGlobs []string
      Progress    ProgressMode
  }

  type ProgressMode int
  const (
      ProgressAuto ProgressMode = iota
      ProgressSimple
      ProgressNone
  )
  ```

### 3.4 Implement Progress Reporter
- **File**: `internal/cli/progress.go`
- **Features**:
  - Auto-detect TTY vs non-TTY
  - Simple mode: periodic text updates
  - Auto mode: full progress bar if TTY, simple if not
  - None mode: silent until completion
- **Implementation**:
  - Use `golang.org/x/term` to detect TTY
  - Consider `github.com/schollz/progressbar/v3` for progress bars

**Testing**:
- Test worker pool scaling (1, 4, 8, 16 workers)
- Test queue buffering
- Test directory depth limits
- Test extension filtering
- Test ignore patterns
- Test progress modes in TTY and non-TTY environments
- Test summary output

---

## Phase 4: TUI Enhancements

**Goal**: Add multi-file selection and clickable links
**Duration**: 3-4 hours
**Dependencies**: None (can be parallel with Phase 3)

### 4.1 Multi-File Selection in Browser
- **File**: `internal/tui/models/browser.go`
- **Features**:
  - Add checkbox-style multi-select mode
  - Toggle selection with spacebar
  - Select/deselect all with ctrl+a
  - Show selection count
  - Visual indicator for selected files
  - New mode: `BrowserModeMultiSelect`
- **Implementation**:
  ```go
  type BrowserModel struct {
      // ... existing fields ...
      mode         BrowserMode      // File, Batch, or MultiSelect
      selectedFiles map[string]bool  // Track selected files
      selectAll    bool             // Whether "select all" is active
  }

  const (
      BrowserModeFile BrowserMode = iota
      BrowserModeBatch
      BrowserModeMultiSelect
  )
  ```
- **Key Bindings**:
  - `space`: Toggle selection on current item
  - `ctrl+a`: Toggle select all
  - `enter`: Confirm selection (with selected files)
  - Display: `[x]` for selected, `[ ]` for unselected

### 4.2 Add Multi-Select Menu Option
- **File**: `internal/tui/models/menu.go`
- **Changes**:
  - Add "Validate Multiple Files" option
  - Add "Repair Multiple Files" option
  - These launch browser in multi-select mode
- **Menu Structure**:
  ```
  1. Validate Single EPUB/PDF
  2. Repair Single EPUB/PDF
  3. Validate Multiple Files (new)
  4. Repair Multiple Files (new)
  5. Batch Validate Directory
  6. Batch Repair Directory
  7. Quit
  ```

### 4.3 Handle Multi-Select Results
- **File**: `internal/tui/app.go`
- **Changes**:
  - Add `MultiFileSelectMsg` message type
  - Handle multi-file selection
  - Run operations on selected files with progress
  - Show aggregated results
- **Implementation**:
  ```go
  type MultiFileSelectMsg struct {
      Paths []string
  }
  ```

### 4.4 Clickable Links in TUI
- **File**: `internal/tui/models/report.go`
- **Features**:
  - Add OSC 8 hyperlinks for file paths
  - Make report file paths clickable
  - Add copy-to-clipboard hint
- **Implementation**:
  ```go
  // OSC 8 hyperlink format: \x1b]8;;<URL>\x1b\\<TEXT>\x1b]8;;\x1b\\
  func makeClickable(path string) string {
      absPath, _ := filepath.Abs(path)
      fileURL := "file://" + absPath
      return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", fileURL, path)
  }
  ```
- **Display**:
  - Show "Click to open" hint for supported terminals
  - Display full path with file:// protocol support
  - Add "Press 'c' to copy path" keybinding

### 4.5 Report File Generation from TUI
- **File**: `internal/tui/models/report.go`
- **Features**:
  - Add option to save report to file
  - Support multiple formats (text, json, markdown)
  - Show success message with clickable link
- **Key Bindings**:
  - `s`: Save report to file
  - `f`: Choose format (cycle through text/json/markdown)
  - Show saved path as clickable link

**Testing**:
- Test multi-select with 0, 1, and multiple files
- Test select all / deselect all
- Test clickable links in various terminals (iTerm2, Terminal.app, Windows Terminal)
- Test report file generation
- Verify hyperlinks work in supported terminals

---

## Phase 5: Integration & Polish

**Goal**: Integrate all features, comprehensive testing, documentation
**Duration**: 2-3 hours
**Dependencies**: Phases 1-4

### 5.1 Update Help & Documentation
- **Files**:
  - `CLAUDE.md` - Update development commands
  - `README.md` - Create comprehensive user guide
  - `cmd/ebm/main.go` - Update help text
  - `internal/cli/root.go` - Comprehensive command help
- **Documentation Sections**:
  - Installation
  - Quick Start (TUI and CLI)
  - Command Reference
  - Examples for each command
  - Configuration options
  - Output formats
  - Batch processing guide
  - TUI keyboard shortcuts
  - Hyperlink support by terminal

### 5.2 Add Shell Completions
- **Files**: `cmd/ebm/completion.go`
- **Features**:
  - Bash completion
  - Zsh completion
  - Fish completion
  - PowerShell completion
- **Command**: `ebm completion <shell>`

### 5.3 Comprehensive Testing
- **Unit Tests**:
  - All CLI commands
  - Output formatters
  - Severity filtering
  - Batch processing
  - Multi-select browser
- **Integration Tests**:
  - End-to-end CLI workflows
  - TUI mode switching
  - File generation and cleanup
  - Backup operations
- **Manual Testing**:
  - TUI visual appearance
  - Hyperlinks in different terminals
  - Progress reporting
  - Error handling
- **Target**: Maintain 95%+ code coverage

### 5.4 Performance Optimization
- **Areas**:
  - Worker pool efficiency
  - Memory usage in batch processing
  - Concurrent file operations
  - Progress update frequency
- **Benchmarks**:
  - Create benchmarks for batch operations
  - Test with 100, 1000, 10000 files
  - Measure memory footprint

### 5.5 Error Handling & Edge Cases
- **Scenarios**:
  - Invalid file paths
  - Permission errors
  - Corrupted files
  - Network filesystem issues
  - Ctrl+C handling (graceful shutdown)
  - Disk full errors
  - Concurrent file modifications
- **Recovery**:
  - Rollback on failure
  - Clear error messages
  - Helpful suggestions

### 5.6 CI/CD Integration
- **GitHub Actions**:
  - Run tests on push
  - Build binaries for multiple platforms
  - Generate coverage reports
  - Lint code
- **Release Automation**:
  - Semantic versioning
  - Changelog generation
  - Binary releases for Linux, macOS, Windows

**Testing**:
- Run full test suite
- Test on Linux, macOS, Windows
- Test in various terminal emulators
- Benchmark batch processing
- Verify all documentation examples work

---

## Implementation Order

### Week 1: Core CLI (Phases 1-2)
1. Day 1: Dual-mode architecture setup
2. Day 2: Validate command + output formatters
3. Day 3: Repair command + severity filtering

### Week 2: Advanced Features (Phases 3-4)
4. Day 4: Batch validate with worker pools
5. Day 5: Batch repair + progress reporting
6. Day 6: Multi-select browser
7. Day 7: Clickable links + report generation

### Week 3: Polish (Phase 5)
8. Day 8: Documentation + shell completions
9. Day 9: Comprehensive testing + bug fixes
10. Day 10: Performance optimization + CI/CD

## File Structure After Implementation

```
ebook-mechanic-cli/
├── cmd/
│   └── ebm/
│       ├── main.go           # Entry point with mode detection
│       └── completion.go     # Shell completion generation
├── internal/
│   ├── cli/                  # CLI mode implementation
│   │   ├── root.go          # Root command
│   │   ├── validate.go      # Validate command
│   │   ├── repair.go        # Repair command
│   │   ├── batch.go         # Batch commands
│   │   ├── flags.go         # Shared flags
│   │   ├── output.go        # Output formatters
│   │   ├── reporter.go      # Report generation
│   │   ├── progress.go      # Progress reporting
│   │   └── *_test.go        # Tests
│   ├── operations/           # Business logic (existing)
│   │   ├── validate.go
│   │   ├── repair.go
│   │   └── batch.go         # Enhanced with worker pools
│   └── tui/                  # TUI mode (existing)
│       ├── app.go           # Enhanced with multi-select
│       ├── models/
│       │   ├── menu.go      # Enhanced menu options
│       │   ├── browser.go   # Multi-select support
│       │   ├── progress.go
│       │   └── report.go    # Clickable links + save
│       └── styles/
├── CLAUDE.md                 # Updated dev guide
├── README.md                 # Comprehensive user guide
├── PLAN.md                   # This file
└── go.mod                    # Additional dependencies

```

## Dependencies to Add

```go
// Add to go.mod
require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/pflag v1.0.5
    github.com/fatih/color v1.16.0
    github.com/schollz/progressbar/v3 v3.14.1
    golang.org/x/term v0.15.0
)
```

## Success Criteria

- [x] TUI mode works exactly as before (no regressions)
- [x] CLI mode supports all ebook-mechanic-lib CLI features
- [x] Multi-file selection works in TUI
- [x] Clickable links work in supported terminals
- [x] All output formats (text, json, markdown) work correctly
- [x] Batch processing with worker pools performs well (test with 1000+ files)
- [x] 95%+ test coverage maintained
- [x] Comprehensive documentation
- [x] CI/CD pipeline functioning
- [x] Binaries built for Linux, macOS, Windows

## Risk Mitigation

1. **Breaking TUI**: Keep TUI code isolated, test after each phase
2. **Performance**: Profile early, optimize worker pool, test with large datasets
3. **Terminal Compatibility**: Test hyperlinks in multiple terminals, graceful degradation
4. **Complexity**: Keep CLI and TUI modes separate, clear boundaries
5. **Testing Burden**: Write tests alongside implementation, not after

## Next Steps

1. ✅ Create feature branch: `feature/cli-parity-and-enhancements`
2. ✅ Save plan to PLAN.md
3. ✅ Save plan to Pieces LTM
4. ✅ Save plan to Claude Code Memory
5. ✅ Begin Phase 1: Dual-Mode Architecture
6. ✅ Daily progress updates to PLAN.md

## Progress Tracking

Update this section as features are completed:

### Phase 1: Dual-Mode Architecture
- [x] 1.1 Restructure Entry Point
- [x] 1.2 Create CLI Package Structure
- [x] 1.3 Add Dependencies

### Phase 2: Core CLI Commands
- [x] 2.1 Implement Validate Command
- [x] 2.2 Implement Repair Command
- [x] 2.3 Implement Output Formatters
- [x] 2.4 Implement Severity Filtering

### Phase 3: Batch Processing
- [x] 3.1 Batch Validate Command
- [x] 3.2 Batch Repair Command
- [x] 3.3 Enhance Batch Operations
- [x] 3.4 Progress Reporter

### Phase 4: TUI Enhancements
- [x] 4.1 Multi-File Selection
- [x] 4.2 Multi-Select Menu
- [x] 4.3 Handle Multi-Select Results
- [x] 4.4 Clickable Links
- [x] 4.5 Report File Generation

### Phase 5: Integration & Polish
- [x] 5.1 Documentation
- [x] 5.2 Shell Completions
- [x] 5.3 Comprehensive Testing
- [x] 5.4 Performance Optimization
- [x] 5.5 Error Handling
- [x] 5.6 CI/CD Integration

---

**Last Updated**: 2026-01-04 13:00 EST
**Current Phase**: Complete (Documentation Updated)
