# Balance.md

## Principles
- This document defines **how much** things happen: thresholds, tables, coefficients, and curves.
- If you want the game to feel faster/slower, harder/easier, richer/poorer—change it here first.
- Avoid adding new mechanics here; mechanics belong in System.md.

---

## Global Targets (Design Intent)
- **Specialists are stronger when matched**, but can struggle badly when mismatched.
- **Tactics can turn “nearly impossible” into “possible over time”** by shifting Momentum.
- Retreat yields **zero rewards**, so risk assessment matters.
- Upkeep should discourage infinite hiring, but not become constant micromanagement.

---

## Party Aggregation Weights
Define how party stats are derived from 1–4 members.

### PartyMight Aggregation
- Method: (choose one)
  - Sum
  - Average
  - Weighted (top 2 heavier)
- Tunable weights: `MIGHT_AGG_WEIGHTS`

### PartyMastery Aggregation
- Method and weights: `MASTERY_AGG_WEIGHTS`

### PartyTactics Aggregation (Recommended)
- Suggested baseline (editable):
  - `PartyTactics = w_max * max(Tactics) + w_avg * avg(Tactics)`
- Tunables: `TACTICS_W_MAX`, `TACTICS_W_AVG`

---

## Momentum System (Stage: -2..+2)

### Daily Momentum Step Rules
- Define how `Δ = PartyTactics - BossCunning` maps to Momentum step changes.
- Tunables:
  - `MOMENTUM_STEP_THRESHOLDS` (Δ ranges)
  - `MOMENTUM_MAX_STEP_PER_DAY` (e.g., 1 or 2)
  - Clamp range fixed: [-2, +2]

### Momentum Effects (What It Modifies)
Momentum influences:
- Hit evaluation (Mastery vs Evasion)
- Break evaluation (Might vs Guard)
- ProgressGain
- Injury pressure / mitigation

Tunables:
- `MOMENTUM_HIT_BONUS_BY_STAGE`
- `MOMENTUM_BREAK_BONUS_BY_STAGE`
- `MOMENTUM_PROGRESS_MULT_BY_STAGE`
- `MOMENTUM_INJURY_MULT_BY_STAGE`

---

## Hit Evaluation (Mastery vs Evasion)

### Hit Grade Definitions
Define Hit Grades used by System.md:
- `Fail`
- `Normal`
- `Great`

### Thresholds
- Based on `HitDelta = PartyMastery - BossEvasion`, adjusted by Momentum.
Tunables:
- `HIT_FAIL_THRESHOLD`
- `HIT_GREAT_THRESHOLD`
- (Anything between is Normal)

---

## Break Evaluation (Might vs Guard)

### Break Grade Definitions
- `Fail`
- `Normal`
- `Great`

### Thresholds
- Based on `BreakDelta = PartyMight - BossGuard`, adjusted by Momentum.
Tunables:
- `BREAK_FAIL_THRESHOLD`
- `BREAK_GREAT_THRESHOLD`

---

## Progress Gain (Linear Accumulation)

### Completion Threshold
- Default concept: 100.
Tunable:
- `QUEST_PROGRESS_COMPLETE` (keep as 100 unless you have a strong reason)

### Progress Gain Table
ProgressGain is determined by:
- Hit Grade
- Break Grade
- Party weapon ATK contribution
- Momentum stage modifiers

Tunables:
- `PROGRESS_GAIN_TABLE[HitGrade][BreakGrade]` (base gains)
- `ATK_TO_PROGRESS_FACTOR` (how weapon ATK adds on top, or scales base)
- `MIN_PROGRESS_PER_DAY` (optional; for avoiding hard-stalls if desired)

> Keep this table small and readable. This is one of the most important “feel” controls.

---

## Injury System (Boss Power vs Armor PROT)

### Injury Stages
Stages referenced by System.md:
- `Healthy`
- `Injured`
- `Severe`
- `Down`

### Daily Injury Pressure
Driven by:
- Boss Power
- Momentum injury multiplier
- Armor PROT mitigation

Tunables:
- `POWER_TO_INJURY_FACTOR`
- `PROT_TO_INJURY_FACTOR`
- `INJURY_STAGE_THRESHOLDS` (how pressure maps to stage change)
- `MAX_STAGE_WORSEN_PER_DAY` (cap to prevent instant wipe if desired)

### Quest Failure / Auto-Retreat Policy (If Enabled)
Tunables:
- `AUTO_RETREAT_ON_DOWN_COUNT` (e.g., retreat if >=2 members Down)
- `FAIL_ON_ALL_DOWN` (bool)
- `ALLOW_CONTINUE_WHEN_DOWN` (bool; typically false for MVP)

---

## Equipment Generation Ranges (MVP: No Traits)

### Weapons
- Requirements: `ReqStat` ∈ {Might, Mastery}
- Tunables:
  - `WEAPON_ATK_RANGE_BY_TIER`
  - `WEAPON_REQ_RANGE_BY_TIER`

### Armor
- Requirements: `ReqMight`
- Tunables:
  - `ARMOR_PROT_RANGE_BY_TIER`
  - `ARMOR_REQ_RANGE_BY_TIER`

> You can keep tiers implicit for MVP and just define “early/mid/late” ranges.

---

## Quest Rewards (On Clear)

### Reward Types
- Gold
- Fame
- Materials

### Reward Ranges
Tunables (by quest tier / region / boss tier):
- `REWARD_GOLD_RANGE`
- `REWARD_FAME_RANGE`
- `REWARD_MATERIALS_RANGE`

### Retreat / Failure Rewards
- Retreat: **always 0** (fixed rule; not tunable here)
- Failure: recommend same as Retreat for MVP (0)

---

## Guild Rank & Hiring Cap

### Rank Requirements
Tunables:
- `RANK_UP_FAME_REQUIREMENTS` (cumulative Fame thresholds)
- (Optional) `RANK_UP_REGION_REQUIREMENTS`

### Hiring Cap by Rank
Tunables:
- `HIRING_CAP_BY_RANK`

---

## Upkeep & Morale (Out of Quest)

### Upkeep Growth Curve
Tunables:
- `UPKEEP_BASE_PER_MEMBER`
- `UPKEEP_PER_LEVEL_FACTOR` (or table by level bands)

### Non-Payment Penalties
Define how quickly morale worsens and what it affects.
Tunables:
- `MORALE_DROP_PER_UNPAID_DAY`
- `MORALE_PENALTY_EFFECTS` (e.g., reduced progress gain, increased injury risk, or hiring friction)

> Keep penalties noticeable but not oppressive. The goal is to prevent infinite hiring, not force constant penny-pinching.

---

## Recovery (Survival)

### Recovery Rules
Survival influences recovery rate outside quests.
Tunables:
- `BASE_RECOVERY_PER_DAY`
- `SURVIVAL_RECOVERY_BONUS_FACTOR`
- `MAX_RECOVERY_PER_DAY` (cap)

---

## Difficulty Display Mapping (UI Ranks)

### Threat / Toughness / Trickiness Buckets
Mapping from boss stats to UI labels (Low/Med/High/etc.).
Tunables:
- `THREAT_BUCKETS` (based on Power)
- `TOUGHNESS_BUCKETS` (based on Guard)
- `TRICKINESS_BUCKETS` (based on Evasion/Cunning)

### Overall Rank
Combine the three into a single grade (e.g., E/D/C/B/A/S).
Tunables:
- `OVERALL_RANK_RULE` (max-of-three, weighted sum, etc.)
- `OVERALL_RANK_THRESHOLDS`

---

## Rival Guilds (Result-Only)

### Update Frequency
Tunables:
- `RIVAL_UPDATE_INTERVAL_DAYS`

### Rival Fame Growth
Tunables:
- `RIVAL_FAME_CURVE` (simple curve or table by time)

---

## Playtest Checklist (Balance Focus)
- Can a mismatched specialist feel “nearly impossible” unless Momentum swings?
- Does a strong tactician meaningfully swing Momentum over multiple days?
- Are clears fast enough to feel rewarding but slow enough to feel like a campaign?
- Is retreat painful (0 rewards) but not so punishing that players never take risks?
- Does upkeep discourage infinite hiring without becoming a constant burden?