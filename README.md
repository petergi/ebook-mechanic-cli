# ebook-mechanic-cli

A powerful terminal interface for validating and repairing EPUB and PDF ebooks, powered by `ebook-mechanic-lib`. This tool offers both an interactive Terminal User Interface (TUI) and a scriptable Command Line Interface (CLI).

## Features

- **Interactive TUI**:
  - Intuitive file browser with multi-selection support
  - Visual validation and repair reports
  - Real-time progress tracking
  - Clickable file paths and saved reports
  - Save reports directly from the interface

- **Powerful CLI**:
  - Comprehensive `validate` and `repair` commands
  - Advanced **Batch Processing** with worker pools
  - Multiple output formats: Text, JSON, Markdown
  - Filtering by severity and error limits
  - In-place repair with automatic backups
  - Robust file discovery (extensions, max depth, ignore patterns)

## Installation

### From Source

Requirements: Go 1.25+

```bash
go install github.com/petergi/ebook-mechanic-cli/cmd/ebm@latest
```

### Build Manually

```bash
git clone https://github.com/petergi/ebook-mechanic-cli.git
cd ebook-mechanic-cli
make build
```

## Quick Start

### Interactive Mode (TUI)

Simply run the command without arguments to launch the interactive interface:

```bash
ebm
```

**Key Bindings:**
- `↑`/`↓` or `j`/`k`: Navigate menus and lists
- `Space`: Toggle file selection (Multi-Select mode)
- `Enter`: Select item / Confirm action
- `a`: Select all files
- `A`: Deselect all
- `s`: Save report (in Report view) / Submit selection (in Multi-Select mode)
- `1-4`: Filter batch report tabs (Invalid, Errored, Valid, All)
- `q` / `Esc`: Go back or Quit

### Command Line Mode (CLI)

Use `ebm` with subcommands for scriptable operations, or simply provide a path for quick validation.

#### Quick Validation (Bypass Mode)

Validate a file or directory immediately without TUI:

```bash
ebm mybook.epub          # Validate single file
ebm ./library            # Batch validate directory
```

#### Detailed Reporting

Batch reports now categorize files into three groups:
- **Valid**: Passed validation successfully.
- **Invalid**: Failed validation (e.g., corrupt zip, missing manifest).
- **System Errors**: IO errors, permission denied, etc.

#### Command Examples

Validate with JSON output:
```bash
ebm validate book.epub --format json --output report.json
```

Filter issues by severity:
```bash
ebm validate book.epub --min-severity error
```

#### Repair

Repair a file to a new location:
```bash
ebm repair book.epub --output fixed_book.epub
```

Repair in-place with a backup:
```bash
ebm repair book.epub --in-place --backup
```

#### Batch Processing

Batch validate a directory with 8 workers:
```bash
ebm batch validate ./library --jobs 8
```

Batch repair recursively, ignoring "draft" folders:
```bash
ebm batch repair ./library --in-place --backup --ignore "*/draft/*"
```

Limit batch processing depth and extensions:
```bash
ebm batch validate ./library --max-depth 2 --ext .epub
```

## CLI Reference

### Global Flags

- `--format, -f`: Output format (`text`, `json`, `markdown`) [default: text]
- `--output, -o`: Write report to file instead of stdout
- `--verbose, -v`: Enable verbose output
- `--color`: Enable/disable colored output [default: true]
- `--min-severity`: Minimum severity to include (`info`, `warning`, `error`)
- `--max-errors`: Limit number of errors per report (0 = unlimited)

### Batch Flags

- `--jobs, -j`: Number of concurrent workers [default: num_cpu]
- `--recursive, -r`: Process subdirectories recursively [default: true]
- `--max-depth`: Maximum directory depth (-1 = unlimited)
- `--ext`: File extensions to include (default: .epub, .pdf)
- `--ignore`: Glob patterns to ignore
- `--progress`: Progress output mode (`auto`, `simple`, `none`)
- `--summary-only`: Only print summary output

## Development

```bash
make test    # Run tests
make docs    # Generate documentation
make lint    # Run linter
make fmt     # Format code
```

## License

TBD
