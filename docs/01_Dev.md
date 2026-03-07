# Dev.md

## Purpose of This Document
- Records technical decisions and constraints for implementation.
- Keeps Core deterministic and UI thin.
- Anything game-rule related belongs to `System.md` / tuning to `Balance.md`.

---

## Tech Stack
- Language: Go
- UI: REPL-based CUI
- Prompt/Completion: `go-prompt`
- Target OS: Windows / Linux (local terminal)

---

## Architecture Principles

### Core / UI Separation
**Core**
- Owns: state, rules, time progression, quest resolution, event/log generation, save/load
- Must not depend on: terminal I/O, `time.Now()`, `fmt.Println`, OS-specific UI

**UI (REPL)**
- Owns: input parsing, command routing, help text, formatting output, completion/hints
- Must not contain: game rules, random decisions, state mutations outside Core APIs

### Determinism
- Same input + same saved state ⇒ same outputs.
- Any randomness must be deterministic (seeded and stored, or derived reproducibly).
- No wall-clock time in Core logic.

---

## Suggested Module Boundaries
- `core/`  
  - Domain models (Member, Boss, Quest, Equipment, Guild)
  - Systems (time tick, quest tick, guild tick)
  - Event/log generation
- `ui/`  
  - REPL loop, completion, rendering
  - Command parsing & mapping to Core calls
- `cmd/cli/`  
  - Entry point wiring (compose UI + Core, load/save)
- `docs/`  
  - Overview / Dev / System / Balance

(Exact folder names are flexible; the boundaries are not.)

---

## Core API Shape (Conceptual)
- `AdvanceDays(state, days) -> (newState, events)`
- `ApplyCommand(state, command) -> (newState, events)`
- `GetView(state) -> viewModel` (optional; UI-friendly read model)
- `Save(state) / Load()`

(Keep Core APIs UI-agnostic; return structured events, not formatted strings.)

---

## REPL / Command System

### Command List (MVP)
- `status` (overview of clan, quests, gold/fame, roster)
- `hire` / `dismiss`
- `equip` / `unequip` (assign from storage)
- `quests` (list available)
- `accept` (start a quest)
- `retreat` (end quest; no rewards)
- `advance` (time progression)
- `save` / `load` (optional; can be auto-save)

(Command names may change; this is just the initial intent.)

### Command Parsing
- Keep parsing in UI; convert to typed command structs/enums.
- Validate syntax in UI, validate rules in Core.

### Autocomplete / Hints
- Completion is UI-only; uses read-only view data from Core.
- No business logic in completion callbacks.

---

## Events & Logs (Technical)
- Core emits structured events (e.g., `QuestProgressed`, `InjuryWorsened`, `QuestCleared`).
- UI renders events to text lines.
- Keep log templates in Core or a shared `core/text/` layer, but still deterministic.
- Avoid huge narrative text; prefer short, composable templates.

---

## Data & Persistence

### Save/Load Strategy
- Storage: local file
- Format: JSON (human-readable) or gob (compact). JSON recommended early.
- Must serialize: full game state + RNG seed/state (if used)

### Versioning / Migration (Optional)
- Include `save_version` in the save file.
- Provide best-effort migration for minor changes, or fail with a clear message.

---

## Testing Strategy

### Unit Tests (Core)
- Determinism: same seed/state ⇒ same events/state.
- Daily tick ordering invariants.
- Quest resolution invariants (progress bounds, momentum bounds, retreat yields no rewards).

### Simulation / Golden Tests (Optional)
- Snapshot expected event sequences for known seeds.
- Run `advance 30d` and compare outcome to a golden file.

---

## Build Notes
- Single-binary output.
- Keep dependencies minimal.
- Avoid platform-specific assumptions in Core.

---

## Coding Conventions
- Prefer small packages with explicit dependencies.
- Keep domain types in Core stable; tweak numbers in `Balance.md`.
- All tunables should be centralized (e.g., `core/balance/` loads from constants or a config file later).
- Don’t leak UI formatting concerns into Core types.