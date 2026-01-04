# Repository Guidelines

## Project Overview

ebook-mechanic-cli is a CLI tool for ebook processing and management, written in Go.

## Project Structure & Module Organization

- `cmd/ebm/` holds the CLI entrypoint.
- `internal/operations/` contains validation/repair/batch orchestration.
- `internal/tui/` contains the Bubble Tea app coordinator, models, and styles.
- `pkg/` exposes reusable packages.
- `docs/` contains architecture and ADRs.
- Build artifacts go to `build/` and `dist/`.
- Completion scripts are generated into `completions/`.
- HTML reports and markdown reports follow patterns `*.html` and `ebook_mechanic_report_*.md`.
- Temporary test data is stored in `test-library/`.

## Build, Test, and Development Commands

Build:

- `go build -o ebook-mechanic`

Run:

- `go run main.go [command] [flags]`

Testing:

- `go test ./...`
- `go test -cover ./...`
- `go test -run TestName ./path/to/package`
- `go test -v ./...`
- `go test -coverprofile=coverage.out ./...`
- `go tool cover -html=coverage.out`

Linting:

- `golangci-lint run`

Build for Multiple Platforms:

- Linux: `GOOS=linux GOARCH=amd64 go build -o ebook-mechanic-linux`
- macOS: `GOOS=darwin GOARCH=amd64 go build -o ebook-mechanic-darwin`
- Windows: `GOOS=windows GOARCH=amd64 go build -o ebook-mechanic-windows.exe`

## Coding Style & Naming Conventions

- Format with `gofmt`.
- Keep comments minimal and only for complex logic.
- Prefer small, focused packages aligned to operations vs. TUI boundaries.

## Testing Guidelines

- Use `go test ./...` for full runs.
- Keep unit tests fast; avoid slow filesystem/network dependencies.
- Name tests with Go conventions: `TestXxx`, `BenchmarkXxx`.

## Commit & Pull Request Guidelines

- Use Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, `perf:`, `ci:`).
- Keep PRs scoped and include a short summary plus how you tested.

## Code Organization

- Use a standard CLI framework (cobra, urfave/cli, or similar) for command structure.
- Organize commands in separate packages under `cmd/`.
- Place business logic in dedicated packages separate from CLI handling.
- Implement concurrent file processing using goroutines and worker pools for performance.
- Return errors up the call stack; handle at appropriate levels with context.

## Architecture Overview

- TUI uses Bubble Tea for state management and views.
- Operations layer orchestrates validation/repair via the library.

## Development Notes

- `.gitignore` excludes IDE-specific directories (`.vscode/`, `.idea/`, `.cursor/`, etc.).
- Temporary files and debug artifacts are excluded from version control.
- Report files are generated dynamically and not committed.
