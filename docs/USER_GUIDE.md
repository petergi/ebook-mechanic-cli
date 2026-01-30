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

### Processing Options

- **Batch job count**: Number of concurrent workers for batch processing (default: number of CPU cores).
- **Skip post-repair validation**: Skip validation after repairs to speed up processing (default: validation enabled).
- **No backup mode**: Repair files in-place without creating backups (default: backups enabled).
- **Aggressive repair mode**: Enable aggressive repairs that may drop content or reorder sections (shows a warning when enabled).

### Cleanup Options

- **Remove system errors**: Automatically delete files with system-level errors (corrupt archives, permission issues, etc.).
- **Move failed repairs**: Move unrepairable files to an `INVALID` subfolder for manual review.
- **Cleanup empty directories**: Remove empty parent directories and Calibre metadata-only folders (directories containing only `cover.jpg`, `metadata.opf`, etc. with no ebook files) [default: enabled].

These settings apply to batch operations and are recorded in batch reports.

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

### Batch Reports

Batch reports include:

- **Options Section**: Shows all settings used for the batch operation (workers, validation, backup, cleanup options).
- **Summary Statistics**: Total files processed, repairs attempted/succeeded/failed, system errors.
- **File Lists**: Categorized lists of Invalid, Errored, and Valid files.
- **Cleanup Actions**: Lists files removed (system errors) and moved (failed repairs).

Batch reports can be saved from the TUI (`s` key) and are automatically saved to the `reports/` directory with timestamps.

### Single File Reports

Original backups use `*_original` when using the backup option (default).
Choose "No backup (in-place)" to repair without creating a backup.

## Related Docs

- [CLI Reference](CLI_REFERENCE.md)
- [Error Codes](ERROR_CODES.md)
- [Architecture](ARCHITECTURE.md)
