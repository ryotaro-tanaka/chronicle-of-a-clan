# 02_System.md
# System

## Purpose
Defines what happens and in what order in the Core system.

## Scope (Keep this file rules-only)
- No tuning numbers or closed-form formulas (see `03_Balance.md`).
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
- **MVP intent**: Crafted equipment is the long-term progression path; rental gear is deliberately weaker. Weapons provide modest combat bonuses; armor provides the main survivability lever (PROT) in boss hunts.

## Systems (Order of Operations)
### Guild Tick (Out-of-quest)
- Upkeep attempt
- Morale update hooks
- Recovery tick hooks
- Consistency checks (rank/hiring cap)
- List refresh hooks (may change over time; examples only)

### Quest Tick (Boss Hunt)
MVP uses **no Momentum**. Tactics affects **daily progress** against boss Cunning (see `03_Balance.md` for formulas and constants).

Daily steps (conceptual):
1. Derive party **average** Might, Mastery, and Tactics (member stats plus applicable equipment bonuses).
2. Derive party **average** armor PROT.
3. Compute daily **progress gain** from those averages vs boss Guard, Evasion, and Cunning.
4. Compute daily **injury gain** from boss Power vs average PROT.
5. Accumulate quest progress and injury; apply same-day outcome order (**success** if progress threshold met, else **retreat** if injury threshold met, else continue)—exact thresholds in `03_Balance.md`.

### World Tick (Ranking snapshots)
- Result-only updates and persistence rules:

## Boss hunt combat (MVP rules)

### Stat roles (member)
- **Might** — contributes to daily progress; opposed by boss **Guard**.
- **Mastery** — contributes to daily progress; opposed by boss **Evasion**.
- **Tactics** — contributes to daily progress; opposed by boss **Cunning**.
- **Survival** — does **not** feed the daily boss-hunt combat formula in MVP (out-of-quest / future systems).
- **PROT** (armor) — reduces daily injury; opposed by boss **Power**.

### Stat roles (boss)
- **Guard / Evasion / Cunning** — resist party progress from Might / Mastery / Tactics respectively.
- **Power** — drives injury against party PROT.

### Budget and allocation (conceptual order)
- **Boss**: Total monster stat budget from level (`03_Balance.md`, `data/combat/levels.json`) → split into Power/Guard/Evasion/Cunning using boss profile ratios and variation (`data/combat/boss_profiles.json`).
- **Member**: Total member stat budget from level → split into Might/Mastery/Tactics/Survival using growth type (`data/combat/member_growth_types.json`) → then apply equipment bonuses from item data (`data/items/…`).
- **Crafting**: Materials link to source boss via `source_profile_id` in `data/items/materials.json`; crafted recipes reference `material_id`. Boss profiles do **not** need embedded `material_id`.

## Aggregation & Derived Values
- For boss hunts, party combat inputs use **per-member averages** over active participants (see `03_Balance.md`).
- Difficulty descriptors as derived values (no thresholds in this file).

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
  - For each region, profiles and per-region variation rules are defined in `data/combat/boss_profiles.json`. Variation is chosen by the number of focused stats (`focused_stats_count`) from the region’s `variation[]`.
- Generation flow (caller supplies **profile_id** and optional **seed**):
  1. **Resolve profile**: Look up the profile by id across all regions. If not found, fail.
  2. **Pick level**: Using the seed (or a time-based one if omitted), draw a level uniformly from the profile’s `level_min`..`level_max`.
  3. **Compute total stat budget**: The level is converted into a total stat budget per `03_Balance.md` (`monster_level_budget_model` in `data/combat/levels.json`).
  4. **Derive base stat ratios**: From the profile’s `stats`; remaining stats share the leftover ratio equally.
  5. **Apply variation**: The region’s variation rule for this profile’s focused-stat count perturbs the ratios (e.g. random noise) and they are renormalised to sum to 1.0.
  6. **Scale to concrete stats**: The final ratios and total budget give concrete Power/Guard/Evasion/Cunning.
- The generated boss output includes: profile_id, name, description, region, level, and stats only.
