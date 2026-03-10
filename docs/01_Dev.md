# Dev

## Purpose
Implementation constraints and technical decisions that affect long-term maintainability and portability.

## Scope (Keep this file small)
- No game rules or processing order (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).
- No UI layout details (see `04_TerminalUI.md`).

## Save Boundary
This project is designed so that copying a save folder to another machine reproduces the same game state.

### Save unit
- A “save” is a directory (a save slot).
- The program loads a save by directory path.

### Startup contract (Stage 1)
- Canonical invocation: `./myapp <save_dir>`
- If `<save_dir>` is not provided: print an actionable message and exit.
- If `<save_dir>` is invalid: print an actionable message and exit.

### Files
Required:
- `clan.json` (authoritative state snapshot)

Optional:
- `quests.json` (active quest snapshot; may be absent in early stages)
- `chronicle.jsonl` (append-only history log; non-authoritative)

### Required vs optional behavior
- If `clan.json` is missing: fail to load with a clear error.
- If optional files are missing: load must still succeed, treating them as empty/not available.

### Versioning
- `clan.json` must include `meta.save_version` (integer).
- If `save_version` is unsupported: fail to load with a clear error.
- Migration strategy is “best effort later”; Stage 1 can hard-fail on unknown versions.

### Atomic write policy
When writing save files:
- Write to `<name>.tmp`, flush, then rename to `<name>`.
- Never partially overwrite a live save file.

## JSON shape (minimal requirements)
`clan.json` must contain at least:
- `meta.save_version`
- `clan` object with fields required by the Status view contract (see `04_TerminalUI.md`)

A sample file lives at:
- `docs/examples/clan.sample.json`

## IDs
- IDs are strings (memberId, equipmentId, questId, monsterId).
- Uniqueness is required within a save.

## Randomness & Reproducibility
- Randomness may occur during generation/initialization (e.g., equipment craft/upgrade results, monster stat rolls at quest start).
- Once committed to state, outcomes are immutable.
- Reproducibility definition: given the same save state, advancing from that state produces the same outcomes (because committed outcomes are stored in the save).

## Core Events (Contract)
- Core emits structured events (no formatted strings).
- UI renders events into terminal output.

## Testing notes
- Save/load round-trip invariants
- Cross-machine portability test (copy folder)
- Negative tests for missing/invalid JSON and unsupported versions