# Terminal UI

## Purpose
Defines the terminal UI contract: inputs, outputs, and presentation rules. This document does not define game rules.

## Scope (UI contract only)
- No game rules or order of operations (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).

## UI Model
- UI reads Core state via a read-only view model.
- UI renders Core events into terminal output.
- UI must not contain game rules.

## Startup Flow (Stage 1)
- Canonical invocation: `./myapp <save_dir>`
- If `<save_dir>` is missing:
  - print an actionable message (usage + example) and exit
- If `<save_dir>` is invalid:
  - print an actionable error and exit
- If load succeeds:
  - enter an interactive prompt

## Navigation model (Virtual FS)
The interactive session exposes a virtual filesystem-like navigation:

- `ls`
  - lists available entries at the current location
  - each entry includes a type tag:
    - `[DIR]` navigable category
    - `[VIEW]` read-only view command
    - `[ACT]` state-changing action
- `cd <path>`
  - changes the current location (navigation only)
  - supports relative paths and `..` segments
- Commands may be invoked by name or by path:
  - `status`
  - `clan/status`
  - `../exit`

## Commands (Stage 1 minimum)
- `ls`
- `cd <path>`
- `status` (view)
- `exit` (action)

No `help` command in Stage 1.

## Status Output Contract (Fields)
The `status` view displays a stable set of fields. Formatting is flexible; fields are not.

Required fields:
- Save:
  - save directory path
  - save_version
- Clan:
  - clan name
  - current day (in-game day counter)
  - gold
  - fame
- Summary:
  - total members count
  - active quests count
  - equipment counts (weapons count, armor count)
  - in-progress crafting/upgrading count (if present in save; otherwise display 0 or “not available”)

Optional fields (display if available; otherwise omit or show “-”):
- recent chronicle entry count or last entry timestamp (if `chronicle.jsonl` is present)
- last saved timestamp (if present in `clan.json`)

Defaulting rules:
- Missing optional sections must not crash the UI.
- Empty lists must be displayed as 0 counts.

## Error Handling Conventions (Stage 1)
Errors must be actionable and categorized:
- Missing required file: `clan.json` not found
- Invalid JSON: parse error details (file name + reason)
- Unsupported version: show the detected version and supported versions
- Invalid path: directory not found / not a directory

## Exit behavior
- `exit` ends the interactive session.
- Ctrl+C behavior may be added later; Stage 1 only guarantees `exit`.