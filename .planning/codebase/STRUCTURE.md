# Code Structure

**Analysis Date:** 2026-05-07

## Directory Layout

```
simple-computer/                        ← repo root
├── go.mod                              # Go module: "simple-computer" (top-level, no code)
├── .gitignore
├── .planning/codebase/                 # GSD planning docs (this file lives here)
├── circuit/                            # STANDALONE: primitive gates/wires (early-phase impl)
│   ├── wires.go
│   ├── wires_test.go
│   ├── gates.go
│   └── gates_test.go
├── simple-computer/                    # PRIMARY: all production code lives here
│   ├── circuit/                        # Primitive: Wire, NAND/AND/OR/NOT/XOR/NOR gates
│   │   ├── wires.go
│   │   ├── wires_test.go
│   │   ├── gates.go
│   │   └── gates_test.go
│   ├── components/                     # Layer 2: multi-bit components built from gates
│   │   ├── bus.go                      # Bus (N-wire shared signal line)
│   │   ├── iobus.go                    # IOBus (4-wire I/O control bus)
│   │   ├── storage.go                  # Bit (SR latch), Word (16×Bit)
│   │   ├── register.go                 # Register (Word + Enabler + set/enable wires)
│   │   ├── stepper.go                  # Stepper (6-step clock sequencer)
│   │   ├── adder.go                    # HalfAdder, FullAdder, Adder (ripple-carry)
│   │   ├── big_gates.go                # ORGate3..6, ANDGate3..8, ORGate3..5 (fan-in variants)
│   │   ├── decoders.go                 # Decoder2x4, Decoder3x8, Decoder8x256
│   │   ├── components.go               # Component interface; Enabler, Shifters, IsZero,
│   │   │                               # NOTer, ANDer, ORer, XORer, Comparator, BusOne
│   │   └── *_test.go
│   ├── alu/
│   │   ├── alu.go                      # ALU: 8-operation unit (ADD/SHR/SHL/NOT/AND/OR/XOR/CMP)
│   │   └── alu_test.go
│   ├── memory/
│   │   ├── memory.go                   # Memory64K (256×256 Cell grid + MAR + row/col decoders)
│   │   └── memory_test.go
│   ├── cpu/
│   │   ├── cpu.go                      # CPU: full control unit, 9 registers, all gate arrays
│   │   └── cpu_test.go
│   ├── io/
│   │   ├── peripheral.go               # Peripheral interface (Connect, Update)
│   │   ├── keyboard.go                 # KeyboardAdapter + Keyboard goroutine
│   │   ├── display.go                  # DisplayAdapter + ScreenControl
│   │   └── display_ram.go              # displayRAM (separate RAM for screen pixels)
│   ├── computer/
│   │   └── computer.go                 # SimpleComputer: top-level orchestrator + run loop
│   ├── asm/
│   │   ├── parser.go                   # regex-based line parser → []Instruction
│   │   ├── assembler.go                # two-pass assembler → []uint16
│   │   ├── instructions.go             # all Instruction structs (DATA, JMP, ADD, LD, ...)
│   │   ├── markers.go                  # LABEL, SYMBOL, NUMBER, DEFLABEL, DEFSYMBOL types
│   │   └── *_test.go
│   ├── utils/
│   │   └── common.go                   # ValueToString(uint16) → "0x?????" formatting helper
│   ├── cmd/
│   │   ├── simulator/
│   │   │   ├── main.go                 # Entry: reads .bin, starts GLFW + computer goroutines
│   │   │   └── glfw_io.go              # OpenGL/GLFW window + keyboard event loop
│   │   ├── assembler/
│   │   │   ├── main.go                 # Entry: reads .asm, writes .bin (or string dump)
│   │   │   └── README.md
│   │   └── generator/
│   │       ├── main.go                 # Entry: generates .asm programs using asm package API
│   │       └── common.go               # Shared subroutines (renderString, updatePenPosition, ...)
│   └── _programs/
│       ├── Makefile                    # Build targets: assemble all .asm → .bin
│       ├── README.md
│       ├── ascii.asm / ascii.bin       # ASCII table display program
│       ├── brush.asm / brush.bin       # Drawing program
│       ├── text-writer.asm / .bin      # Text rendering program
│       ├── me.asm / me.bin             # Author bio display program
│       └── screenshots/
├── blog/                               # Blog-related content
├── PHASE-01-*.md ... PHASE-17-*.md     # Implementation plan documents (one per feature phase)
├── PHASES.md                           # Master list of all 17 implementation phases
└── prompt.md                           # Original project prompt/specification
```

## Module Organization

Code is grouped **by hardware layer / functional subsystem**, mirroring real computer architecture:

| Layer | Package path | What it contains |
|-------|-------------|-----------------|
| Primitives | `simple-computer/circuit` | Wire, 2-input gates |
| Components | `simple-computer/components` | Multi-bit combinational + storage components |
| ALU | `simple-computer/alu` | Arithmetic/logic unit |
| Memory | `simple-computer/memory` | 64K word RAM |
| CPU | `simple-computer/cpu` | Control unit wiring all components |
| I/O | `simple-computer/io` | Keyboard and display peripherals |
| Computer | `simple-computer/computer` | Final assembly + run loop |
| Assembler | `simple-computer/asm` | Source language → binary |
| CLI | `simple-computer/cmd/*` | Runnable programs |
| Utilities | `simple-computer/utils` | Shared formatting helpers |

**Dependency direction** (strict — lower layers never import upper layers):

```
circuit ← components ← alu
                     ← memory
                     ← cpu (also imports alu, memory, io, circuit)
                     ← io (also imports circuit)
                     ← computer (also imports cpu, memory, io)
asm ← cmd/assembler
asm ← cmd/generator
computer ← cmd/simulator
utils ← components, alu, asm, cpu
```

**Duplicate directory:** `circuit/` exists at both the repo root and inside `simple-computer/`. The root-level `circuit/` appears to be an earlier standalone experiment. All production code imports from `simple-computer/circuit` via the `github.com/djhworld/simple-computer/circuit` path.

## Key Files

| File | Role |
|------|------|
| `simple-computer/circuit/wires.go` | `Wire` struct — the atomic boolean signal carrier |
| `simple-computer/circuit/gates.go` | All 2-input gate types (NAND, AND, OR, NOT, XOR, NOR) |
| `simple-computer/components/storage.go` | `Bit` (SR latch), `Word` (16-bit) — foundational stateful elements |
| `simple-computer/components/components.go` | `Component` interface + all 16-bit bulk operators (Enabler, ANDer, ORer, XORer, NOTer, Comparator, BusOne, Shifters, IsZero) |
| `simple-computer/components/register.go` | `Register` — 16-bit Word + Enabler with named set/enable control wires |
| `simple-computer/components/stepper.go` | `Stepper` — 6-step sequencer driving CPU micro-op timing |
| `simple-computer/components/decoders.go` | `Decoder2x4`, `Decoder3x8`, `Decoder8x256` — address and instruction decoding |
| `simple-computer/alu/alu.go` | `ALU` — routes inputs through 7 operation units, selects one via `Decoder3x8` |
| `simple-computer/memory/memory.go` | `Memory64K` — 256×256 `Cell` array, indexed by 8-bit row/col decoders |
| `simple-computer/cpu/cpu.go` | `CPU` — the largest file (960 lines); full control unit with all gate arrays, instruction decoding, enable/set logic for every component |
| `simple-computer/io/peripheral.go` | `Peripheral` interface — two methods: `Connect` and `Update` |
| `simple-computer/io/keyboard.go` | `KeyboardAdapter` + `Keyboard` goroutine; maps key codes to memory-mapped register |
| `simple-computer/io/display.go` | `DisplayAdapter` + `ScreenControl`; writes pixels, renders at 30fps |
| `simple-computer/computer/computer.go` | `SimpleComputer` — wires all subsystems, enforces memory map, owns run loop |
| `simple-computer/asm/parser.go` | Regex-based parser for the assembly language |
| `simple-computer/asm/assembler.go` | Two-pass assembler; resolves labels and symbols |
| `simple-computer/asm/instructions.go` | Every opcode as a concrete `Instruction` struct |
| `simple-computer/cmd/simulator/main.go` | CLI entry point: loads `.bin`, starts GLFW window + computer goroutine |
| `simple-computer/cmd/assembler/main.go` | CLI entry point: assembles `.asm` → `.bin` |
| `simple-computer/cmd/generator/main.go` | CLI entry point: generates `.asm` programs programmatically |

## File Naming Conventions

**Source files:**
- Named after the hardware component or concept they model: `wires.go`, `gates.go`, `stepper.go`, `register.go`, `bus.go`, `iobus.go`, `adder.go`, `decoders.go`
- Compound components with many types: `components.go`, `big_gates.go` (fan-in gate variants)
- Subsystem files: `cpu.go`, `alu.go`, `memory.go`, `display.go`, `keyboard.go`, `computer.go`
- Assembler files: `parser.go`, `assembler.go`, `instructions.go`, `markers.go`
- Test files: `<name>_test.go` co-located with source (standard Go convention)
- CLI entry points: always `main.go` inside each `cmd/<name>/` subdirectory

**Program files (in `_programs/`):**
- `<program-name>.asm` — source
- `<program-name>.bin` — compiled binary (committed alongside source)

**Directories:**
- All lowercase, single-word or hyphenated: `circuit`, `components`, `alu`, `memory`, `cpu`, `io`, `computer`, `asm`, `utils`, `cmd`
- CLI subdirectories: named by their function — `simulator`, `assembler`, `generator`
- The `_programs/` prefix underscore marks it as a data/resource directory rather than a Go package

## Entry Points

**Simulator** (`simple-computer/cmd/simulator/main.go`):
- Flags: `-bin <file>` (binary to load), `-print-state`, `-print-state-every <n>`
- Loads `.bin` as `[]uint16` (little-endian), creates `SimpleComputer`, starts GLFW loop
- Requires OS main thread lock for GLFW (via `runtime.LockOSThread()` in `init()`)
- Runs computer and keyboard in goroutines; GLFW on main thread

**Assembler** (`simple-computer/cmd/assembler/main.go`):
- Flags: `-i <input.asm>`, `-o <output.bin>`, `-s` (string/human-readable dump)
- Reads `.asm` from file or stdin; writes `.bin` to file or stdout
- `USER_CODE_START = 0x0500`

**Generator** (`simple-computer/cmd/generator/main.go`):
- Arg: program name (`ascii`, `brush`, `text-writer`, `me`)
- Uses `asm` package API directly to build instruction lists programmatically
- Writes `.asm` to stdout (then assembled separately)

## Where to Add New Code

**New logic gate type:** Add to `simple-computer/circuit/gates.go` following the existing pattern (struct with `output Wire`, `NewXGate()`, `Update(...)`, `Output() bool`).

**New multi-bit component:** Add to `simple-computer/components/components.go` if it implements `Component` interface; add to `simple-computer/components/big_gates.go` if it is a fan-in variant of an existing gate.

**New peripheral device:** Create `simple-computer/io/<device>.go`. Implement `Peripheral` interface (`Connect(*IOBus, *Bus)`, `Update()`). Register in `computer/computer.go` via `cpu.ConnectPeripheral(...)`.

**New CPU instruction (hardware):** Modify `simple-computer/cpu/cpu.go` — add control signals to `runStep4/5/6Gates()`, `runEnable()`, and `runSet()` methods. Update opcode comments at top of file.

**New assembly instruction (software):** Add instruction struct to `simple-computer/asm/instructions.go` implementing `Instruction` interface. Add parsing case to `simple-computer/asm/parser.go` `parseInstruction()` switch. Update `INSTRUCTION` regex.

**New example program:** Add `.asm` source to `simple-computer/_programs/`. Add a generation function to `simple-computer/cmd/generator/main.go` and a build target to `simple-computer/_programs/Makefile`.

**Tests:** Place `<file>_test.go` alongside the source file in the same package (all existing tests use this pattern).

---

*Structure analysis: 2026-05-07*
