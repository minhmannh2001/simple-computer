# REQUIREMENTS.md — Simple Computer Rebuild

## Core Requirements

### R1 — Phase-gated construction
Each phase must be fully implemented and tested before the next phase begins. No phase skipping. Tests must pass (`go test ./...` exits 0) before the phase is committed.

### R2 — No copying from reference
All implementations are written from understanding of the book and the phase spec. The `simple-computer/` reference folder exists only to read when stuck. No copy-paste.

### R3 — TDD within each phase
Tests are written first (or alongside), verifying the truth table / contract of each component before moving to the next.

### R4 — Blog post per phase
After each phase completes, one blog post is written in `blog/BLOG-NN.md` following the standard structure: Problem → Context → Building It → Key Insights → Conclusion → What's Next.

### R5 — Progress tracking
`STATE.md` is updated after each phase: current phase, completed phases, open notes, challenges encountered.

### R6 — Commit discipline
Each phase ends with a commit: `phase-NN: <phase-name>`. Blog commits are separate: `blog: phase-NN blog post`.

---

## Phase Requirements (Summary)

| Phase | Package | Key Contract |
|-------|---------|-------------|
| 01 | `circuit` | `Wire`, 6 gate types — truth tables verified |
| 02 | `components` (big_gates) | `ANDGate3/4/5/8`, `ORGate3/4/5/6` — N-input gates work correctly |
| 03 | `components` (storage) | `Bit` latches on set=true; holds on set=false; `Word` = 16×Bit |
| 04 | `components` (bus) | `Bus` carries a uint16 bit-by-bit across 16 wires |
| 05 | `components` (bitwise) | `Enabler`, `NOTer`, `ANDer`, `ORer`, `XORer`, `LeftShifter`, `RightShifter`, `IsZero` |
| 06 | `components` (compare) | `Compare2`, `Comparator`, `BusOne` — equal/larger flags; BusOne outputs input OR 1 |
| 07 | `components` (decoders) | `Decoder2x4`, `3x8`, `4x16`, `8x256` — exactly 1-of-2^N high |
| 08 | `components` (adder) | `Add2` full adder; `Adder` 16-bit ripple-carry |
| 09 | `components` (register) | `Register` — set latches bus; enable drives bus |
| 10 | `components` (stepper) | `Stepper` — 6-step sequencer, resets on step 6 |
| 11 | `alu` | `ALU` — 8 ops, 4 flags (carry/larger/equal/zero) |
| 12 | `memory` | `Memory64K` — write addr, read back same value |
| 13 | `io` | `IOBus`, `Peripheral` interface, `KeyboardAdapter` |
| 14 | `io` (display) | `DisplayAdapter`, `displayRAM`, `ScreenControl` |
| 15 | `cpu` | `CPU` — all instructions produce correct register/memory state |
| 16 | `computer` | `SimpleComputer` — program runs; IAR wraps to code start |
| 17 | `asm` | Assembler — text → binary; labels and symbols resolve |

---

## Non-Requirements (explicitly out of scope)

- Performance optimisation (this is a learning simulator, not a fast one)
- GUI or visual debugger
- Publishing blogs to an external platform
- Supporting architectures other than the book's design
