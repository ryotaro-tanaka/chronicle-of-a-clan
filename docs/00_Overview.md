# Chronicle of a Clan — Overview

## What this game is
Chronicle of a Clan is a single-player, fully offline clan management game built as a REPL-based CUI in Go.  
You lead a clan, send members on quests, advance time manually, and watch the clan’s history accumulate through system-driven logs.

## Core pillars
- Clan management is the main game (recruit, equip, assign, sustain).
- Time is advanced manually (`advance`) and drives all progression.
- Growth is the primary fun (members get stronger, gear improves, bigger wins).
- The story emerges from systems and logs (no heavy hand-written narrative).
- Lightweight and deterministic (strict Core/UI separation).

## MVP scope (current focus)
- Boss hunt quests as the first complete gameplay loop.
- Party-based planning (1–4 members per quest).
- A simple out-of-quest layer to prevent “infinite hiring” from becoming optimal.

## Goals
### Short-term
Improve members and gear, learn what works, clear tougher hunts.

### Mid-term
Expand influence through progression systems (regions / guild rank).

### Long-term (win condition)
Reach **#1 on the Fame ranking**.

## How to read this spec set
- **01_Dev.md**: tech stack, architecture rules, repo structure, REPL notes.
- **02_System.md**: the game logic (quests, stats responsibilities, time ticks, guild management). What happens and in what order.
- **03_Balance.md**: tuning knobs only (thresholds, tables, reward ranges, upkeep curve). How much it happens and where the thresholds are.
