# CLI Reference

The CLI is a terminal UI. Most actions are driven by the on-screen menu and key
bindings.

## Binary

```sh
ebm-cli
```

## Install

```sh
make install-cli
```

To uninstall:

```sh
make uninstall-cli
```

## Common Flow

1. Launch the CLI.
2. Pick an operation (Validate, Repair, Batch).
3. Select a file or directory.
4. Review results.

## Key Bindings

- `↑/↓` or `j/k`: move selection
- `enter`: select
- `esc`: back
- `ctrl+c`: quit

## Repair Options

Repairs run in-place. Backups are created by default.

- `--no-backup`: repair in place without creating backups.
- `--backup-dir`: directory for `_original` backups.
- `--skip-validation`: skips the post-repair validation pass for faster repairs.
- `--aggressive`: enable aggressive repairs that may drop content or reorder sections.

## Batch Flags

### Performance Options

- `--jobs N`: number of concurrent workers for batch operations.
- `--progress`: progress display mode (`auto`, `simple`, `none`).
- `--summary-only`: only display summary statistics.

### Processing Options

- `--no-backup`: repair in place without creating backups.
- `--aggressive`: enable aggressive repairs for batch runs.
- `--skip-validation`: skip post-repair validation for faster processing.

### Discovery Options

- `--recursive, -r`: process subdirectories recursively [default: true].
- `--max-depth`: maximum directory depth (-1 = unlimited).
- `--ext`: file extensions to include (e.g., `--ext .epub`).
- `--ignore`: glob patterns to exclude from processing.

### Cleanup Options

Control automatic file and directory cleanup after batch operations:

- `--remove-system-errors`: automatically delete files with system errors (IO errors, corrupt archives, permission denied). These files are typically unrecoverable and clutter the library.
- `--move-failed-repairs`: move unrepairable files to an `INVALID` subfolder instead of leaving them in place.
- `--cleanup-empty-dirs`: remove empty parent directories and Calibre metadata-only folders (directories with only `cover.jpg`, `metadata.opf`, etc. but no ebook files) [default: true].

### Cleanup Example

```bash
# Full cleanup mode for a Calibre library
ebm batch repair ~/Books/Calibre \
  --remove-system-errors \
  --move-failed-repairs \
  --cleanup-empty-dirs \
  --jobs 8
```

This will process the library and automatically:

1. Remove files with system-level errors
2. Move failed repairs to `INVALID/` folder
3. Clean up empty directories and Calibre metadata folders
4. Record all actions in the batch report with an "Options" section

## Notes

Batch operations currently run validation across all matching files.
