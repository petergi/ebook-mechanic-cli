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
- `--skip-validate`: skips the post-repair validation pass for faster repairs.
- `--aggressive`: enable aggressive repairs that may drop content or reorder sections.

## Batch Flags

- `--jobs N`: number of concurrent workers for batch operations.
- `--no-backup`: repair in place without creating backups.
- `--aggressive`: enable aggressive repairs for batch runs.

## Notes

Batch operations currently run validation across all matching files.
