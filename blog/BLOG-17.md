# Building a Computer from Scratch — Part 17: The Assembler

_In Parts 1–16, we built every physical layer of the machine — wires, gates, storage, a shared bus, arithmetic, comparison, decoding, a 16-bit adder, registers, a stepper, an ALU, 64K of RAM, an I/O bus, a keyboard, a display, a CPU, and a complete computer that can run programs loaded into RAM. This phase adds the final layer: a program that translates human-readable instructions into the binary words the computer executes._

---

## The Problem

Writing a program for the computer means writing 16-bit numbers — one per instruction word — and loading them into RAM starting at address `0x0500`.

An instruction like "add register 0 and register 1" encodes as the number `0x0080`. "Jump to address 0x0600" encodes as two words: `0x0040` followed by `0x0600`. You have to know the opcode table, calculate every address by hand, and count exactly how many words each instruction takes so you can figure out where the next one lands.

Hand-encoding even a ten-instruction program is tedious and fragile. Change one instruction's size, and every jump address that comes after it is wrong.

The solution is a program that does this work for us. You write instructions in a human-readable form — `ADD R0, R1`, `JMP end`, `end:` — and the assembler translates them into the binary the CPU expects. It's the simplest possible bridge between "what the programmer means" and "what the hardware sees."

---

## Context

The book _But How Do It Know?_ introduces assembly mnemonics alongside their binary encodings throughout the instruction chapters, and discusses the assembler's role in the "Hardware and Software," "Programs," and "Languages" chapters. The core insight from those chapters: a program that reads text and emits binary is just software — it manipulates the same kinds of data as any other program running on the machine.

The two-pass structure used here (one pass to measure, one to emit) is not described algorithmically in the book, but it emerges directly from the problem of forward references: a jump instruction near the top of a program needs to know the address of a label near the bottom, and you can't know that address until you've counted the sizes of everything in between.

---

## Building It

### Three Files, One Pipeline

The assembler package has three layers:

1. **Marker types** (`markers.go`) — `LABEL`, `SYMBOL`, and `NUMBER`. These represent the three kinds of operand values an instruction can carry: a named code position, a named constant, or a literal number.

2. **Instruction types** (`instructions.go`) — one struct per instruction mnemonic. Each struct knows its own size in words and how to emit its binary encoding given resolved addresses.

3. **Assembler** (`assembler.go`) — runs two passes over the instruction list, then calls `Emit()` on each instruction with resolvers that look up the addresses and values recorded in pass 1.

There is also a **parser** (`parser.go`) that reads assembly source text line by line and produces the same instruction structs, so you can go from text file to binary in a single pipeline.

### The Instruction Interface

Every instruction type implements:

```go
type Instruction interface {
    Size() int
    Emit(LabelResolver, SymbolResolver) ([]uint16, error)
    String() string
}
```

`Size()` returns how many 16-bit words the instruction occupies — always 1, except `DATA` and `JMP`/`JMPF` (2 words each) and the `CALL` pseudo-instruction (4 words).

`Emit()` takes two resolver functions — one for label addresses, one for symbol values — and returns the actual binary words. Most instructions only use arithmetic on their register numbers:

```go
func (i ADD) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
    return []uint16{uint16(0x0080 + (i.ARegister * 4) + i.BRegister)}, nil
}
```

Instructions that reference code positions delegate to a resolver:

```go
func (i JMP) Emit(lr LabelResolver, _ SymbolResolver) ([]uint16, error) {
    addr, err := lr(i.JumpLoc)
    if err != nil {
        return nil, err
    }
    return []uint16{0x0040, addr}, nil
}
```

The resolvers are closures wired up by the assembler — the instruction itself never knows about the instruction list or the current position.

### Two Passes

Pass 1 walks all instructions once, accumulating a byte-position counter:

```go
position := uint16(0)
for _, ins := range instructions {
    switch v := ins.(type) {
    case DEFLABEL:
        labels[v.Name] = position + codeStartOffset
    case DEFSYMBOL:
        symbols[v.Name] = v.Value
    }
    position += uint16(ins.Size())
}
```

`DEFLABEL` ("end:") and `DEFSYMBOL` ("%SCREEN = 0x0400") both have `Size() = 0`, so they don't advance the position counter — they only record information. When the loop reaches a `DEFLABEL`, `position` already holds the total size of all preceding real instructions, so `position + codeStartOffset` is the exact address where the next real instruction will land.

Pass 2 emits:

```go
position = 0
for i, ins := range instructions {
    // skip labels and symbols — they produce no bytes
    if _, ok := ins.(DEFLABEL); ok { continue }
    if _, ok := ins.(DEFSYMBOL); ok { continue }

    symbols[CURRENTINSTRUCTION] = position + codeStartOffset
    symbols[NEXTINSTRUCTION]    = getNextExecutableInstructionLoc(...)

    words, _ := ins.Emit(resolveLabel, resolveSymbol)
    output = append(output, words...)
    position += uint16(ins.Size())
}
```

Before emitting each instruction, the assembler injects two special symbols: `CURRENTINSTRUCTION` (this instruction's address) and `NEXTINSTRUCTION` (the next instruction's address). These are re-injected fresh for every instruction, which is what makes `CALL` work.

### The CALL Pseudo-Instruction

`CALL` is not a real CPU instruction — the CPU has no call/return mechanism built in. It's a convenience that expands into two real instructions:

```
DATA R3, <return address>
JMP  <routine>
```

Register 3 holds the return address by convention. The return address is the address of the instruction immediately after the `CALL`. That's exactly `NEXTINSTRUCTION` — which the assembler injects into the symbol map before calling `CALL.Emit()`.

```go
func (c CALL) Emit(lr LabelResolver, sr SymbolResolver) ([]uint16, error) {
    next, _ := sr(SYMBOL{NEXTINSTRUCTION})
    dataWords, _ := DATA{REG3, NUMBER{next}}.Emit(lr, sr)
    jmpWords, _ := JMP{c.Routine}.Emit(lr, sr)
    return append(dataWords, jmpWords...), nil
}
```

`CALL` declares `Size() = 4` (two 2-word instructions). The assembler uses this size when computing `NEXTINSTRUCTION` for the instruction that follows `CALL`, ensuring the return address is always correct.

### The Parser

The parser reads source text line by line. Each trimmed, non-empty line matches one of three patterns:

- `myloop:` — a label definition, produces `DEFLABEL{"myloop"}`
- `%SCREEN = 0x0400` — a symbol definition, produces `DEFSYMBOL{"SCREEN", 0x0400}`
- Everything else — an instruction mnemonic with operands

For instructions, the first word is the mnemonic. The parser dispatches on it:

```
ADD, AND, OR, XOR, CMP, LD, ST → two-register instruction
SHR, SHL, NOT, JR              → one-register instruction
DATA                            → register + number/symbol
IN, OUT                         → I/O mode + register
CLF                             → no operands
JMP, CALL                       → label
JMP[CAEZ]+                      → conditional jump with flag subset
```

Conditional jumps like `JMPCAEZ myloop` parse the flag characters directly from the mnemonic name after "JMP", producing `JMPF{[]string{"C","A","E","Z"}, LABEL{"myloop"}}`. The assembler then joins the flags into "CAEZ" and looks up the opcode in a table.

---

## Key Insights

### Two passes exist because of forward references

A single pass can't work: when the assembler encounters `JMP end` at position 0, "end" hasn't been seen yet. Pass 1 reads the whole program and builds a table of where every label lives. Pass 2 can then resolve any reference, forward or backward, without difficulty. The same problem appears in any system where you name something before you define it — the two-pass structure is the minimal fix.

### Labels and symbols are different things by necessity

A label (`end:`) gets its address from where it sits in the instruction stream — it's computed during pass 1 by counting instruction sizes. A symbol (`%SCREEN = 0x0400`) is a programmer-defined constant that doesn't depend on position at all. Conflating them would force the programmer to know the exact load address when they want to write a constant like a display buffer address — which defeats the purpose.

### NEXTINSTRUCTION must be re-injected before every Emit call

`CURRENTINSTRUCTION` and `NEXTINSTRUCTION` are written into the symbol map immediately before each instruction is emitted in pass 2. If they were computed once after pass 1, `CALL` would always see the same stale return address. Re-injecting them per-instruction means the value is always correct for the instruction currently being emitted — a tiny but load-bearing detail.

### SHR, SHL, NOT use stride 5 between registers, not 4

Two-register instructions like `ADD R0, R1` use `0x0080 + (aReg*4) + bReg`. Single-register ALU operations use a stride of 5: `SHR R0 = 0x90`, `SHR R1 = 0x95`, `SHR R2 = 0x9A`, `SHR R3 = 0x9F`. This is because the low nibble in those opcodes encodes both source and destination as the same register (four bits: two for source, two for destination), and the three combinations where source ≠ destination between consecutive register values are simply unused.

---

## Conclusion

Phase 17 delivers the `asm` package: an `Assembler` that resolves labels and symbols in two passes, twenty instruction types that emit their own binary encodings, a `CALL` pseudo-instruction that expands to `DATA R3 + JMP`, and a `Parser` that reads assembly source text and produces instruction structs. Given any assembly program, `Process()` returns a `[]uint16` slice ready to be handed to `SimpleComputer.LoadToRAM()`.

All 28 tests pass. The machine can now be programmed in human-readable assembly.

---

## What's Next

This is the final phase. We now have a complete system: a simulated computer built from first principles — individual wires and NAND gates all the way up — paired with an assembler that lets you write programs for it without counting binary opcodes by hand. The two halves connect through a `[]uint16` slice: the assembler produces it, the computer executes it.
