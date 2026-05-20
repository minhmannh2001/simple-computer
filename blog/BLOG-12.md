# Building a Computer from Scratch — Part 12: Memory

_In Parts 1–11, we built wires, gates, storage primitives, a shared bus, bitwise operations, comparison logic, decoders, a 16-bit adder, a register, a stepper, and an ALU. This phase builds the RAM: 65,536 cells arranged in a 256×256 grid, each holding a 16-bit word, accessed through a Memory Address Register and two 8-bit decoders._

---

## The Problem

The CPU needs to hold programs and data. A single register holds one word. Sixteen registers hold sixteen words. But a real program has thousands of instructions and data values. The CPU needs 65,536 words of storage — far too many to wire directly.

If each of the 65,536 cells had a direct wire to the CPU, the CPU would need 65,536 control lines just to select which cell to talk to. That's not a computer; that's a wiring nightmare.

The solution is the same idea used in the decoder chapters: instead of one wire per destination, use a number — an address — and decode that number into a selection. A 16-bit address can name any of 65,536 cells. Two 8-bit decoders (one for the high byte, one for the low byte) split the address into row and column, selecting exactly one cell in a 256×256 grid. The CPU only needs to know the address; the decoders handle the routing.

---

## Context

The book _But How Do It Know?_ covers this across three chapters: "Numbers", "Addresses", and "First Half of the Computer."

"Numbers" establishes binary encoding — each bit is a power of 2. "Addresses" introduces the Memory Address Register (MAR): a register that holds the address, decoupling "where to access" from "what to access." "First Half of the Computer" assembles the complete RAM with MAR, decoders, and a grid of register cells.

The book's RAM is 256 bytes (8-bit address, 8-bit cells). This implementation is 64K words (16-bit address split into two 8-bit halves, 16-bit cells). The principle is identical.

---

## Building It

### The Cell

Each cell wraps a Register and three AND gates:

```go
type Cell struct {
    value components.Register
    gates [3]circuit.ANDGate
}
```

The three gates implement the set/enable control in hardware style:

```go
func (c *Cell) Update(set, enable bool) {
    c.gates[0].Update(true, true)         // constant 1
    c.gates[1].Update(c.gates[0].Output(), set)    // AND(1, set) = set
    c.gates[2].Update(c.gates[0].Output(), enable) // AND(1, enable) = enable

    if c.gates[1].Output() { c.value.Set() } else { c.value.Unset() }
    if c.gates[2].Output() { c.value.Enable() } else { c.value.Disable() }
    c.value.Update()
}
```

`gates[0]` is always on — a wired constant-1. `gates[1]` and `gates[2]` pass the set and enable signals through. The result is that `c.value.Set()` and `c.value.Enable()` are always called via gate output, never via a raw boolean. This mirrors how real hardware gates these control lines rather than branching on them in logic.

### The Memory Structure

```go
type Memory64K struct {
    AddressRegister components.Register    // exported: CPU controls it directly
    rowDecoder      components.Decoder8x256
    colDecoder      components.Decoder8x256
    data            [256][256]Cell
    set             circuit.Wire
    enable          circuit.Wire
    bus             *components.Bus
}
```

256×256 = 65,536 cells. All cells share the same bus for both input and output — safe because only one cell is ever set or enabled per Update call.

### The Update Sequence

```go
func (m *Memory64K) Update() {
    m.AddressRegister.Update()
    m.rowDecoder.Update(
        m.AddressRegister.Bit(0), ..., m.AddressRegister.Bit(7),  // high byte → row
    )
    m.colDecoder.Update(
        m.AddressRegister.Bit(8), ..., m.AddressRegister.Bit(15), // low byte → col
    )
    row := m.rowDecoder.Index()
    col := m.colDecoder.Index()
    m.data[row][col].Update(m.set.Get(), m.enable.Get())
}
```

The MAR is updated first — if set is active, it latches the current bus value (the address). The decoders read the MAR's stored bits (not the raw bus) to identify row and column. Only the selected cell is updated.

---

## Key Insights

### The MAR decouples address from data

The read/write protocol is two-phase:

**Phase 1 (address):** Put the address on the bus, set MAR, call `Update()`. The MAR latches the address. The selected cell does nothing (set=false, enable=false).

**Phase 2 (data):** Put the data on the bus (or clear it for read), signal set or enable on the memory, call `Update()`. The MAR retains its address (set=false means the word doesn't latch). The same cell is selected again. It reads or writes.

Without the MAR, the CPU would have to put address and data on the bus in the same cycle. On a single-bus architecture, that's impossible — the bus can only carry one thing at a time.

### All 65,536 cells share the same bus

Every cell's Register was created with `NewCell(bus, bus)` — the same bus for both input and output. During a write, only the selected cell latches from the bus (set=true). During a read, only the selected cell drives the bus (enable=true). All other cells ignore both signals. Because `Register.Update()` only writes to `outputBus` when `enable.Get()` is true, non-selected cells never affect the bus.

### The Register retains value when set=false

`Register.Update()` always reads the bus into its word's input wires. But `word.Update(set)` only latches those inputs when `set=true`. When `set=false`, the word holds its previous state — the Bit latches don't change. This is what makes the two-phase protocol work: the MAR retains the address through the data phase even though the bus now carries data.

### AddressRegister is exported and CPU-controlled

The MAR is a plain exported field, not hidden behind a method. The CPU loads it exactly as it loads any other register — by putting a value on the bus and calling `AddressRegister.Set()`. There is no special "load address" instruction; the MAR is just another register from the control unit's perspective.

---

## Conclusion

Phase 12 delivers `Memory64K` — 65,536 cells in a 256×256 grid, each a Register-backed word, all sharing the main bus. Two `Decoder8x256` units (one for each byte of the 16-bit address) select the active cell per cycle. The Memory Address Register decouples address loading from data transfer, enabling the two-phase read/write protocol that a single-bus architecture requires.

---

## What's Next

**Phase 13: I/O Bus and Keyboard**

The CPU and memory can compute and store. The next phases connect the computer to the outside world: an I/O bus that lets the CPU read from input devices and write to output devices. Phase 13 adds the keyboard — the first peripheral that can inject values into the main bus.
