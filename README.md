# Chronicle of a Clan

A single-player, fully offline clan management game built in Go for terminal-first play (interactive CLI with optional TUI flows).

You are a clan leader. Recruit members, manage equipment, take quests, and manually advance in-game time to resolve outcomes. The clan’s history is recorded through reproducible, system-driven logs.

## Usage

Create a new save slot:
```bash
./coc init <save_dir>
```

Start the game with an existing save slot:
```bash
./coc <save_dir>
```

`<save_dir>` is a save slot name under `saves/` (not a filesystem path).

## Documentation

Design docs live in `docs/`:

- `docs/00_Overview.md` — concept, pillars, goals
- `docs/01_Dev.md` — implementation constraints and architecture notes
- `docs/02_System.md` — game rules and tick order (no tuning values)
- `docs/03_Balance.md` — tuning knobs and numeric tables
- `docs/04_TerminalUI.md` — terminal UI contract (commands, outputs, and presentation rules)
- `docs/mvp/` — staged MVP plans (acceptance criteria only; not the final source of truth)
  - `docs/mvp/Guide.md` — shared rules for writing MVP stage docs

## Status

Design-first phase. Implementation will start after the core loop and balance knobs are stable enough to iterate safely.