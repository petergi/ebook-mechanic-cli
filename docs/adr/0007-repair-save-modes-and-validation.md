# ADR 0007: Repair Save Modes and Optional Post-Repair Validation

## Status

Accepted

## Context

Repair workflows now need to support different file handling strategies:

- Default behavior should preserve the original file via an `_original` backup and repair in place.
- Users should be able to opt out of backups for faster in-place repairs.
- Post-repair validation can be costly for large batches and needs to be optional.

Both the CLI and the TUI should surface these controls.

## Decision

- Introduce explicit repair save modes: `backup-original`, `rename-repaired`, and `no-backup`.
- Default in-place repairs to `backup-original`.
- Add a CLI flag to skip post-repair validation (`--skip-validate`).
- Add a TUI Settings screen to control batch job count and skip-validation.

## Consequences

- Repair results now distinguish between backups and repaired artifacts in reports.
- Users can trade safety for speed with `no-backup` or `--skip-validate`.
- TUI state includes settings that affect batch processing and single repairs.
