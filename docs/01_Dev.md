# Dev

## Purpose
Implementation constraints and technical decisions that affect long-term maintainability and portability.

## Scope (Keep this file small)
- No game rules or processing order (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).
- No UI layout details (see `04_TerminalUI.md`).

## Binary
- Canonical binary name: `coc`

## Project layout (minimal)
- Entry point: `cmd/coc/main.go`
- Core (domain, persistence, save validation): `internal/core/`
- UI (REPL, virtual FS, formatting): `internal/ui/` (or equivalent)
- Docs: `docs/`
- Examples: `examples/`

## Interactive input library
- REPL uses `go-prompt` for line editing, history, and tab completion.
- The virtual FS model remains authoritative; completion candidates are derived from the FS tree.

## Save Boundary
This project is designed so that copying a save folder to another machine reproduces the same game state.

### Save unit
- A “save” is a directory (a save slot).
- The program loads a save by directory path.

### Startup contract
- Canonical invocation: `./coc <save_dir>`
- If `<save_dir>` is not provided: print an actionable message and exit.
- If `<save_dir>` is invalid: print an actionable message and exit.

### Files
Required:
- `clan.json` (authoritative state snapshot)

Optional:
- `quests.json` (active quest snapshot; may be absent in early stages)
- `chronicle.jsonl` (append-only history log; non-authoritative)

### Versioning
- `clan.json` must include `meta.save_version` (integer).
- Unsupported versions hard-fail early stages.

### Sample
- `docs/examples/clan.sample.json`

## Randomness & Reproducibility
- Randomness may occur during generation/initialization.
- Once committed to state, outcomes are immutable.
- A given save state is reproducible when advanced from that state.

## Testing notes
- Cross-machine portability test (copy folder)
- Negative tests for missing/invalid JSON and unsupported versions
- Path resolver tests for virtual FS navigation and listing