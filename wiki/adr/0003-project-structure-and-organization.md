# ADR 0003: Project Structure and Organization

## Status

Accepted

## Context

The ebook-mechanic-cli codebase needs a clear organizational structure that:

- Separates concerns between UI, business logic, and utilities
- Supports high test coverage (95%+)
- Makes the codebase easy to navigate and maintain
- Follows Go community best practices
- Allows for future extensibility

### Considered Alternatives

1. **Flat Structure**: All code in root with minimal organization

   - Pros: Simple, no import paths
   - Cons: Poor organization, hard to maintain as project grows

2. **Feature-Based Structure**: Organize by feature (validate/, repair/, batch/)

   - Pros: Related code together
   - Cons: Circular dependencies, harder to test layers independently

3. **Layer-Based Structure**: Organize by technical layer (ui/, operations/, utils/)

   - Pros: Clear separation, testable layers, scalable
   - Cons: More directories, slightly longer import paths

4. **Hexagonal/Clean Architecture**: Strict ports/adapters separation

   - Pros: Very clean separation, highly testable
   - Cons: Overkill for CLI tool, too much indirection

## Decision

We will use a **layer-based structure with internal packages** following Go best practices:

```text
ebook-mechanic-cli/
├── cmd/
│   └── ebm/
│       └── main.go
├── internal/
│   ├── tui/
│   │   ├── models/
│   │   ├── styles/
│   │   └── app.go
│   ├── operations/
│   ├── config/
├── pkg/
│   └── utils/
├── tests/
├── docs/
└── go.mod
```

## Rationale

### Clear Layering

Three distinct layers with clear responsibilities:

1. **TUI Layer** (`internal/tui/`): All Bubbletea models and UI logic
2. **Operations Layer** (`internal/operations/`): Business logic and ebook-mechanic-lib integration
3. **Utilities** (`pkg/utils/`): Reusable helpers (exported as they may be useful externally)

This separation ensures:

- UI can be tested independently of operations
- Operations can be used from different UIs (future: web, API)
- Utilities can be easily unit tested

### Internal Package Pattern

Using `internal/` follows Go best practices:

- Prevents external packages from importing internal code
- Signals these are implementation details
- Provides encapsulation without sacrificing organization

### cmd/ for Entry Points

Standard Go convention:

- `cmd/ebm/main.go` is the CLI entry point
- Keeps main package small (just wiring)
- Allows for future additional commands if needed

### pkg/ for Exportable Utilities

- `pkg/utils/` contains truly reusable code
- Can be imported by external projects if needed
- Clearly signals "this is stable API"

### tests/ for Integration Tests

- Separate from unit tests (kept alongside code)
- Contains end-to-end tests, fixtures, golden files
- Keeps test data organized

### docs/ for Documentation

- Architecture documentation (ARCHITECTURE.md)
- ADRs (adr/)
- API documentation
- User guides

## Consequences

### Positive

1. **Testability**: Each layer can be tested independently
2. **Maintainability**: Clear where to find/add code
3. **Scalability**: Easy to add new models or operations
4. **Go Conventions**: Follows community best practices
5. **Encapsulation**: Internal package prevents misuse

### Negative

1. **Import Paths**: Slightly longer imports (mitigated by clear naming)
2. **Navigation**: More directories to navigate (mitigated by clear structure)

### Mitigation

- Use consistent naming conventions
- Provide clear README.md with structure overview
- Use IDE features for navigation

## Implementation Details

### Detailed Structure

```text
ebook-mechanic-cli/
│
├── cmd/
│   └── ebm/
│       └── main.go              # Entry point, wires up app
│
├── internal/
│   ├── tui/
│   │   ├── models/
│   │   │   ├── menu.go         # Main menu model
│   │   │   ├── browser.go      # File browser model
│   │   │   ├── progress.go     # Progress indicator model
│   │   │   ├── report.go       # Report viewer model
│   │   │   └── repair.go       # Repair preview model
│   │   ├── styles/
│   │   │   └── theme.go        # Lipgloss styles
│   │   └── app.go              # Main Bubbletea application
│   │
│   ├── operations/
│   │   ├── validate.go         # Validation operations
│   │   ├── repair.go           # Repair operations
│   │   └── batch.go            # Batch processing
│   │
│   └── config/
│       └── config.go           # Configuration management
│
├── pkg/
│   └── utils/
│       ├── fileutils.go        # File operations
│       └── formatters.go       # Output formatters
│
├── tests/
│   ├── tui/
│   │   ├── models_test.go      # Model tests
│   │   ├── golden/             # Golden files
│   │   └── fixtures/           # Test fixtures
│   ├── operations/
│   │   └── operations_test.go  # Operations tests
│   └── integration/
│       └── e2e_test.go         # End-to-end tests
│
├── docs/
│   ├── ARCHITECTURE.md         # This document
│   ├── adr/                    # Architecture Decision Records
│   └── USER_GUIDE.md           # User documentation
│
├── .github/
│   └── workflows/              # CI/CD workflows
│
├── CLAUDE.md                   # Claude Code guidance
├── README.md                   # Project overview
├── go.mod                      # Go module definition
└── Makefile                    # Build and test commands
```

### Package Dependencies

```text
cmd/ebm
  └─> internal/tui/app
        ├─> internal/tui/models/*
        ├─> internal/tui/styles
        └─> internal/operations/*
              ├─> pkg/utils/*
              └─> ebook-mechanic-lib
```

**Rules**:

- TUI models should NOT import operations directly
- Operations should NOT import TUI code
- Utils should NOT import anything internal
- Communication via messages (Bubbletea pattern)

### File Naming Conventions

- **Models**: Noun describing the screen/component (e.g., `menu.go`, `browser.go`)
- **Operations**: Verb describing the action (e.g., `validate.go`, `repair.go`)
- **Tests**: `<filename>_test.go` alongside source files
- **Integration tests**: In `tests/` directory with descriptive names

### Import Organization

Standard Go import grouping:

```go
import (
    // Standard library
    "context"
    "fmt"
    "io"

    // External dependencies
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    // Internal packages
    "github.com/petergi/ebook-mechanic-cli/internal/operations"
    "github.com/petergi/ebook-mechanic-cli/internal/tui/styles"

    // Library integration
    "github.com/example/project/pkg/ebmlib"
)
```

## Testing Organization

### Unit Tests

Located alongside source files:

```text
internal/tui/models/menu.go
internal/tui/models/menu_test.go
```

### Integration Tests

In `tests/` directory:

```text
tests/integration/validation_flow_test.go
tests/integration/repair_flow_test.go
tests/integration/batch_processing_test.go
```

### Test Fixtures

Organized by test type:

```text
tests/tui/golden/           # Expected output snapshots
tests/tui/fixtures/         # Test data
tests/operations/fixtures/  # Sample EPUB/PDF files
```

## Migration Path

If structure needs to change:

1. Document reason in new ADR
2. Use Go module replace directives during transition
3. Update imports in phases
4. Verify tests pass at each step

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Blog: Organizing Go Code](https://go.dev/blog/organizing-go-code)
- [Effective Go: Package Names](https://go.dev/doc/effective_go#package-names)
