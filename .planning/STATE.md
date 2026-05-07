# STATE.md — Project Memory

_Updated after each phase. This is the living record of where we are and what we've learned._

---

## Current State

**Active phase:** 02 — multi-input-gates
**Last completed:** Phase 01 — wire-and-gates
**Overall progress:** 1 / 17 phases done

---

## Completed Phases

### Phase 01 — wire-and-gates ✅
**Commit:** `phase-01: wire and logic gates`
**Package:** `circuit`
**Delivered:**
- `Wire` struct — named, mutable single-bit state holder
- `NANDGate`, `ANDGate`, `NOTGate`, `ORGate`, `XORGate`, `NORGate`
- Full truth-table tests for all gate types

**Key insight discovered:** The "chain-of-ownership" pattern — every component stores its own output in an internal `Wire` and exposes it via `Output()`. This is what makes feedback loops stable: `Output()` returns a committed snapshot, not a live formula. Without this, composing gates into SR latches would cause infinite re-evaluation.

**Surprise:** XOR is implemented as `!((!a && !b) || (a && b))` — the NAND-derivable form that asks "are the inputs the same?" and inverts — rather than the more intuitive `(a || b) && !(a && b)`.

**Blog:** `blog/BLOG-01.md` ✅

---

## Open Notes

_Things to remember when starting the next phase or when returning after a break._

- Phase 02 (`multi-input-gates`) belongs in a new `components` package — separate from `circuit`
- The book calls these "More Gate Combinations" — they're just chained AND/OR gates, not new primitives
- The go module structure needs to be decided: one module for the whole project, or a module per package? (Reference had two diverging modules — avoid this)

---

## Challenges Log

| Phase | Challenge | Resolution |
|-------|-----------|-----------|
| 01 | — | Smooth start, no blockers |

---

## Architecture Decisions Made

| Decision | Rationale |
|----------|-----------|
| `Wire` as named mutable holder, not bare `bool` | Enables chain-of-ownership pattern; necessary for stable feedback loops |
| Gates store output in internal `Wire` | `Output()` returns committed snapshot — composable and feedback-safe |
| XOR via NAND-derivable form | Closer to real hardware; same truth table |
| 16-bit words instead of book's 8-bit bytes | Matches reference Go simulator; all book concepts apply unchanged |
