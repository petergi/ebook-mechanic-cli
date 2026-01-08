# User Guide

ebook-mechanic-cli is a terminal UI for validating and repairing EPUB and PDF
files. It uses ebook-mechanic-lib for the core validation and repair logic.

## Launch

```bash
./build/ebm-cli
```

To install the CLI:

```bash
make install-cli
```

To uninstall:

```bash
make uninstall-cli
```

## Navigation

- Use `↑/↓` or `j/k` to move.
- Use `enter` to select.
- Use `esc` to go back.
- Use `ctrl+c` to quit.

## Settings

Open Settings from the main menu to adjust:

- Batch job count for concurrent processing.
- Skip post-repair validation to speed up repairs (default: validation enabled).
- Aggressive repair mode (shows a warning when enabled).

## Validate

1. Choose "Validate EPUB/PDF".
2. Select a file to validate.
3. Review the report.

## Repair

1. Choose "Repair EPUB/PDF".
2. Choose how to save the repaired file.
3. Select a file to repair.
4. Review the repair report.

When aggressive repair is enabled, the tool may drop content or reorder sections
to make an EPUB or PDF valid.

## Batch

1. Choose "Batch Process".
2. Select a directory or file.
3. The CLI scans for `.epub` and `.pdf` files and runs batch validation.

## Reports

Reports are rendered in the TUI with styled summaries and issue details.
Original backups use `*_original` when using the backup option (default).
Choose "No backup (in-place)" to repair without creating a backup.

## Related Docs

- [CLI Reference](CLI_REFERENCE.md)
- [Error Codes](ERROR_CODES.md)
- [Architecture](ARCHITECTURE.md)
