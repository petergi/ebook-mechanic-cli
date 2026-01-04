# ebook-mechanic-cli

A terminal UI for validating and repairing EPUB and PDF ebooks, powered by `ebook-mechanic-lib`.

## Features

- Interactive Bubbletea-based TUI
- EPUB/PDF validation and repair
- Batch processing support
- Styled reports with Lipgloss

## Requirements

- Go 1.25+ (see `go.mod`)
- `ebook-mechanic-lib` dependency (managed via Go modules)

## Install

```bash
go install github.com/petergi/ebook-mechanic-cli/cmd/ebm@latest
```

## Build

```bash
make build
```

The binary is written to `./build/ebm-cli`.

## Run

```bash
./build/ebm-cli
```

## Development

```bash
make test
make docs
make lint
make fmt
```

## Documentation

- `docs/ARCHITECTURE.md`
- `docs/adr/`
- `docs/`

## Wiki

Edit docs under `docs/` and run:

```bash
make wiki-update
```

## License

TBD
