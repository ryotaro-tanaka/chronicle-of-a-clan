# 01_Dev.md
# Dev

## Purpose
Implementation constraints and technical decisions.

## Scope (Keep this file small)
- No game rules (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).
- No UI formatting (see `04_TerminalUI.md`).
- For doc boundaries, see `00_Overview.md`.

## Target Platform
- Language: Go
- Runs locally in terminal on Windows/Linux

## Save/Load
- Format: JSON
- Save folder layout:
- Atomic save strategy (write temp → fsync → rename):
- `save_version` and migration policy:

## IDs
- ID types (memberId, equipmentId, questId, monsterId):
- ID generation policy:

## Randomness & Reproducibility
- Randomness exists during generation/initialization.
- Once committed to state, outcomes are immutable.
- Reproducibility is defined as: a given save state reproduces the same state and future processing outcomes when advanced from that state.

## Core Events (Contract)
- Core emits structured events (no formatted strings).
- UI renders events into terminal output.

## Testing Notes
- Save/load round-trip invariants
- Cross-machine portability test
- Invariant tests for processing order and state validity