# Terminal UI

## Purpose
Defines the terminal UI contract: inputs, outputs, and presentation rules. This document does not define game rules.

## Scope (UI contract only)
- No game rules or order of operations (see `02_System.md`).
- No tuning numbers (see `03_Balance.md`).

## Startup (CLI) contract
- The program is invoked with a save slot name (not a filesystem path).
- `coc <save_dir>` loads from: `saves/<save_dir>/` and enters the interactive Bubble Tea TUI.
- `coc init <save_dir>` creates: `saves/<save_dir>/` from the template and exits (or prints success and exits).

### No-arg usage messages
- `coc` (no args) prints:
  - `Usage: coc <save_dir>`
  - `Usage: coc init <save_dir>`

- `coc init` (no args) prints at least:
  - `Usage: coc init <save_dir>`
  - (Optional) list existing slots under `saves/` if easy

### Slot name rules (user-visible)
- `<save_dir>` accepts only `A-Za-z0-9._-` and must not start with `-`.
- `<save_dir>` is a slot name; paths are not accepted.

## Interactive session model
- Interactive input is handled inside Bubble Tea.
- The primary screen is a navigation view with a one-line command input, output log, and current virtual path.
- Focused task screens may replace the navigation view temporarily (for example: party setup and equipment selection) and then return to navigation.
- The UI must not contain game rules.

## Navigation model (Virtual FS)

### Commands
- `ls` / `ls <path>`
  - `ls` lists entries at the current location.
  - `ls <path>` lists entries at `<path>` without changing the current location.
  - entries include type tags:
    - `[DIR]` navigable category
    - `[VIEW]` read-only view
    - `[ACT]` state-changing action
- `cd <path>`
  - navigation only
  - supports `.` and `..`

### Path conventions (virtual FS)
- Accept both `clan` and `clan/` as directory paths.
- Normalize simple segments: `.` and `..`.
- Invocation by path is allowed (examples):
  - `clan/status`
  - `../exit`
  - `quests/keys/hunt_ambushjaw_gator/info`

### Quests (virtual FS)
- Root `ls` includes `[DIR] quests`.
- `quests/` children: `[DIR] keys`, `[DIR] forest`, `[DIR] volcano`, `[DIR] desert`, `[DIR] tundra`, `[DIR] swamp`.
- **`quests/keys/`**: Lists only the key quest whose `order` equals `key_quest_progress.current_order` (the single “next” quest to advance the story). One quest per order, so `keys/` shows one entry. Each entry: `[DIR] hunt_<monster_slug>`.
- **`quests/<region>/`**: Lists all key quests with `order <= current_order` in that region (filter by the profile’s region). Each entry: `[DIR] hunt_<monster_slug>` (e.g. `hunt_ambushjaw_gator`).
- Under each `hunt_<slug>`: `[VIEW] info`, `[ACT] party`, `[ACT] clear`.
- Invocation by path (e.g. `quests/keys/hunt_ambushjaw_gator/info`) is supported.
- **Quest info view contract**: Output lines: `Name: <name>`, `Lv: <min>-<max>`, `Specialties: <stats>`, `Reward:`, `Description: <description>`. Reward is empty for MVP.

## Completion and line input
- Completion includes:
  - command names: `ls`, `cd`, `status`, `exit`
  - directory paths after `cd ` and `ls `
  - path invocation forms like `clan/status`
- Completion candidates come from the virtual FS tree and current location.
- The navigation input supports basic line editing, space entry, and `Tab` completion.
- Command history is optional; MVP does not require shell-like history behavior.

## Commands (Stage 1–3 baseline)
- `ls` / `ls <path>`
- `cd <path>`
- `status` (view)
- `exit` (action)

No `help` command.

## Quest party actions (Stage 9)
- `party`
  - Available under a quest directory as `[ACT] party`.
  - Opens a dedicated member-selection screen instead of printing a text-only response.
  - Restores any previously selected members and equipment for that quest from session memory.
- `clear`
  - Available under a quest directory as `[ACT] clear`.
  - Clears the stored party selection for that quest only.
  - Does not clear the navigation log.

## Party Setup Screen Contract
- Title: `Party Setup - Select Members`
- Shows the save roster as a vertical list with cursor and selected markers.
- Supports:
  - `Up` / `Down`: move cursor
  - `Enter`: toggle current member
  - `F`: confirm selected members and move to equipment selection
  - `Esc`: cancel and return to navigation without committing changes
- The screen enforces a maximum of 4 selected members.

## Equipment Selection Screen Contract
- Title format: `Equip Member - <member_name>`
- Shows weapon candidates, armor candidates, current stats, and a preview block.
- Supports:
  - `Up` / `Down`: move within the focused equipment list
  - `Tab`: switch focus between weapon and armor
  - `Enter`: confirm the highlighted option for the focused list
  - `N`: advance to the next selected member
  - `Esc`: return to the member-selection screen
- After the last selected member is confirmed, control returns to the navigation screen and the per-quest party state remains in memory for the session.

## Dev commands (Stage 5+)

> These commands are intended for development and testing only. They are not part of the player-facing command set.

- `dev/create_boss <profile_id> [seed]`
  - **Purpose**: Generate a single boss for the given profile and print a compact summary for inspection.
  - **Arguments**:
    - `profile_id` (required): boss profile identifier, e.g. `forest_001`.
    - `seed` (optional): integer seed for reproducible generation. If omitted, an internal time-based seed is used.
  - **Behaviour**:
    - Calls Core with the given `profile_id` and `seed` (if provided).
    - Prints a summary including: profile_id, name, description, region, level, and stats (Power, Guard, Evasion, Cunning).
  - **Example output** (format is illustrative and may be adjusted):
    - `Boss: profile_id=forest_003 name="Ambushjaw Gator" description="..." region=forest level=3`
    - `Stats: Power=120 Guard=95 Evasion=110 Cunning=130`

## Status Output Contract (compact)
Required lines:
- `Clan: <clan_name>   Day: <day>`
- `Gold: <gold>   Fame: <fame>`
- `Members: <member_count>   ActiveQuests: <active_quest_count>`

## Error handling (minimum)
Errors must be actionable and include the reason.

### CLI-level errors
- Missing args:
  - print usage (including `init`)
- Invalid slot name:
  - explain allowed characters and the leading `-` rule
- Slot not found:
  - `saves/<save_dir>/` does not exist
- Init target exists:
  - `saves/<save_dir>/` already exists

### Save file errors
- Missing required file: `clan.json` not found
- Invalid JSON: parse error details (file name + reason)
- Unsupported version: show detected version and supported versions

## Exit behavior
- `exit` ends the interactive session.
- Ctrl+C handling is optional (may be added later).
