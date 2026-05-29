---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: unknown
last_updated: "2026-05-29T06:20:40.638Z"
---

# STATE.md — Project Memory

_Updated after each phase. This is the living record of where we are and what we've learned._

---

## Current State

**Active phase:** 16 — computer
**Last completed:** Phase 15 — cpu
**Overall progress:** 15 / 17 phases done

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

### Phase 02 — multi-input-gates ✅

**Commit:** `phase-02: multi-input gates`
**Package:** `components`
**Delivered:**

- `ANDGate3`, `ANDGate4`, `ANDGate5`, `ANDGate8`
- `ORGate3`, `ORGate4`, `ORGate5`, `ORGate6`
- Full boundary tests (all-false, single-true, all-true, each-position-false for ANDGate8)

**Key insight:** ANDGate8 uses a balanced tree (pairs → pairs-of-pairs → final) rather than a linear chain — reduces gate depth from 7 to 3, closer to real hardware layout.

**Blog:** `blog/BLOG-02.md` ✅

### Phase 15 — cpu ✅

**Commit:** `phase-15: CPU`
**Package:** `cpu`
**Delivered:**

- `CPU` struct — wires all 14 previous components together: 9 registers (R0–R3, TMP, ACC, IR, IAR, FLAGS), Stepper, ALU, BusOne, Memory64K, IOBus; implements full 6-step fetch-decode-execute cycle
- `NewCPU(mainBus, mem)` — constructor initializes all gate arrays, decoders, OR/AND trees, and internal buses
- `Step()` — advances one full clock cycle (two half-steps: rising and falling edge); drives enable then set for each phase
- `SetIAR(addr)` — loads a 16-bit address into IAR for program start
- `ConnectPeripheral(p)` — wires an I/O peripheral to the CPU's IOBus and mainBus

**Key insight discovered:** Our `Stepper` bootstraps step 0 active in `NewStepper()` by calling `step()` during construction. The reference Stepper starts all-false and becomes active on the first `Update`. Net effect: all stepper indices in the CPU need a +1 shift (mod 6) relative to the reference. Mapping: ref[0]→ours[1], ref[1]→ours[2], ..., ref[5]→ours[0].

**Non-obvious:** Import alias `sio "simple-computer/io"` is required because Go's stdlib `io` package collides with our package name at the identifier level.

**Blog:** `blog/BLOG-15.md` ✅

### Phase 14 — display ✅

**Commit:** `phase-14: display`
**Package:** `io`
**Delivered:**

- `displayRAM` — private frame buffer with 4,800 16-bit cells (30 bytes × 160 rows), `InputAddressRegister` for CPU writes, `OutputAddressRegister` for ScreenControl reads, Set/Unset/Enable/Disable/UpdateIncoming/UpdateOutgoing
- `DisplayAdapter` — I/O peripheral at address `0x0007`; detects address via ANDGate8 (NOT gates on wires 8–12, direct on 13–15); latches selection in `displayAdapterActiveBit`; two-phase write protocol enforced by `writeToRAM` toggle latch; `inputMARSetGate` fires on first write (latches address), `displayRAMSetGate` fires on second write (writes cell)
- `ScreenControl` — goroutine that scans all 4,800 cells at ~30fps; extracts 8 pixels per cell from bits 15–8 (high byte); pushes `[160][240]byte` frames to an output channel

**Key insight discovered:** Two separate address registers on `displayRAM` are the hardware solution to the concurrent-access problem. The control logic and the screen scanner both need the frame buffer, but they run at different rates and in different orders. One address register per accessor means neither can corrupt the other's scan pointer — no locking required.

**Blog:** `blog/BLOG-14.md` ✅

### Phase 13 — iobus-and-keyboard ✅

**Commit:** `phase-13: IOBus and keyboard`
**Package:** `components` (IOBus), `io` (Peripheral, KeyboardAdapter, Keyboard)
**Delivered:**

- `IOBus` — 4-wire control bus (CLOCK_SET, CLOCK_ENABLE, MODE, DATA_OR_ADDRESS); Set/Unset/Enable/Disable/Update/query methods; lets the CPU signal "this cycle is I/O, not RAM"
- `Peripheral` interface — `Connect(*IOBus, *Bus)` + `Update()`, the contract every I/O device implements
- `KeyboardAdapter` — detects address `0x000F` via ANDGate8 (NOT gates on wires 0–3, direct on 12–15), latches selection in a `memoryBit` SR latch across the two IOBus phases, drives `mainBus` with the stored keycode when ENABLE+DATA+INPUT fires; `KeyboardInBus` is exported for the Keyboard goroutine
- `Keyboard` — goroutine reading from a `chan *KeyPress` on a 33ms tick; writes `key.Value` to `KeyboardInBus` when `key.IsDown`

**Key insight discovered:** `memoryBit` bridges two mutually exclusive IOBus states. The address-detection gate fires only during SET+ADDRESS+OUTPUT; the keycode-delivery gate fires only during ENABLE+DATA+INPUT. The Bit latch holds "I am selected" across the transition — without it, the keyboard would forget it was addressed the moment the CPU changed IOBus state.

**Non-obvious:** `keycodeRegister` is a value type (`components.Register`) initialized in `Connect()` rather than `NewKeyboardAdapter()`, because `mainBus` isn't available until `Connect()` is called. Assigning the dereferenced `*Register` is safe — pointer fields (word, enabler, inputBus, outputBus) still reference their heap objects.

**Blog:** `blog/BLOG-13.md` ✅

### Phase 12 — memory ✅

**Commit:** `phase-12: 64K RAM`
**Package:** `memory`
**Delivered:**

- `Cell` — wraps `components.Register` with three AND gates for hardware-fidelity set/enable gating; `NewCell(inputBus, outputBus *Bus) *Cell`; `Update(set, enable bool)`
- `Memory64K` — 256×256 grid of Cells sharing a single bus; `AddressRegister components.Register` (exported for CPU access); two `Decoder8x256` units decode high/low bytes of the 16-bit address to row/column; `Set`/`Unset`/`Enable`/`Disable`/`Update`/`String`
- Two-phase protocol: Phase 1 loads address into MAR; Phase 2 reads or writes the selected cell

**Key insight discovered:** The MAR retains its value between Update calls when set=false because the Bit latches only change on set=true. This is what makes the two-phase protocol possible — the bus can carry different things (address, then data) in sequential cycles while the MAR holds the address across both.

**Non-obvious:** All 65,536 cells share the same bus. This is safe because Register.Update() only writes to outputBus when enable=true, and only one cell is ever enabled per Update call. Non-selected cells read the bus silently (their word inputs are updated) but don't latch (set=false) and don't drive (enable=false).

**Blog:** `blog/BLOG-12.md` ✅

### Phase 11 — alu ✅

**Commit:** `phase-11: ALU`
**Package:** `alu`
**Delivered:**

- `ALU` struct — wires inputABus, inputBBus, outputBus, flagsOutputBus together with an Adder, RightShifter, LeftShifter, NOTer, ANDer, ORer, XORer, Comparator, IsZero, Decoder3x8, 7 Enablers, and 3 carry AND gates
- `NewALU(inputABus, inputBBus, outputBus, flagsOutputBus *Bus) *ALU` — constructor wires the four buses
- `Update()` — decodes 3-bit opcode, runs relevant operation, gates result through corresponding Enabler onto outputBus, writes 4 flags (carry, aIsLarger, isEqual, isZero) to flagsOutputBus
- Operation constants: `ADD=0, SHR=1, SHL=2, NOT=3, AND=4, OR=5, XOR=6, CMP=7`

**Key insight discovered:** The Comparator always runs regardless of opcode — not just during CMP. This keeps `aIsLarger` and `isEqual` flags perpetually fresh. If it only ran during CMP, stale comparison flags would persist through subsequent non-CMP operations and require explicit clearing logic in the control unit.

**Non-obvious:** CMP has no enabler and forces `isZero` inputs to all-true (not all-false). If isZero inputs were all-false, the zero flag would fire — falsely claiming the comparison "result" is zero. Forcing all-true ensures IsZero outputs false. The real CMP result lives in the flags bus, not the output bus.

**Blog:** `blog/BLOG-11.md` ✅

### Phase 10 — stepper ✅

**Commit:** `phase-10: stepper`
**Package:** `components`
**Delivered:**

- `Stepper` — 12 `Bit` latches in 6 master-slave pairs forming a shift register
- `NewStepper()` — bootstraps each `Bit`'s `gates[3]` (same as `NewBit()`), then calls `step()` to establish step-0 active state
- `Update(clockIn bool)` — sets clock wire, checks sentinel output[6], calls `step()`; if sentinel fired, immediately calls `step()` again with `reset=true` to reset in the same call
- `GetOutputWire(index int) bool` — returns step output 0–5
- `String()` — "* - - - - -" style display of active step

**Key insight:** Step 0 uses `OR(reset, NOT(slave[0]))` while steps 1–5 use `AND(slave[N], NOT(slave[N+1]))`. This asymmetry is what makes step 0 the "default on" state at power-on (slave[0]=false → NOT(false)=true → OR=true) without any special initialization. The double `step()` on reset ensures the sentinel clears and step 0 is live before `Update` returns — preventing a ghost 7th step.

**Blog:** `blog/BLOG-10.md` ✅

### Phase 09 — register ✅

**Commit:** `phase-09: register`
**Package:** `components`
**Delivered:**

- `Register` — wraps `*Word` + `*Enabler`; constructor calls `word.ConnectOutput(enabler)` so outputs flow automatically
- `Set()` / `Unset()` — assert/deassert the latch control wire
- `Enable()` / `Disable()` — assert/deassert the bus-drive control wire
- `Update()` — copies inputBus → word → enabler → outputBus in sequence; only writes to outputBus when enabled
- `Value() uint16` — reconstructs stored number from word outputs (wire 0 = MSB = 2^15)
- `Bit(index int) bool` — returns individual wire from word output

**Key insight:** `word.ConnectOutput(enabler)` is the critical line in the constructor — it means `word.Update()` automatically pushes word outputs into enabler inputs, eliminating a manual 16-wire copy loop. The chain-of-ownership pattern established in Phase 1 now spans three components in series: inputBus → word → enabler → outputBus.

**Blog:** `blog/BLOG-09.md` ✅

### Phase 08 — adder ✅

**Commit:** `phase-08: full adder and 16-bit ripple-carry adder`
**Package:** `components`
**Delivered:**

- `Add2` — 1-bit full adder; 5 gates (XOR, XOR, AND, AND, OR); correct truth table for all 8 input combinations
- `Adder` — 16 `Add2` stages chained ripple-carry LSB-to-MSB; 32 input wires (0–15 = A, 16–31 = B); `Carry()` exposes MSB overflow
- `Update(carryIn bool)` — walks stages from index 15 down to 0, each stage's carry feeds the next

**Key insight:** The wire iteration walks `awire` from 15 down to 0 (A's LSB to MSB) while `i` also counts 15 → 0 (stage index). Wire 0 is the MSB of A but addition starts at wire 15. The counter decrement must be tracked separately from the stage index — they decrease in sync but represent different things (wire position vs. adder stage index).

**Blog:** `blog/BLOG-08.md` ✅

### Phase 07 — decoders ✅

**Commit:** `phase-07: decoders`
**Package:** `components`
**Delivered:**

- `Decoder2x4` — 2 NOT gates + 4 AND gates; one output per 2-bit combination
- `Decoder3x8` — 3 NOT gates + 8 ANDGate3; `Index()` scans outputs for active wire
- `Decoder4x16` — 4 NOT gates + 16 ANDGate4; `index int` field set during Update
- `Decoder8x256` — 1 selector Decoder4x16 (high nibble a,b,c,d) + 16 sub-Decoder4x16s (low nibble e,f,g,h); only selected sub-decoder runs Update; `Index()` = sel×16 + sub.Index()

**Key insight:** The bit-ordering in `Decoder8x256.Update(a,b,c,d,e,f,g,h)` has `a`=MSB (value 128) and `h`=LSB (value 1). High nibble = (a,b,c,d) selects bank 0–15; low nibble = (e,f,g,h) selects position within bank. Spec description of which nibble is "high" was inverted relative to test expectations — verified from test cases (0x80→128, 0x01→1).

**Blog:** `blog/BLOG-07.md` ✅

### Phase 06 — comparison-and-bus-one ✅

**Commit:** `phase-06: comparison and BusOne`
**Package:** `components`
**Delivered:**

- `Compare2` — single-bit comparator stage; produces `equalOut` and `isLargerOut` by threading flags from higher-significance bits downward; uses `XORGate`, `NOTGate`, `ANDGate` (Phase 01) and `ANDGate3` (Phase 02)
- `Comparator` — chains 16 `Compare2` stages MSB-first; seed: `equalIn=true`, `isLargerIn=false`; exposes `Equal()` and `Larger()` flags
- `BusOne` — reads from `*Bus`, writes to `*Bus`; when disabled passes input through; when enabled outputs constant `0x0001` via `input[i] AND NOT(bus1)` for upper bits and `input[15] OR bus1` for LSB

**Key insight:** The MSB-first chaining order is load-bearing: the OR in `isLargerOut` only accumulates findings — it cannot retract a wrong "larger" set by a low bit later overridden by a more significant bit. Running LSB-first would produce wrong comparisons for any case where MSB differs. BusOne uses the AND/NOT + OR formula to avoid any branching — the same two formulas implement both pass-through and constant-1 modes.

**Blog:** `blog/BLOG-06.md` ✅

### Phase 05 — enabler-and-bitwise ✅

**Commit:** `phase-05: enabler and bitwise operation components`
**Package:** `components`
**Delivered:**

- `Enabler` — 16 AND gates sharing one enable wire; disabled output is always all-zero
- `NOTer` — 16 NOT gates in parallel; inverts every input bit
- `ANDer` — 16 AND gates across two 16-bit operands (inputs 0–15 = A, 16–31 = B)
- `ORer` — 16 OR gates across two 16-bit operands
- `XORer` — 16 XOR gates across two 16-bit operands
- `LeftShifter` — wiring rearrangement: output[i] = input[i+1]; captures shiftOut, fills shiftIn
- `RightShifter` — mirror of LeftShifter; output[i] = input[i-1]
- `IsZero` — ORs all 16 inputs (feeding same value to both A and B sides), then NOTs the result

**Key insight:** The Enabler's guarantee that disabled outputs are all-zero — not floating — is what makes shared bus wiring safe. IsZero reuses the existing ORer by feeding each input to both operand slots simultaneously (`OR(x, x) = x`), collapsing 16 bits into a single "any true?" check without building a new 16-input OR.

**Blog:** `blog/BLOG-05.md` ✅

### Phase 04 — bus ✅

**Commit:** `phase-04: bus`
**Package:** `components`
**Delivered:**

- `Bus` struct — 16-wire shared channel with stable `circuit.Wire` state
- `SetValue(uint16)` — decomposes a 16-bit number MSB-first into wire array
- `String()` — binary string representation, index 0 (MSB) first
- `Component` interface compliance (`ConnectOutput` is a no-op)

**Key insight:** `SetValue` iterates wire indices from high to low while extracting bits low to high — the index and bit position are reversed. Wire 0 = MSB (bit 15), wire 15 = LSB (bit 0). Getting this order wrong produces a mirrored number that passes naive tests but fails `0x0001` vs `0x8000` edge cases.

**Blog:** `blog/BLOG-04.md` ✅

### Phase 03 — storage-primitives ✅

**Commit:** `phase-03: storage primitives (Bit and Word)`
**Package:** `components`
**Delivered:**

- `Bit` — 4-NAND SR latch with double-pass stabilization
- `Word` — 16 `Bit`s in parallel with shared set signal
- `Component` interface — `ConnectOutput`, `SetInputWire`, `GetOutputWire`
- `BUS_WIDTH = 16` constant

**Key insight:** The zero-value struct is an invalid latch state — Go initializes all NAND outputs to false, but real NAND gates can't all be zero simultaneously. `NewBit()` must bootstrap `gates[3]` to true before the latch can hold state correctly. The double-pass `Update` is necessary because the feedback path through gates[3] needs one pass to propagate and a second to fully settle.

**Blog:** `blog/BLOG-03.md` ✅

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
