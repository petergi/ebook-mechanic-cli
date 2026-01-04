# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ebook-mechanic-cli is a CLI tool for ebook processing and management, written in Go.

## Development Commands

### Build
```bash
go build -o ebook-mechanic
```

### Run
```bash
go run main.go [command] [flags]
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -run TestName ./path/to/package

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting
```bash
# Install golangci-lint if not already installed
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

golangci-lint run
```

### Build for Multiple Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ebook-mechanic-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o ebook-mechanic-darwin

# Windows
GOOS=windows GOARCH=amd64 go build -o ebook-mechanic-windows.exe
```

## Project Structure

The project follows standard Go CLI application conventions:

- Build artifacts go to `build/` and `dist/` directories
- Completion scripts are generated to `completions/` directory
- HTML reports and markdown reports follow patterns: `*.html` and `ebook_mechanic_report_*.md`
- Temporary test data is stored in `test-library/`

## Code Organization

When implementing features:

1. **CLI Framework**: Use a standard CLI framework (cobra, urfave/cli, or similar) for command structure
2. **Command Pattern**: Organize commands in separate packages under `cmd/` directory
3. **Core Logic**: Place business logic in dedicated packages separate from CLI handling
4. **File Processing**: Implement concurrent file processing using goroutines and worker pools for performance
5. **Error Handling**: Return errors up the call stack; handle at appropriate levels with context

## Development Notes

- The project uses `.gitignore` patterns that exclude IDE-specific directories (`.vscode/`, `.idea/`, `.cursor/`, etc.)
- Temporary files and debug artifacts are excluded from version control
- Report files are generated dynamically and not committed
