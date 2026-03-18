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

## Save Boundary (slot-based)
This project is designed so that copying a save slot directory reproduces the same game state.

### Saves root
- All save slots live under the project-root directory: `saves/`

### Save slot naming
- `coc` accepts a slot name (not a filesystem path).
- Allowed characters: `A-Za-z0-9._-`
- Must not start with `-`
- Any other characters are rejected.

### Startup contract
- Canonical invocation: `./coc <save_dir>`
- `<save_dir>` is a slot name.
- Load path is always resolved as: `saves/<save_dir>/`
- If `<save_dir>` is missing or invalid: print an actionable message and exit.

### Init contract
- Canonical invocation: `./coc init <save_dir>`
- `<save_dir>` is a slot name (same validation rules).
- Creates `saves/<save_dir>/` and copies the template contents into it.
- If the target slot already exists: fail (no overwrite by default).

### Template directory
- Init template lives at: `examples/save_init_template/`
- Must include at least: `examples/save_init_template/clan.json`

### Save files
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
- Early stages may hard-fail on unknown versions.

### Atomic write policy
When writing save files:
- Write to `<name>.tmp`, flush, then rename to `<name>`.
- Never partially overwrite a live save file.

## Data (read-only at runtime)
Paths are resolved from the **repository root** (working directory). Game rules and formulas that *use* this data are in `02_System.md` / `03_Balance.md`.

### `data/quests/key_quests.json`
- Key quest list: `order` → boss `profile_id`. Drives quest listing with save `key_quest_progress.current_order` (see `04_TerminalUI.md`).

### `data/combat/levels.json`
- `monster_level_budget_model` — segments for total monster stat budget by level.
- `member_level_budget_model` — segments for total member stat budget by level.
- Other keys (e.g. `quest_levels`, `overall_rating`) as referenced by Core.

### `data/combat/boss_profiles.json`
- Per-region `profiles[]` (`id`, `name`, `description`, `level_min`, `level_max`, `stats` ratios) and `variation[]`. Used for boss generation; **no** `material_id` on profiles (materials reverse-link via `source_profile_id`).

### `data/combat/member_growth_types.json`
- Maps growth type → distribution of raw member budget into Might / Mastery / Tactics / Survival.

### `data/items/materials.json`
- Material entries; each may include `source_profile_id` linking to the boss profile that drops or defines that material.

### `data/items/equipments/rental_equipment.json` / `crafted_equipment.json`
- Rental and crafted weapon/armor definitions. Crafted recipes reference `material_id` in `materials.json`.

## JSON shape (minimal requirements)
`clan.json` must contain at least:
- `meta.save_version`
- `clan` object with fields required by the Status view contract (see `04_TerminalUI.md`)

Optional:
- `key_quest_progress.current_order` (integer): gates which key quests are available for listing; if missing, treated as 1.

## Build artifacts
- Build outputs are placed under `bin/` (and `bin/` is ignored by git).

## Testing notes
- Cross-machine portability test (copy `saves/<slot>/`)
- Negative tests for missing/invalid JSON and unsupported versions
- Slot name validation tests (invalid chars, leading `-`)
- Init tests (create, already exists, invalid name)