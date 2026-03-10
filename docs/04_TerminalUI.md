# Terminal UI

## Purpose
Defines the terminal UI contract: inputs, outputs, and presentation rules. This document does not define game rules.

## Scope (UI contract only)
- No game rules or order of operations (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).

## Startup Flow
- Canonical invocation: `./coc <save_dir>`
- If `<save_dir>` is missing or invalid: print an actionable message and exit.
- If load succeeds: enter an interactive prompt.

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

### Path conventions
- Accept both `clan` and `clan/` as directory paths.
- Normalize simple segments: `.` and `..`.

## Line editing and completion (go-prompt)
- History and tab completion are supported.
- Completion includes:
  - command names: `ls`, `cd`, `status`, `exit`
  - directory paths after `cd ` and `ls `
  - path invocation forms like `clan/status`
- Completion candidates come from the virtual FS tree and current location.

## Status Output Contract (compact)
Required lines:
- `Clan: <clan_name>   Day: <day>`
- `Gold: <gold>   Fame: <fame>`
- `Members: <member_count>   ActiveQuests: <active_quest_count>`

## Exit behavior
- `exit` ends the interactive session.