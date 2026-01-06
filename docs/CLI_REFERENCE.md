# CLI Reference

The CLI is a terminal UI. Most actions are driven by the on-screen menu and key
bindings.

## Binary

```sh
ebm-cli
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

## Repair Save Modes

In-place repairs can control where the repaired file lands.

- `--save-mode backup-original`: creates `*_original` backups and repairs the original filename (default for in-place repairs).
- `--save-mode rename-repaired`: writes `*_repaired` and keeps the original untouched.

## Notes

Batch operations currently run validation across all matching files.
