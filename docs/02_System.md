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

## Key quest availability (listing)
- `key_quest_progress.current_order` in the save controls listing: `quests/keys/` shows only the key quest whose `order` equals `current_order` (the next story quest). `quests/<region>/` shows all key quests with `order <= current_order` in that region (see `04_TerminalUI.md`).

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
  - A Boss Profile has: `id`, `name`, `description`, `level_min`, `level_max`, and a 4-stat ratio pattern over Power, Guard, Evasion, and Cunning.
  - The profile’s `stats` list gives explicit ratios for a subset of stats; any stat not listed shares the remainder \\(1 - R_{focus}\\) equally.
  - For each region, profiles and per-region variation rules are defined in `data/boss_profiles.json`. Variation is chosen by the number of focused stats (`focused_stats_count`) from the region’s `variation[]`.
- Generation flow (caller supplies **profile_id** and optional **seed**):
  1. **Resolve profile**: Look up the profile by id across all regions. If not found, fail.
  2. **Pick level**: Using the seed (or a time-based one if omitted), draw a level uniformly from the profile’s `level_min`..`level_max`.
  3. **Compute total stat budget**: The level is converted into a total stat budget using the level budget model in `03_Balance.md` (backed by `monster_level_budget_model` in `data/levels.json`).
  4. **Derive base stat ratios**: From the profile’s `stats`; remaining stats share the leftover ratio equally.
  5. **Apply variation**: The region’s variation rule for this profile’s focused-stat count perturbs the ratios (e.g. random noise) and they are renormalised to sum to 1.0.
  6. **Scale to concrete stats**: The final ratios and total budget give concrete Power/Guard/Evasion/Cunning.
- The generated boss output includes: profile_id, name, description, region, level, and stats only.