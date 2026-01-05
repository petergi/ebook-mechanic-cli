# ADR 0006: Dual-Mode Architecture (CLI + TUI)

## Status

Accepted

## Context

The initial goal was to build a TUI for ebook validation. However, users often require scriptable interfaces for automation (CI/CD pipelines, batch scripts) alongside interactive tools for exploration. Maintaining two separate binaries (one for TUI, one for CLI) introduces code duplication and release complexity.

We need a way to support both use cases in a single application while keeping the codebase maintainable and the user experience intuitive.

## Decision

We will implement a **Dual-Mode Architecture** where the application behaves differently based on how it is invoked:

1. **TUI Mode (Default)**: Running the application without arguments (or with TUI-specific flags) launches the interactive Bubbletea interface.
2. **CLI Mode**: Running the application with arguments (subcommands like `validate`, `repair`) or file paths launches the Cobra-based command-line interface.

## Implementation

- **Entry Point**: `cmd/ebm/main.go` detects arguments.
  - If `len(os.Args) == 1`: Launch TUI (`tui.Run()`).
  - Else: Execute Cobra root command (`cli.Execute()`).
- **CLI Framework**: Use `spf13/cobra` for robust command and flag parsing.
- **Shared Logic**: Business logic (validation, repair, batch processing) is moved to `internal/operations` to be consumed by both `internal/tui` and `internal/cli`.
- **Bypass Mode**: The CLI root command is configured to accept arbitrary arguments. If a file or directory path is provided as the first argument, it defaults to a `validate` operation, allowing for quick checks (`ebm book.epub`).

## Consequences

### Positive

- **Flexibility**: Satisfies both interactive users and automation needs.
- **Single Binary**: Simplifies distribution and updates.
- **Code Reuse**: Core logic is shared, reducing bugs and maintenance effort.
- **Discoverability**: Users can easily switch between modes using the same tool.

### Negative

- **Complexity**: The `main.go` entry point has slightly more logic to dispatch modes.
- **Dependencies**: Requires both TUI (Bubbletea) and CLI (Cobra) libraries in the final binary, slightly increasing size.

## Alternatives Considered

1. **Separate Binaries**: `ebm-tui` and `ebm-cli`.
    - *Rejected*: Higher maintenance burden, confusing for users.
2. **Flags for Mode**: `ebm --tui` vs `ebm --cli`.
    - *Rejected*: Less intuitive. Defaulting to TUI on no-args is a standard pattern for modern terminal tools (e.g., `lazygit`).
