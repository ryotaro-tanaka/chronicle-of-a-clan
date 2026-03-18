# 03_Balance.md
# Balance

## Purpose
Numeric tuning only: thresholds, coefficients, curves, and closed-form formulas referenced by `02_System.md`. For doc boundaries, see `00_Overview.md`.

## Scope
- No new mechanics (mechanics belong in `02_System.md`).
- No UI presentation (see `04_TerminalUI.md`).
- No implementation layout except **file/key indirection** for parameters that are numeric tables in JSON.

## Calibration assumption
| Name | Value | Meaning |
|------|-------|---------|
| `PRIMARY_TUNING_PARTY_SIZE` | 4 | Boss hunt math is calibrated mainly for four active party members. |

---

## Daily combat (boss hunt)

### Constants
| Constant | Value |
|----------|-------|
| `BASE_PROGRESS` | 30 |
| `BASE_INJURY` | 1 |
| `SCALE` | 10 |
| `SUCCESS_PROGRESS` | 100 |
| `RETREAT_INJURY_THRESHOLD` | 10 |

### Party inputs (for `N` active members)
Use **arithmetic means** (not sums):

- `party_avg_might   = sum(member_might   + equipment_bonus_might)   / N`
- `party_avg_mastery = sum(member_mastery + equipment_bonus_mastery) / N`
- `party_avg_tactics = sum(member_tactics + equipment_bonus_tactics) / N`
- `party_avg_prot    = sum(armor_prot) / N`

### Boss inputs
Concrete stats after generation: `boss_power`, `boss_guard`, `boss_evasion`, `boss_cunning`.

### Daily progress gain
```
progress_gain = max(0,
  BASE_PROGRESS
  + floor((party_avg_might   - boss_guard)   / SCALE)
  + floor((party_avg_mastery - boss_evasion) / SCALE)
  + floor((party_avg_tactics - boss_cunning) / SCALE)
)
```

### Daily injury gain
```
injury_gain = max(0,
  BASE_INJURY
  + floor((boss_power - party_avg_prot) / SCALE)
)
```

### Accumulation and same-day resolution order
- `progress += progress_gain`
- `injury += injury_gain`
- If `progress >= SUCCESS_PROGRESS` → **success** (checked first).
- Else if `injury >= RETREAT_INJURY_THRESHOLD` → **retreat**.
- Else → **continue**.

---

## Level stat budgets (piecewise linear)

Let `clamp(x, a, b) = min(max(x, a), b)`.

### Monster level budget
Backed by `monster_level_budget_model` in `data/combat/levels.json`.

```
monster_level_budget(L) =
  80
  + 4 * clamp(L - 1,  0, 9)
  + 5 * clamp(L - 10, 0, 10)
  + 6 * clamp(L - 20, 0, 15)
  + 7 * clamp(L - 35, 0, 15)
```

### Member level budget
Backed by `member_level_budget_model` in `data/combat/levels.json`.

```
member_level_budget(L) =
  60
  + 3 * clamp(L - 1,  0, 9)
  + 4 * clamp(L - 10, 0, 10)
  + 5 * clamp(L - 20, 0, 15)
  + 6 * clamp(L - 35, 0, 15)
```

### Reference armor PROT (balance target, not necessarily stored in data)
```
appropriate_armor_prot(L) = floor(monster_level_budget(L) / 4)
```
(intended: same-level armor PROT in the ballpark of same-level boss Power share.)

---

## Data indirection (numeric sources in JSON)

| Parameter / model | File (read-only) | JSON key / location |
|-------------------|------------------|---------------------|
| Monster level budget segments | `data/combat/levels.json` | `monster_level_budget_model` |
| Member level budget segments | `data/combat/levels.json` | `member_level_budget_model` |
| Quest level → monster level range (where used) | `data/combat/levels.json` | `quest_levels[]` |
| Overall rating model (where used) | `data/combat/levels.json` | `overall_rating` |
| Boss profile variation | `data/combat/boss_profiles.json` | `regions.{region}.variation[]` |
| Boss profile level range | `data/combat/boss_profiles.json` | `regions.{region}.profiles[].level_min`, `level_max` |
| Member growth split | `data/combat/member_growth_types.json` | (per growth type) |

Optional later: extract `BASE_PROGRESS`, `BASE_INJURY`, `SCALE`, `SUCCESS_PROGRESS`, `RETREAT_INJURY_THRESHOLD` into e.g. `data/combat_balance.json` — not required for MVP.

---

## Placeholders (other systems)

### Rewards
- `REWARD_*_RANGE`
- `FAILURE_REWARD_POLICY` (optional)

### Guild upkeep / morale / recovery
- `UPKEEP_*`
- `MORALE_*`
- `RECOVERY_*`

### World ranking / rivals
- `RIVAL_UPDATE_INTERVAL_DAYS`
- `RIVAL_FAME_CURVE`
