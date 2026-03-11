# MVP Docs — Shared Rules

## Purpose
This file defines shared rules and conventions for all MVP stage documents under `docs/mvp/`.

## Key principle
`docs/mvp/` is for staging and acceptance criteria. It is not an authoritative source of final rules.
Any long-lived decision made in a stage must be recorded in the owning spec document (`01_Dev.md`, `02_System.md`, `03_Balance.md`, `04_TerminalUI.md`) and referenced from the stage doc.

## Conventions
Each stage doc must include:
- Goal
- Non-goals
- Definition of Done (DoD)
- Required Spec Updates (links to owning documents/sections)
- Test Checklist
- Notes / Open Questions (optional)

## Don’ts (What Not To Write Here)
- Do not duplicate final rules from `02_System.md`. Link to the owning sections instead.
- Do not introduce numeric tuning or balancing content (belongs to `03_Balance.md`).
- Do not define terminal layouts in detail. Reference the output contracts in `04_TerminalUI.md`.
- Do not expand scope beyond the stage’s DoD.
- Do not add long “future design” prose. Keep stage docs actionable.
- Do not turn `docs/mvp/` into a second source of truth.