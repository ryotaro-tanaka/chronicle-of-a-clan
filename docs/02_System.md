# System.md

## Scope Notes
- This document defines **what happens** and **in what order**.
- Numeric thresholds, tables, coefficients, and tuning values live in **Balance.md**.
- Current MVP focus: **Boss Hunt quests**. Exploration/Gathering can be added later without changing the core loop.

---

## Core Entities

### Member
A recruitable unit the player can assign to quests.

**Core fields**
- Identity: `id`, `name`
- Progression: `level` (growth distribution is out of scope here)
- Stats (combat): `Might`, `Mastery`, `Tactics`
- Stats (non-combat): `Survival`
- Status: `InjuryStage` (abstract stage), `MoraleState` (optional; for unpaid upkeep effects)
- Equipment: `weaponId?`, `armorId?`

**Responsibilities**
- `Might` contributes to breaking through boss armor (Guard).
- `Mastery` contributes to landing effective hits against boss evasion (Evasion).
- `Tactics` drives the battle flow over days by changing Momentum against boss Cunning.
- `Survival` is **out-of-quest**: recovery efficiency, non-combat counterplay, and later exploration/camping.

---

### Boss Monster
The target of a Boss Hunt quest.

**Core fields**
- Identity: `id`, `name`
- Stats: `Power`, `Guard`, `Evasion`, `Cunning`
- Traits: `traits[]` (extensible list; MVP may use none)

**Responsibilities**
- `Power` drives injury pressure on the party each day.
- `Guard` resists `Might` (break/penetration progress).
- `Evasion` resists `Mastery` (hit effectiveness).
- `Cunning` resists `Tactics` and determines Momentum direction over days.

---

### Equipment

#### Weapon
**Core fields**
- `id`, `name`
- `req`: requires **either** `Might` or `Mastery` (weapon-defined)
- `ATK` (fixed numeric value; no traits in MVP)
- Storage ownership: weapons belong to the clan (warehouse), assigned to members.

**Responsibilities**
- ATK is used as part of daily quest progress gain.
- Weapons differ only by **requirement + ATK** in MVP.

#### Armor
**Core fields**
- `id`, `name`
- `req`: `Might` requirement (MVP)
- `PROT` (fixed numeric value; no traits in MVP)
- Storage ownership: armor belongs to the clan (warehouse), assigned to members.

**Responsibilities**
- PROT reduces/mitigates injury progression caused by boss Power.
- Armors differ only by **requirement + PROT** in MVP.

---

### Quest (Boss Hunt)
A multi-day quest resolved in daily ticks until cleared, retreated, or failed.

**Core fields**
- Identity: `id`, `name`, `regionId`
- Party: `memberIds (1..4)`, `bossId`
- State: `Active | Cleared | Retreated | Failed`
- Progress: `Progress` (0..100)
- Battle Flow: `MomentumStage` in **{-2, -1, 0, +1, +2}**
- Day counter: `daysElapsed`

**Rules**
- **Momentum persists for the entire quest**, changing day-by-day.
- **Retreat yields zero rewards** (Gold/Fame/Materials all zero).
- Clear condition: Progress reaches completion threshold (defined in Balance; conceptually “100%”).

---

## Time Progression

### `advance Xd`
Advancing time runs the **Daily Tick** X times. Each day processes:
1) Guild out-of-quest tick (upkeep, recovery, rank checks, etc.)
2) Each active quest daily resolution (boss hunt loop)

> Ordering can be swapped if needed, but must be consistent and deterministic.

### Determinism
- Daily outcomes are deterministic given the same saved state and seed.
- Any randomness (e.g., name generation, shop inventory later) must be seeded and recorded or derived deterministically.

---

## Daily Tick Order (High Level)

### A) Guild Tick (Out of Quest)
1. Upkeep collection attempt
2. Morale update from unpaid upkeep (if applicable)
3. Recovery tick for members not on quests (and optionally those who returned today)
4. Rank / hiring-cap consistency checks
5. (Optional later) shop refresh, events, rival guild update tick

### B) Quest Tick (For Each Active Boss Hunt)
Runs the **Boss Hunt Daily Resolution Loop** (below).

---

## Boss Hunt Daily Resolution Loop (Per Day)

> This is the core “quest combat” loop.

### Step 1 — Momentum Update (Tactics vs Cunning)
- Compute party tactical capability vs boss cunning capability.
- Update `MomentumStage` by at most a small number of steps per day (tuned in Balance).
- Momentum is clamped to [-2, +2].
- Momentum influences the effectiveness of the party’s actions and the boss’s pressure (exact effects tuned in Balance).

### Step 2 — Hit Evaluation (Mastery vs Evasion)
- Determine a **Hit Grade** (e.g., Fail / Normal / Great) based on party Mastery vs boss Evasion, adjusted by Momentum.
- Hit Grade represents how well attacks connect and convert into meaningful progress.

### Step 3 — Break Evaluation (Might vs Guard)
- Determine a **Break Grade** (e.g., Fail / Normal / Great) based on party Might vs boss Guard, adjusted by Momentum.
- Break Grade represents how much the party can exploit openings / penetrate defenses.

### Step 4 — Progress Gain
- Compute `ProgressGain` from:
  - Weapon ATK values in the party
  - Hit Grade
  - Break Grade
  - MomentumStage
- Apply `Progress += ProgressGain` (linear accumulation).
- If progress reaches completion threshold => quest becomes **Cleared**.

### Step 5 — Counterattack & Injury
- Boss applies injury pressure based on `Power`, adjusted by Momentum.
- Armor PROT mitigates this pressure.
- Members’ `InjuryStage` may worsen (abstract stage progression; tuning in Balance).
- Severe injury can force retreat/Failure conditions (policy in Balance, rule hook here).

### Step 6 — Continue / Retreat Decision
- If the party is in a critical condition (e.g., too many members at severe stages), the quest may:
  - Continue automatically, or
  - Auto-retreat, or
  - Require a player decision at next interaction (design choice; recommended: auto-policy + manual override later).
- If retreat occurs: state => **Retreated**, rewards => **zero**.

---

## Party Aggregation (How Party Stats Are Derived)
System defines *what* is aggregated; Balance defines the *weights*.

- `PartyMight` is derived from members’ Might (aggregation method defined in Balance).
- `PartyMastery` is derived from members’ Mastery (aggregation method defined in Balance).
- `PartyTactics` is derived from members’ Tactics (aggregation method defined in Balance).
  - Recommended to reflect “one strong tactician helps, but not infinitely.”

---

## Difficulty Display (UI-Facing Abstractions)
Quests display:
- **Threat**: primarily driven by boss Power (injury danger).
- **Toughness**: primarily driven by boss Guard (break difficulty).
- **Trickiness**: primarily driven by boss Evasion and/or Cunning (hit and flow difficulty).
- **Overall**: a combined rank/grade derived from the above.

Exact threshold mapping is in Balance.md.

---

## Rewards
On **Clear**:
- `Gold`
- `Fame`
- `Materials`

On **Retreat**:
- **No rewards** (all zero).

On **Failure**:
- MVP can treat as equivalent to Retreat (no rewards), unless later differentiated.

Reward ranges and scaling are in Balance.md.

---

## Guild Management (Out of Quest)

### Guild Rank & Hiring Cap
- Hiring cap is determined by `GuildRank`.
- The player cannot exceed the cap; must rank up to hire more members.

### Hiring / Dismissal
- Hiring creates a new member (name generation is deterministic).
- Dismissal removes a member from the roster (equipment returns to storage).

### Upkeep & Non-Payment Effects
- Each member incurs daily upkeep; higher level increases upkeep.
- If upkeep cannot be paid, MoraleState worsens (and may trigger penalties/events later).
- System specifies the state transitions; Balance defines the numbers.

### Recovery & Rest
- Injury does not instantly reset after quests.
- `Survival` affects recovery efficiency during rest.
- Recovery occurs in daily ticks.

---

## Rival Guilds (Result-Only)
- The game periodically updates the “world ranking” via deterministic event snapshots (no simulation).
- Player win condition: **Fame rank #1**.
- Update frequency and rival progression curves are in Balance.md.

---

## Open Questions (Intentionally Deferred)
- Exploration / gathering quest types
- Shops/crafting details
- Trait system for bosses and equipment
- Player-driven retreat decisions vs auto-policy
