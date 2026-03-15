# 03_Balance.md
# Balance

## Purpose
Numeric tuning values only: thresholds, tables, coefficients, and curves referenced by `02_System.md`.

## Scope (Numbers only)
- No new mechanics or state transitions.
- Parameters must be referenced by name from `02_System.md`.
- For doc boundaries, see `00_Overview.md`.

## Naming Conventions
- ALL_CAPS constants
- Group by subsystem (MOMENTUM_*, HIT_*, BREAK_*, PROGRESS_*, INJURY_*, REWARD_*, UPKEEP_*, RECOVERY_*, RIVAL_*)

## Parameter Placeholders (Empty OK Initially)
### Party aggregation
- MIGHT_AGG_*
- MASTERY_AGG_*
- TACTICS_AGG_*

### Momentum
- MOMENTUM_STEP_THRESHOLDS
- MOMENTUM_MAX_STEP_PER_DAY
- MOMENTUM_*_BY_STAGE

### Hit / Break evaluation
- HIT_*_THRESHOLD
- BREAK_*_THRESHOLD

### Progress gain
- QUEST_PROGRESS_COMPLETE
- PROGRESS_GAIN_TABLE
- ATK_TO_PROGRESS_FACTOR
- MIN_PROGRESS_PER_DAY (optional)

### Injury
- POWER_TO_INJURY_FACTOR
- PROT_TO_INJURY_FACTOR
- INJURY_STAGE_THRESHOLDS
- MAX_STAGE_WORSEN_PER_DAY
- AUTO_RETREAT_* (if enabled in System)

### Rewards
- REWARD_*_RANGE
- FAILURE_REWARD_POLICY (optional)

### Guild upkeep / morale / recovery
- UPKEEP_*
- MORALE_*
- RECOVERY_*

### World ranking / rivals
- RIVAL_UPDATE_INTERVAL_DAYS
- RIVAL_FAME_CURVE

### Boss generation (MVP)
- `BOSS_PROFILES_FILE`
  - Path to the boss profile data file.
  - Backed by: `data/boss_profiles.json`.
- `BOSS_PROFILE_VARIATION`
  - Per-region variation rules keyed by the number of focused stats (`focused_stats_count`).
  - Backed by: `regions.{region}.variation[]` in `data/boss_profiles.json`.
- `BOSS_PROFILE_LEVEL_RANGE`
  - Per-profile allowed level range (min and max level for that profile).
  - Backed by: `regions.{region}.profiles[].level_min` and `level_max` in `data/boss_profiles.json`.
- `MONSTER_LEVEL_STAT_BUDGET_MODEL`
  - Converts Monster Level into a total stat budget via base-at-level-1 and per-level increase segments.
  - Backed by: `monster_level_budget_model` in `data/levels.json`.