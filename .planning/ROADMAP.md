# ROADMAP.md — Simple Computer Rebuild

A 17-phase journey from a single wire to a working assembled computer.

Each phase maps to one or more chapters from _But How Do It Know?_ by J. Clark Scott. Phases must be completed in order — each one depends on the layer below it.

---

## Phases

| # | Name | Package | Status | Blog |
|---|------|---------|--------|------|
| 01 | wire-and-gates | `circuit` | ✅ Done | ✅ blog/BLOG-01.md |
| 02 | multi-input-gates | `components` | ✅ Done | ✅ blog/BLOG-02.md |
| 03 | storage-primitives | `components` | ✅ Done | ✅ blog/BLOG-03.md |
| 04 | bus | `components` | ✅ Done | ✅ blog/BLOG-04.md |
| 05 | enabler-and-bitwise | `components` | ✅ Done | ✅ blog/BLOG-05.md |
| 06 | comparison-and-bus-one | `components` | ✅ Done | ✅ blog/BLOG-06.md |
| 07 | decoders | `components` | ✅ Done | ✅ blog/BLOG-07.md |
| 08 | adder | `components` | ✅ Done | ✅ blog/BLOG-08.md |
| 09 | register | `components` | ✅ Done | ✅ blog/BLOG-09.md |
| 10 | stepper | `components` | ✅ Done | ✅ blog/BLOG-10.md |
| 11 | alu | `alu` | ✅ Done | ✅ blog/BLOG-11.md |
| 12 | memory | `memory` | ✅ Done | ✅ blog/BLOG-12.md |
| 13 | iobus-and-peripheral | `io` | ⬜ Not started | ⬜ |
| 14 | display | `io` | ⬜ Not started | ⬜ |
| 15 | cpu | `cpu` | ⬜ Not started | ⬜ |
| 16 | computer | `computer` | ⬜ Not started | ⬜ |
| 17 | assembler | `asm` | ⬜ Not started | ⬜ |

---

## Phase Detail

### Phase 01 — wire-and-gates ✅
**Book chapters:** "Just a Little Bit", "What The...?", "Simple Variations", "Diagrams"
**Delivers:** `Wire`, `NANDGate`, `ANDGate`, `NOTGate`, `ORGate`, `XORGate`, `NORGate`
**Proves:** Truth tables for all 6 gate types
**Commit:** `phase-01: wire and logic gates`

### Phase 02 — multi-input-gates
**Book chapters:** "More Gate Combinations" (first part)
**Delivers:** `ANDGate3`, `ANDGate4`, `ANDGate5`, `ANDGate8`, `ORGate3`, `ORGate4`, `ORGate5`, `ORGate6`
**Proves:** Chained AND/OR gates produce correct single output with N inputs
**Package:** `components`

### Phase 03 — storage-primitives
**Book chapters:** "Remember When", "What Can We Do With a Bit?", "A Rose By Any Other Name", "Eight Is Enough"
**Delivers:** `Bit` (NAND SR latch), `Word` (16×Bit)
**Proves:** Bit latches on set=true; holds value when set=false

### Phase 04 — bus
**Book chapters:** "The Magic Bus"
**Delivers:** `Bus` (16 wires, SetValue/GetOutputWire)
**Proves:** Bus correctly carries a uint16 value bit-by-bit

### Phase 05 — enabler-and-bitwise
**Book chapters:** "Back to the Byte" (Enabler section), "More Gates", "Messing With Bytes", various shifter/logic chapters
**Delivers:** `Enabler`, `NOTer`, `ANDer`, `ORer`, `XORer`, `LeftShifter`, `RightShifter`, `IsZero`
**Proves:** Each 16-bit operation produces the correct output word

### Phase 06 — comparison-and-bus-one
**Book chapters:** "The Comparator and Zero", "More of the Processor" (Bus 1)
**Delivers:** `Compare2`, `Comparator`, `BusOne`
**Proves:** Equal/larger flags correct; BusOne outputs input OR constant 1

### Phase 07 — decoders
**Book chapters:** "More Gate Combinations" (second part: 2×4 Decoder)
**Delivers:** `Decoder2x4`, `Decoder3x8`, `Decoder4x16`, `Decoder8x256`
**Proves:** Given N input bits, exactly 1 of 2^N outputs is high

### Phase 08 — adder
**Book chapters:** "The Adder"
**Delivers:** `Add2` (full adder), `Adder` (16-bit ripple-carry)
**Proves:** 1+1=2, max+1=overflow with carry, arbitrary additions correct

### Phase 09 — register
**Book chapters:** "Back to the Byte", "The Magic Bus"
**Delivers:** `Register` (Word + Enabler + bus interface)
**Proves:** Set latches bus value; Enable drives it back onto bus

### Phase 10 — stepper
**Book chapters:** "The Clock", "Step by Step", "Doing Something Useful"
**Delivers:** `Stepper` (6-step clock sequencer)
**Proves:** Steps 0→1→…→5→0; reset fires instantly on step 6

### Phase 11 — alu
**Book chapters:** "Logic", "The Arithmetic and Logic Unit"
**Delivers:** `ALU` (8 ops: ADD/SHR/SHL/NOT/AND/OR/XOR/CMP; 4 flags)
**Proves:** All 8 operations produce correct output and flags

### Phase 12 — memory
**Book chapters:** "Numbers", "Addresses", "First Half of the Computer"
**Delivers:** `Cell`, `Memory64K` (MAR + 2×Decoder8x256 + 256×256 cells)
**Proves:** Write to address, read back same value; different addresses are independent

### Phase 13 — iobus-and-peripheral
**Book chapters:** "The Outside World", "The Keyboard"
**Delivers:** `IOBus`, `Peripheral` interface, `KeyboardAdapter`
**Proves:** IN Data from keyboard delivers keycode to main bus

### Phase 14 — display
**Book chapters:** "The Display Screen"
**Delivers:** `DisplayAdapter`, `displayRAM`, `ScreenControl`
**Proves:** OUT Addr then OUT Data writes pixel to correct frame buffer cell

### Phase 15 — cpu
**Book chapters:** "The Other Half of the Computer" through "Ta Daa!"
**Delivers:** `CPU` (9 registers, stepper, ALU, full control unit)
**Proves:** All instructions produce correct register/memory state

### Phase 16 — computer
**Book chapters:** "Ta Daa!" (complete), "A Few More Words on Arithmetic"
**Delivers:** `SimpleComputer` (CPU + RAM + keyboard + display)
**Proves:** Program loaded at 0x0500 runs; IAR wraps to code start at end

### Phase 17 — assembler
**Book chapters:** Instruction mnemonic tables, "Hardware and Software", "Programs", "Languages"
**Delivers:** `Instruction` types, `Assembler` (two-pass), `Parser`
**Proves:** Assembly source → correct binary; labels and symbols resolve
