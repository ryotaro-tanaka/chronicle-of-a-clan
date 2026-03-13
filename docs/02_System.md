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

## Boss monster generation (MVP)

- Boss monsters are generated from data-driven **Boss Profiles** and **level models**:
  - A Boss Profile defines a 4-stat ratio pattern over Power, Guard, Evasion, and Cunning, plus:
    - a fixed **Rank** (tier, e.g. forest Rank 1–5) used for progression,
    - a display **Name**, and
    - a per-region selection **weight**.
  - For each region, a finite set of Boss Profiles and variation rules are defined in `data/boss_profiles.json`.
- Generation flow (high level):
  1. **Pick Quest Level**: The caller provides a Quest Level (integer).
  2. **Pick Monster Level**: The Quest Level is mapped to an allowed Monster Level range via a table in `03_Balance.md` (backed by `data/levels.json`). A specific Monster Level is drawn uniformly from that range.
  3. **Compute total stat budget**: The Monster Level is converted into a total stat budget (sum of Power/Guard/Evasion/Cunning) using a level budget model defined in `03_Balance.md` (backed by `monster_level_budget_model` in `data/levels.json`).
  4. **Select Boss Profile**: For the chosen region, one Boss Profile is selected at random using its weight as a selection probability (relative to the other profiles in that region).
  5. **Derive base stat ratios**:
     - The profile’s `stats` list specifies a subset of stats with explicit ratios; their sum is \\(R_{focus}\\).
     - Any remaining stats not listed in `stats` share the leftover \\(1 - R_{focus}\\) equally.
  6. **Apply variation**:
     - The number of focused stats (length of `stats`) is used to pick a variation rule for that region.
     - The base ratios are perturbed according to this variation rule (e.g., small random noise) and then renormalised so that the four ratios still sum to 1.0.
  7. **Scale to concrete stats**: The final ratios and the total stat budget determine the concrete Power/Guard/Evasion/Cunning values.
  8. **Compute Overall**:
     - Using the Monster Level and the Quest Level’s range, a relative position within the range is computed.
     - This position is mapped to an integer Overall rating in \\([1..N]\\), where \\(N\\) is configured in `03_Balance.md` (backed by `overall_rating` in `data/levels.json`).
- **Rank vs Overall**:
  - **Rank** is a fixed tier attached to each Boss Profile (e.g. forest Rank1 = early boss, Rank5 = top boss). It is used for progression/unlock logic (outside the scope of this MVP stage).
  - **Overall** is a per-instance measure of how strong this particular roll is **within its allowed Quest/Monster Level range** (e.g. under-tuned vs over-tuned for that quest). It is derived from the Monster Level and the Quest Level’s range, not fixed per boss.