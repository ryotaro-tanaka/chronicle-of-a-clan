# Chronicle of a Clan

A single-player, fully offline clan management game built in Go for terminal-first play (interactive CLI with optional TUI flows).

You are a clan leader. Recruit members, manage equipment, take quests, and manually advance in-game time to resolve outcomes. The clan’s history is recorded as deterministic, system-driven logs.

## Documentation

Design docs live in `docs/`:

- `docs/00_Overview.md` — concept, pillars, goals
- `docs/01_Dev.md` — implementation constraints and architecture notes
- `docs/02_System.md` — game rules and tick order (no tuning values)
- `docs/03_Balance.md` — tuning knobs and numeric tables
- `docs/04_TerminalUI.md` — terminal UI contract (commands, outputs, and presentation rules)
- `docs/mvp/` — MVP-specific notes and scope constraints

## Status

Design-first phase. Implementation will start after the core loop and balance knobs are stable enough to iterate safely.