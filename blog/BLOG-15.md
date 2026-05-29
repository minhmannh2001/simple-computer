# Building a Computer from Scratch — Part 15: The CPU

_In Parts 1–14, we built wires, gates, storage primitives, a shared bus, bitwise operations, comparison logic, decoders, a 16-bit adder, a register, a stepper, an ALU, 64K of RAM, an I/O bus, a keyboard, and a display. This phase wires all of them into a complete CPU: a control unit that fetches instructions, decodes them, and drives everything else._

---

## The Problem

Every previous phase had a clear boundary. The ALU adds. The Register latches. The Stepper counts. The RAM reads and writes. None of them decide anything — they respond to signals that tell them what to do.

The CPU is different. It is the thing that produces those signals.

Each clock cycle, the CPU has to answer: what instruction am I on? What does it do? Which registers need to enable? Which need to set? What should the ALU compute? Does the bus carry an address or data?

The answers change with every instruction, and the CPU has to work them out from a handful of bits — the opcode, the register selectors, the current stepper step, the flag register — and produce a precise combination of enable/set pulses in the right order.

---

## Context

The book _But How Do It Know?_ covers the control unit across two chapters: "The Central Processing Unit" and "Programming." The core idea is a 6-step fetch-decode-execute cycle. Each step is a half-clock: enable on the rising edge, set on the falling edge (or vice versa). The stepper cycles through steps 1–6, and a combinatorial network of AND/OR gates maps each step and instruction bit to a specific set of control signals.

This implementation adds a `cpu` package with a single `CPU` struct that wires together all 14 previous components and runs the complete cycle.

---

## Building It

### The 6-Step Fetch-Decode-Execute Cycle

Every instruction takes exactly 6 steps. Steps 1–3 fetch and decode; steps 4–6 execute.

```
Step 1: IAR → MAR (load instruction address into memory address register)
         BusOne → ACC (stage IAR+1 computation)
Step 2: RAM → IR (load instruction word from memory)
         IAR+1 → ACC
Step 3: ACC → IAR (write IAR+1 back; instruction address now advanced)
Step 4: decode and route operands (varies by instruction)
Step 5: execute (varies by instruction)
Step 6: writeback (for some instructions)
```

The stepper fires each step in sequence. The gate network checks `stepper.GetOutputWire(N)` and ANDs it with instruction bits to produce the final enable/set signals.

### The Stepper Index Offset

The biggest surprise in this phase: our `Stepper` initializes differently from the reference implementation.

In `NewStepper()`, we call `step()` during construction to bootstrap step 0 as active. This means by the time `Step()` is called for the first time during execution, `stepper.GetOutputWire(0)` is already true — the stepper is already at step 0 before the first `Update`.

The reference Stepper starts with all bits false. Its first `Update(true)` triggers a sentinel-reset cycle that makes step 0 active *after* the update. The net effect: what the reference calls "step 0" fires on the first `Update` call, but what we call "step 0" is already active when `step()` is called the first time.

When the CPU calls `stepper.GetOutputWire(0)` during the *first* `step()` half-cycle, our stepper reports step 0 active. But after that first half-cycle advances the stepper, subsequent calls land at step 1, step 2, etc. The reference would land at step 0, step 1, etc.

The fix: shift all stepper indices by +1, wrapping 5 → 0:

```
Reference step[0] → our stepper.GetOutputWire(1)
Reference step[1] → our stepper.GetOutputWire(2)
Reference step[2] → our stepper.GetOutputWire(3)
Reference step[3] → our stepper.GetOutputWire(4)
Reference step[4] → our stepper.GetOutputWire(5)
Reference step[5] → our stepper.GetOutputWire(0)
```

Every `stepper.GetOutputWire(N)` call in the CPU uses the shifted index.

### Instruction Encoding

Instructions are 16-bit words:

```
Bits 15–12: register B (destination or second operand)
Bits 13–12: register A (first operand) — for some instructions
Bit  11–8:  opcode family
Bit  8:     ALU flag (1 = ALU operation)
Bits 11–9:  ALU operation selector (when bit 8 = 1)
```

The `InstructionDecoder3x8` decodes bits 11–9 into one of 8 instruction families (LD, ST, DATA, JMPR, JMP, CLF, ALU-class, IO). An additional NOT gate on bit 8 gates the decoder outputs — only non-ALU instructions pass through.

Two `Decoder2x4` units decode the 2-bit register selectors for "register A" and "register B" into one-hot enables/sets for the four general-purpose registers.

### Enable and Set

Each step fires two sub-steps: enable then set (for a full clock cycle).

```go
func (c *CPU) step(clockState bool) {
    c.stepper.Update(clockState)
    c.runStep4Gates()
    c.runStep5Gates()
    c.runStep6Gates()

    c.runEnable(clockState)
    c.updateStates()
    if clockState {
        c.runEnable(false)
        c.updateStates()
    }

    c.runSet(clockState)
    c.updateStates()
    if clockState {
        c.runSet(false)
        c.updateStates()
    }

    c.clearMainBus()
}
```

The gate precomputation (`runStep4Gates`, `runStep5Gates`, `runStep6Gates`) evaluates AND combinations of stepper outputs and instruction bits before the enable/set phases consume them. Enable drives data onto the bus; set latches data from the bus. The `if clockState` double-pass ensures that the rising edge both asserts and de-asserts within the same call.

### Wiring All Previous Components

The CPU struct holds every previous component by value or pointer:

```go
type CPU struct {
    gpReg0, gpReg1, gpReg2, gpReg3 components.Register
    tmp, acc, ir, iar, flags        components.Register
    memory    *memory.Memory64K
    alu       *alu.ALU
    stepper   *components.Stepper
    busOne    components.BusOne
    ioBus     *components.IOBus

    mainBus, tmpBus, busOneOutput, controlBus, accBus  *components.Bus
    aluToFlagsBus, flagsBus                            *components.Bus

    // ... gate arrays for step4, step5, step6 logic ...
    // ... decoders for instruction and register selection ...
    // ... OR/AND gate trees for each control signal ...
}
```

`mainBus` is the spine: IAR, RAM, registers, and ACC all connect to it. The `tmp` and `acc` registers have their own buses (`tmpBus`, `accBus`) because they need to carry values through the ALU without conflicting with the main bus.

The `flags` register reads from `aluToFlagsBus` (ALU output) and drives `flagsBus` (which the conditional jump gates read from). It is always enabled and always set during the ALU step — the flags update automatically every ADD/CMP/NOT/shift instruction.

### The `sio` Package Alias

The `io` package in Go's standard library collides with our `simple-computer/io` package name. Both are just `io`. The fix: import our package under an alias:

```go
import sio "simple-computer/io"
```

`sio` for "simple-computer io." The alias avoids the shadowing without renaming anything in the package itself.

---

## Key Insights

### The stepper offset is a construction-time decision, not a runtime patch

The Stepper's bootstrap call in `NewStepper()` sets step 0 active before the CPU ever runs. Every stepper index in the CPU has to account for this — it's not a quirk to work around per-instruction, it's a global +1 shift applied uniformly. Once the mapping is written down, it's stable: the same offset holds for every step gate in every instruction.

### The double-pass enable/set is not redundant

The `if clockState { runEnable(false); updateStates() }` after the first `runEnable(clockState)` looks like extra work, but it's structural. Enable signals put values on the bus; set signals latch those values. If the enable stays asserted through the set phase, the component being set might see bus values from two enables simultaneously. The double-pass asserts enable, propagates state, then de-asserts enable before proceeding to set. Bus values are clean for each phase.

### The flag register persists between cycles on purpose

`flags` is always enabled and always set. It reads the ALU output flags after every operation, not just CMP. This means the conditional jump gates (`JMPZ`, `JMPC`, etc.) always see the flags from the *most recent ALU instruction*, whatever that was. CLF (`0x0060`) clears the flags explicitly by running a zero-producing ALU operation and latching the result.

### Sharing the bus is safe because only one driver is active at a time

All four general-purpose registers share `mainBus`. The `Enabler` in each register outputs all-zero when disabled — not floating. The OR gate trees for register enable ensure that only one register (or one other driver: IAR, RAM, ACC) asserts the bus in any given half-step. Because everything is deterministic and single-threaded, there's no race condition.

---

## Conclusion

Phase 15 delivers the CPU: a `cpu` package with a `CPU` struct that wires together all 14 previous components and implements the complete 6-step fetch-decode-execute cycle. The test suite covers all instruction families — LD, ST, DATA, JMPR, JMP, all 15 conditional jump variants, CLF, all ALU operations (ADD with carry, SHL, SHR, NOT, AND, OR, XOR, CMP), IAR auto-increment, IO input/output, and multi-instruction programs (subtract, multiply). All 46 tests pass.

The critical discovery this phase: our Stepper bootstraps step 0 active in the constructor, shifting all stepper index references by +1 (mod 6) relative to the reference implementation.

---

## What's Next

**Phase 16: The Complete Computer**

The CPU is now a standalone unit. Phase 16 assembles it with the clock, the memory, the I/O bus, the keyboard, and the display into a single runnable `Computer` struct — and wires up the goroutines that let keyboard input and screen output happen in real time.
