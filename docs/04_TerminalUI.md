# 04_TerminalUI.md
# Terminal UI

## Purpose
Defines the terminal UI contract: inputs, outputs, and presentation rules. This document does not define game rules.

## Scope (UI contract only)
- No game rules or order of operations (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).
- For doc boundaries, see `00_Overview.md`.

## UI Model
- UI reads Core state via a read-only view model.
- UI renders Core events into terminal output.
- UI must not contain game rules.

## Startup Flow
- New game:
- Load game:
- Save folder selection (if any):

## Commands (Names and Args Only)
- status
- hire / dismiss
- equip / unequip
- quests
- accept
- retreat
- advance
- save / load (if exposed)

## Output Contracts (Fields, Not Layout)
### Status view
- Fields displayed:

### Quest list view
- Fields displayed:

### Quest detail view (optional)
- Fields displayed:

## Error Handling Conventions
- Validation errors vs rule errors
- File I/O errors policy

## Optional TUI Flows
- Selection-heavy flows (equipment selection, party assembly)
- Contract for selection (confirm/cancel behavior)