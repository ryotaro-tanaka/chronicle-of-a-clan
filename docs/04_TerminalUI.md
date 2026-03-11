# Terminal UI

## Purpose
Defines the terminal UI contract: inputs, outputs, and presentation rules. This document does not define game rules.

## Scope (UI contract only)
- No game rules or order of operations (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).

## Startup (CLI) contract
- The program is invoked with a save slot name (not a filesystem path).
- `coc <save_dir>` loads from: `saves/<save_dir>/` and enters the interactive prompt.
- `coc init <save_dir>` creates: `saves/<save_dir>/` from the template and exits (or prints success and exits).

### No-arg usage messages
- `coc` (no args) prints:
  - `Usage: coc <save_dir>`
  - `Usage: coc init <save_dir>`

- `coc init` (no args) prints at least:
  - `Usage: coc init <save_dir>`
  - (Optional) list existing slots under `saves/` if easy

### Slot name rules (user-visible)
- `<save_dir>` accepts only `A-Za-z0-9._-` and must not start with `-`.
- `<save_dir>` is a slot name; paths are not accepted.

## Interactive session model
- Interactive input uses `go-prompt` for history, line editing, and tab completion.
- The UI must not contain game rules.

## Navigation model (Virtual FS)

### Commands
- `ls` / `ls <path>`
  - `ls` lists entries at the current location.
  - `ls <path>` lists entries at `<path>` without changing the current location.
  - entries include type tags:
    - `[DIR]` navigable category
    - `[VIEW]` read-only view
    - `[ACT]` state-changing action
- `cd <path>`
  - navigation only
  - supports `.` and `..`

### Path conventions (virtual FS)
- Accept both `clan` and `clan/` as directory paths.
- Normalize simple segments: `.` and `..`.
- Invocation by path is allowed (examples):
  - `clan/status`
  - `../exit`

## Completion (go-prompt)
- Completion includes:
  - command names: `ls`, `cd`, `status`, `exit`
  - directory paths after `cd ` and `ls `
  - path invocation forms like `clan/status`
- Completion candidates come from the virtual FS tree and current location.

## Commands (Stage 1–3 baseline)
- `ls` / `ls <path>`
- `cd <path>`
- `status` (view)
- `exit` (action)

No `help` command.

## Status Output Contract (compact)
Required lines:
- `Clan: <clan_name>   Day: <day>`
- `Gold: <gold>   Fame: <fame>`
- `Members: <member_count>   ActiveQuests: <active_quest_count>`

## Error handling (minimum)
Errors must be actionable and include the reason.

### CLI-level errors
- Missing args:
  - print usage (including `init`)
- Invalid slot name:
  - explain allowed characters and the leading `-` rule
- Slot not found:
  - `saves/<save_dir>/` does not exist
- Init target exists:
  - `saves/<save_dir>/` already exists

### Save file errors
- Missing required file: `clan.json` not found
- Invalid JSON: parse error details (file name + reason)
- Unsupported version: show detected version and supported versions

## Exit behavior
- `exit` ends the interactive session.
- Ctrl+C handling is optional (may be added later).