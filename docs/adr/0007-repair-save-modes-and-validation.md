# ADR 0007: Repair Save Modes and Optional Post-Repair Validation

## Status

Accepted

## Context

Repair workflows now need to support different file handling strategies:

- Default behavior should preserve the original file via an `_original` backup and repair in place.
- Users should be able to opt out of backups for faster in-place repairs.
- Post-repair validation can be costly for large batches and needs to be optional.
- Some repair failures require aggressive reconstruction that may drop content.

Both the CLI and the TUI should surface these controls.

## Decision

- Introduce explicit repair save modes: `backup-original` and `no-backup`.
- Default in-place repairs to `backup-original`, with a `no-backup` option for speed.
- Add a CLI flag to skip post-repair validation (`--skip-validate`).
- Add a TUI Settings screen to control batch job count and skip-validation.
- Add an `aggressive` repair option in CLI/TUI with a warning before enabling.

## Consequences

- Reports only reference backups when they are created.
- Users can trade safety for speed with `no-backup` or `--skip-validate`.
- TUI state includes settings that affect batch processing and single repairs.
- Aggressive repair is opt-in and clearly warns about potential data loss.
