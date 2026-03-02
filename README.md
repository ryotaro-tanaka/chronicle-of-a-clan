# Chronicle of a Clan

A single-player, offline, idle-style clan management game built as a REPL-based CUI in Go.

This project intentionally avoids cloud services and online requirements. It runs locally on Windows and Linux in the terminal.

---

## 🎯 Project Direction

### Current Goals

- Single-player only
- Offline only
- Runs in terminal (CUI)
- Minimal external dependencies
- Fast iteration on gameplay mechanics
- Clean separation between core logic and UI

---

## 🧠 Game Concept (Early Draft)

The player is a clan leader.

Core loop:

- Recruit clan members
- Send them on quests
- Advance in-game time
- Receive results (gold, loot, stat growth)

Current minimal concept:

- Clan members have stats
- Quests take in-game time (e.g., 3 days)
- The player manually advances time

---

## ⏳ Time System Design (Current Idea)

Instead of real-time waiting:

- A quest may require 3 in-game days
- The player executes a command like:

  ```text
  advance 3d
  ```

- The game processes all time-dependent events
- Quest results are resolved

This makes the game:

- Deterministic
- Offline-friendly
- Testable
- Free from real-time pressure

### Potential Considerations (Not Yet Designed)

- Member salary cost per day
- Food consumption
- Risk of death or injury
- Multiple simultaneous quests
- Idle income vs. active progression
- Infinite time-skipping exploits

Time-system trade-offs will be explored later.

---

## 🏗 Technical Stack

### Language

- Go

Reasoning:

- Cross-platform (Windows/Linux)
- Single-binary distribution
- Fast compilation
- Simple deployment
- No cloud dependency

---

### Interface

- REPL-based CUI
- Terminal interaction
- Tab completion via `go-prompt`

Chosen over TUI because:

- Lighter to implement
- Faster to iterate
- Easier to refactor
- Simpler core/UI separation

---

### Input System

Using:

- `github.com/c-bata/go-prompt`

Features:

- Tab completion
- Command suggestions
- History navigation
- Context-aware hints

Example session:

```text
> status
> hire fighter
> send 2 goblin_cave
> advance 3d
> claim
```

Future ideas:

- Contextual hints
- Smart suggestions
- Command-aware argument completion

---

## 🧩 Architecture Design

Critical decision: strict separation between core and UI.

### Directory Structure (Planned)

```text
/cmd/cli        → REPL entry point
/core           → Game logic, state, rules
/ui             → CLI adapter layer
```

---

### Core Responsibilities

Core must:

- Contain all game rules
- Contain state transitions
- Handle time progression
- Handle quest resolution
- Handle save/load
- Have zero knowledge of UI

Core should expose functions like:

```text
ApplyCommand(state, command) -> (newState, events)
AdvanceTime(state, duration) -> (newState, events)
Save(state)
Load()
```

Core must **not** contain:

- `fmt.Println`
- `os.Stdin`
- direct `time.Now()` calls
- terminal-specific logic

---

### UI Responsibilities

UI must:

- Parse user input
- Call core functions
- Display events and results
- Provide help and hints
- Provide tab completion

UI should remain thin.

---

## 📦 Save System

- JSON-based save file
- Stored locally beside the executable
- No cloud sync
- Deterministic

---

## 🚀 Future Possibilities

Current probability estimate:

- 90%: Remains REPL-only
- 9%: Add GUI for easier distribution
- 1%: Steam release

Design reflects this probability.

---

### If GUI Is Added (9%)

Options:

1. Wails
   - Go core reused
   - Web-based GUI
   - Clean migration path

2. Full TUI (Bubble Tea)
   - More visual terminal interface
   - Still terminal-based

3. Electron / Tauri
   - Wrap logic in desktop app
   - Web UI
   - Heavier, but flexible

---

### If Steam Release Happens (1%)

Likely requirements:

- GUI layer
- Windowed rendering
- Packaging adjustments

Engine decisions may be revisited at that stage. Not a current concern.

---

## 🧪 Development Philosophy

1. Make it playable fast
2. Keep the core pure
3. Avoid premature abstraction
4. Avoid online dependencies
5. Make the time system deterministic
6. Refactor only after gameplay is fun

---

## 🔥 Next Step

Implement a minimal MVP:

- Clan state
- 1 quest
- 1 member
- `advance`
- `status`
- `send`
- `claim`
- `save/load`

Once the loop is satisfying, expand.

---

## 🧭 Summary

This project is:

- A local, offline idle-style management game
- Built in Go
- REPL-based
- Core/UI separated
- Designed for flexibility but optimized for simplicity

Cloud-free. Overengineering-free. Future-proof enough. Fun-first.
