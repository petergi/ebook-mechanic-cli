# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ebook-mechanic-cli is a dual-mode (CLI/TUI) tool for ebook processing and management, written in Go.

## Development Commands

### Build

```bash
go build ./cmd/ebm
```

### Run

```bash
# Run TUI
go run ./cmd/ebm

# Run CLI command
go run ./cmd/ebm validate test.epub
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -run TestName ./path/to/package

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
golangci-lint run
```

### Build for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o build/ebm-linux ./cmd/ebm

# macOS
GOOS=darwin GOARCH=amd64 go build -o build/ebm-darwin ./cmd/ebm

# Windows
GOOS=windows GOARCH=amd64 go build -o build/ebm-windows.exe ./cmd/ebm
```

## Project Structure

- `cmd/ebm/`: Application entry point
- `internal/cli/`: CLI command implementations (cobra)
- `internal/tui/`: Interactive TUI implementation (bubbletea)
- `internal/operations/`: Core business logic and batch processing
- `internal/config/`: Configuration handling

## Code Organization

When implementing features:

1. **Dual-Mode**: Ensure features are accessible via both TUI and CLI where appropriate.
2. **CLI Framework**: Use `cobra` for CLI commands in `internal/cli`.
3. **TUI Framework**: Use `bubbletea` and `lipgloss` for TUI components in `internal/tui`.
4. **Shared Logic**: Place core business logic in `internal/operations` to be shared between modes.
5. **Batch Processing**: Use `internal/operations/batch.go` for concurrent file processing.

## Development Notes

- **Conventions**: Follow Go standard project layout.
- **TUI**: Test TUI changes manually as automated testing for TUI is limited.
- **Performance**: Be mindful of large batch operations; use the worker pool implementation.
