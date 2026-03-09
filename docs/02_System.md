# 02_System.md
# System

## Purpose
Defines what happens and in what order in the Core system.

## Scope (Keep this file rules-only)
- No tuning numbers (see `03_Balance.md`).
- No UI command/output formatting (see `04_TerminalUI.md`).
- For doc boundaries, see `00_Overview.md`.

## Terminology (Minimal)
- “Monster” is the canonical term.
- Monsters are classified into minor monsters and boss monsters.

## Time & Processing Model (Invariants)
- Time unit: day
- `AdvanceDays(days)` runs the Daily Tick `days` times.
- Daily Tick order (fixed):
  1) Guild/Out-of-quest Tick
  2) Quest Tick for each active quest
- Randomness & reproducibility:
  - Randomness may occur during equipment crafting/upgrades and monster stat initialization at quest start.
  - Once committed into state, outcomes are immutable.
  - A given save state is reproducible when advanced from that state.

## State Machines (Lifecycle Specs)
### Quest lifecycle
- States and transitions:
- Persistence requirements:

### Member lifecycle
- Injury stages and recovery hooks:
- Availability rules:

### Equipment lifecycle
- Ownership and assignment model:
- Craft/upgrade in-progress → completed model:

## Systems (Order of Operations)
### Guild Tick (Out-of-quest)
- Upkeep attempt
- Morale update hooks
- Recovery tick hooks
- Consistency checks (rank/hiring cap)
- List refresh hooks (may change over time; examples only)

### Quest Tick (Boss Hunt)
- Daily resolution steps (names only; tuning references by parameter name):
  1) Momentum update
  2) Hit evaluation
  3) Break evaluation
  4) Progress gain
  5) Counterattack & injury
  6) Continue/retreat/fail decision hook

### World Tick (Ranking snapshots)
- Result-only updates and persistence rules:

## Aggregation & Derived Values
- Party aggregation responsibilities (what is derived, not weights)
- Difficulty descriptors as derived values (no thresholds here)

## Rewards & Outcomes (Rule-level)
- Outcome types and reward categories
- Retreat yields zero rewards (rule)
- Numeric ranges live in `03_Balance.md`

## Core Events & Logs (Contract)
- Event categories emitted by Core
- Log generation boundary (Core emits structured events; UI renders)