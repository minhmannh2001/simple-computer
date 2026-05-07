# Architecture

**Analysis Date:** 2026-05-07

## System Overview

This is a gate-level simulation of a 16-bit computer (the "Scott CPU" / "Simple Computer") implemented in Go. The system models real digital hardware — wires, logic gates, registers, ALU, memory, and I/O peripherals — all as Go structs that propagate boolean signals on every clock tick. A separate assembler toolchain compiles a custom assembly language to binary, which the simulated computer executes.

## Architectural Style

**Hardware simulation / layered bottom-up construction.** The architecture mirrors the physical layers of a real computer:

1. Primitive gates and wires (`circuit/`)
2. Multi-bit components built from gates (`components/`)
3. Subsystems (ALU, memory, I/O) built from components
4. CPU built from subsystems
5. Computer built from CPU + memory + I/O
6. CLI programs that drive the computer

There are two completely separate sub-systems in the repository that have no runtime dependency on each other:
- **Simulator**: runs the hardware simulation with a GUI
- **Assembler**: compiles `.asm` files to `.bin` files loaded into the simulator

## Core Components

| Component | Package | Responsibility | File |
|-----------|---------|----------------|------|
| Wire | `circuit` | Single boolean signal carrier | `circuit/wires.go` |
| Logic Gates | `circuit` | NAND, AND, OR, NOT, XOR, NOR — single-bit boolean operations | `circuit/gates.go` |
| Bit / Word / Register | `components` | 1-bit latch (NAND SR), 16-bit word, 16-bit named register with set/enable | `components/storage.go`, `components/register.go` |
| Bus | `components` | 16-wire shared signal bus | `components/bus.go` |
| IOBus | `components` | 4-wire control bus (clock-set, clock-enable, mode, data-or-address) | `components/iobus.go` |
| ALU | `alu` | 8-operation arithmetic/logic unit (ADD, SHR, SHL, NOT, AND, OR, XOR, CMP) | `alu/alu.go` |
| Stepper | `components` | 6-step clock sequencer controlling CPU micro-operations | `components/stepper.go` |
| Memory64K | `memory` | 65536-cell 16-bit RAM with 8x256 row/column decoders | `memory/memory.go` |
| CPU | `cpu` | Full fetch-decode-execute control unit with instruction decoder, 4 GP registers, IAR, IR, ACC, TMP, flags | `cpu/cpu.go` |
| KeyboardAdapter | `io` | Peripheral: maps key presses to a keycode register accessible via IN instruction | `io/keyboard.go` |
| DisplayAdapter | `io` | Peripheral: writes pixel data to display RAM on OUT instruction; address-selected at I/O address 0x0007 | `io/display.go` |
| ScreenControl | `io` | Reads display RAM at 33ms intervals, renders 160×240 pixel frame to a Go channel | `io/display.go` |
| SimpleComputer | `computer` | Top-level orchestrator: wires CPU + Memory + peripherals together, owns the run loop | `computer/computer.go` |
| Assembler | `asm` | Two-pass assembler: Parser → `[]Instruction` → Assembler emits `[]uint16` binary | `asm/assembler.go`, `asm/parser.go`, `asm/instructions.go` |

## Data Flow

### Simulator execution path

```
clock tick (time.Tick)
    │
    ▼
SimpleComputer.Run()                     [computer/computer.go:92]
    │
    ▼
CPU.Step()                               [cpu/cpu.go:534]
    │  (2 half-clock edges per Step)
    ▼
CPU.step(clockState bool)                [cpu/cpu.go:564]
    │
    ├─► Stepper.Update(clock)            advances step counter 0-5
    │       [components/stepper.go:63]
    │
    ├─► runStep4/5/6Gates()              decode instruction → control signals
    │       [cpu/cpu.go:674-703]
    │
    ├─► runEnable(clock)                 assert enable lines on registers/RAM
    │       [cpu/cpu.go:705]
    │
    ├─► updateStates()                   propagate signals through all components
    │       [cpu/cpu.go:594]            (IAR→MAR→IR→RAM→TMP→FLAGS→BUS1→ALU→ACC→R0-R3→peripherals)
    │
    └─► runSet(clock)                    assert set lines, latch values into registers
            [cpu/cpu.go:791]
```

### Assembly/binary pipeline

```
.asm source file
    │
    ▼
asm.Parser.Parse(reader)                 [asm/parser.go:32]
    │  regex-based line scanner
    ▼
[]Instruction (interface)               [asm/instructions.go:28]
    │
    ▼
asm.Assembler.Process(offset, instrs)   [asm/assembler.go:42]
    │  two-pass: collect labels/symbols, then emit uint16 words
    ▼
[]uint16 (little-endian binary)
    │
    ▼
binary.Write → .bin file                [cmd/assembler/main.go:55]
```

### Display rendering path

```
CPU OUT instruction
    │  (writes to I/O address 0x0007)
    ▼
DisplayAdapter.Update()                  [io/display.go:69]
    │  alternates: write to InputMAR, then write to displayRAM
    ▼
displayRAM (256×256 Cell grid)           [io/display_ram.go]
    │
    ▼  (every 33ms, separate goroutine)
ScreenControl.Update()                   [io/display.go:189]
    │  scans 160 rows × 30 bytes = 240 pixels
    ▼
screenChannel <- &[160][240]byte         [io/display.go:184]
    │
    ▼
GLFW renderer (OpenGL)                   [cmd/simulator/glfw_io.go]
```

### I/O bus control flow

The IOBus carries 4 control signals decoded from IR bits 12-13 (mode/data-or-address). Peripherals poll these signals on every `Update()` call. A peripheral activates only when its address matches what is driven onto the main bus and the IOBus shows the correct set/enable/mode combination.

## Key Design Patterns

**1. Update-propagation model (push-pull simulation)**
Every component exposes an `Update()` method. The CPU calls `updateStates()` in strict topological order on every half-clock edge, manually propagating signals: IAR → MAR → IR → RAM → TMP → FLAGS → BUS1 → ALU → ACC → R0-R3 → Instruction decoder → IO Bus → Peripherals. There is no event system; all propagation is explicit.

**2. Set/Enable duality on registers**
Every register follows the hardware convention: `Set()` latches data from the input bus into the register's Word; `Enable()` drives the register's stored value onto the output bus. These are controlled by separate wires and are explicitly asserted/deasserted each clock phase by the CPU control unit.

**3. Peripheral interface**
```go
// simple-computer/simple-computer/io/peripheral.go
type Peripheral interface {
    Connect(*components.IOBus, *components.Bus)
    Update()
}
```
Peripherals register themselves with the CPU via `CPU.ConnectPeripheral()`. The CPU calls `Update()` on all peripherals at every state update. Peripherals self-select based on IOBus signals and main bus address.

**4. Component interface for bus-connectable elements**
```go
// simple-computer/simple-computer/components/components.go
type Component interface {
    ConnectOutput(Component)
    SetInputWire(int, bool)
    GetOutputWire(int) bool
}
```
Bulk operations (ANDer, ORer, XORer, NOTer, Comparator, Enabler, Adder) all implement this interface, allowing the ALU to wire them uniformly.

**5. Instruction interface (assembler)**
```go
// simple-computer/simple-computer/asm/instructions.go
type Instruction interface {
    String() string
    Emit(LabelResolver, SymbolResolver) ([]uint16, error)
    Size() int
}
```
Each opcode (ADD, LD, ST, JMP, DATA, OUT, IN, etc.) is a concrete struct implementing `Instruction`. The assembler's two-pass design resolves forward label references.

**6. Bit (SR latch) as the foundational memory primitive**
`components.Bit` is a 4-NAND-gate SR latch — the lowest-level stateful element. `Word` contains 16 `Bit`s; `Register` wraps a `Word` and an `Enabler`. All state in the machine (registers, memory cells, flags, carry temp) bottoms out here.

## Architectural Decisions

**Gate-level fidelity over performance.** Every logic operation is modeled as a gate or wire struct. A 16-bit addition is implemented as a ripple-carry `Adder` made of 16 `FullAdder` cells, each made of XOR/AND/OR gates. This makes the simulation accurate but inherently slow (suitable for educational visualization, not performance).

**Separate module roots.** The repository has two Go module roots:
- `/go.mod` declares module `simple-computer` (the top-level, currently empty — only holds this module declaration)
- All actual packages live under `simple-computer/` and are imported as `github.com/djhworld/simple-computer/*` (indicating this is a fork/port of the `djhworld` original).

**Memory map is fixed and hardcoded.**
- `0x0000–0x04FF`: Reserved (ASCII table, pen position, keycode register at fixed addresses)
- `0x0500–0xFEFD`: User code + heap
- `0xFEFE–0xFEFF`: Jump-back stub written by `SimpleComputer.Run()`
- `0xFF00–0xFFFF`: Temporary variables (convention)

**Two-channel concurrency model.** The simulator uses three goroutines: the main GLFW/OpenGL thread (required to be OS-thread-locked), a keyboard goroutine (`keyboard.Run()`), and the computer goroutine (`computer.Run()`). Communication is via typed Go channels: `screenChannel chan *[160][240]byte`, `keyPressChannel chan *io.KeyPress`, and `quitChannel chan bool`.

**Assembler is a self-contained two-pass assembler.** Labels and symbols are collected in pass one (with byte offsets relative to `CODE_REGION_START = 0x0500`). Opcodes are emitted in pass two. The special symbols `CURRENTINSTRUCTION` and `NEXTINSTRUCTION` are reserved and injected per-instruction during emission.

## Boundaries & Interfaces

```
┌─────────────────────────────────────────────────────────────────┐
│                    cmd/simulator/main.go                         │
│  Reads .bin, creates SimpleComputer, starts GLFW loop            │
└────────────────────────────┬────────────────────────────────────┘
                             │ channels: screenChannel, quitChannel
                             │ keyPressChannel
           ┌─────────────────▼──────────────────┐
           │        computer/computer.go          │
           │  SimpleComputer: owns mainBus,       │
           │  wires CPU + Memory + Peripherals    │
           └──┬─────────────┬────────────────────┘
              │             │
    ┌─────────▼──┐    ┌─────▼───────┐
    │ cpu/cpu.go │    │ memory/     │
    │ CPU struct │    │ Memory64K   │
    │ (control   │◄──►│ (256×256    │
    │  unit +    │    │  cells)     │
    │  ALU +     │    └─────────────┘
    │  registers)│
    └──────┬─────┘
           │ ioBus + mainBus
    ┌──────▼──────────────┐
    │  io/ Peripherals    │
    │  KeyboardAdapter    │
    │  DisplayAdapter     │
    │  ScreenControl      │
    └─────────────────────┘

┌──────────────────────────────────────────────────────────┐
│              cmd/assembler/main.go  (independent)         │
│  asm.Parser → []Instruction → asm.Assembler → []uint16   │
└──────────────────────────────────────────────────────────┘
```

**Internal bus topology (CPU):**
- `mainBus` (16-bit): shared bus connecting all GP registers, IAR, Memory, ACC output, IR input
- `controlBus` (16-bit): IR output (instruction register drives this separately to avoid feedback)
- `tmpBus` (16-bit): TMP register output → BusOne input
- `busOneOutput` (16-bit): BusOne output → ALU input B
- `accBus` (16-bit): ALU result output → ACC input
- `aluToFlagsBus` (16-bit): ALU flags output → FLAGS register input
- `flagsBus` (16-bit): FLAGS register output (used by conditional jump logic)
- `ioBus` (4-bit): control signals for I/O peripherals

---

*Architecture analysis: 2026-05-07*
