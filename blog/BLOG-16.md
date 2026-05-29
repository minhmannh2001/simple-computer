# Building a Computer from Scratch — Part 16: The Complete Computer

_In Parts 1–15, we built every individual layer — wires, gates, storage, a shared bus, bitwise operations, comparison logic, decoders, a 16-bit adder, registers, a stepper, an ALU, 64K of RAM, an I/O bus, a keyboard, and a display — and wired all of them into a CPU that fetches, decodes, and executes instructions. This phase connects the CPU to the rest of the hardware into a single runnable machine._

---

## The Problem

After fifteen phases, we have a fully working CPU sitting next to a fully working RAM, a keyboard, and a display. None of them are connected to each other. The CPU needs a bus to exist. The RAM needs the same bus. The keyboard and display need to be registered as peripherals. The memory regions need to be agreed upon. And once you start the clock, something needs to keep it running.

Think of it like having assembled every part of a car — engine, gearbox, wheels, dashboard — and leaving them on the garage floor. The parts are finished. The car isn't. Someone still has to bolt everything together in the right order, wire the ignition to the engine, and put fuel in the tank before it can drive.

`SimpleComputer` is that final assembly step.

---

## Context

The book _But How Do It Know?_ reaches this point in its "Ta Daa!" chapter: the complete machine is shown for the first time as a single diagram with all subsystems wired together. A closing section, "A Few More Words on Arithmetic," discusses the memory map — the convention that the lowest addresses are reserved for the system and user programs start at a fixed offset.

This phase is almost entirely wiring. There is no new hardware. The interesting parts are the startup sequence, the memory convention, and the clock loop.

---

## Building It

### Wiring Order

The `SimpleComputer` struct holds every top-level component:

```go
type SimpleComputer struct {
    mainBus         *components.Bus
    memory          *memory.Memory64K
    cpu             *cpu.CPU
    keyboardAdapter *io.KeyboardAdapter
    displayAdapter  *io.DisplayAdapter
    screenControl   *io.ScreenControl
    screenChannel   chan *[160][240]byte
    quitChannel     chan bool
}
```

The construction order in `NewComputer()` matters:

```go
c.mainBus = components.NewBus(16)
c.memory  = memory.NewMemory64K(c.mainBus)
c.cpu     = cpu.NewCPU(c.mainBus, c.memory)

c.keyboardAdapter = io.NewKeyboardAdapter()
c.cpu.ConnectPeripheral(c.keyboardAdapter)

c.displayAdapter  = io.NewDisplayAdapter()
c.screenControl   = io.NewScreenControl(c.displayAdapter, c.screenChannel, c.quitChannel)
c.cpu.ConnectPeripheral(c.displayAdapter)
```

The bus must exist first because both RAM and the CPU share it — they both need the same pointer to the same 16-wire object. RAM comes before the CPU because the CPU constructor receives the RAM pointer and wires the memory address register into its internal bus routing. Peripherals are connected last because `ConnectPeripheral` reaches into the CPU's I/O bus and wires the peripheral's detection gates to it.

### Memory Regions

The 16-bit address space spans 65,536 addresses. By convention:

- **0x0000 – 0x04FF** — reserved system space. The display adapter lives at address 0x0007 inside this region; the keyboard at 0x000F. User programs must not write here.
- **0x0500 – 0xFEFD** — user code region. Programs are loaded here.
- **0xFEFE – 0xFFFF** — reserved end-of-memory sentinel.

`LoadToRAM` enforces the boundary:

```go
func (c *SimpleComputer) LoadToRAM(offset uint16, values []uint16) {
    if offset < 0x0500 {
        panic("0x0000 - 0x04FF is a reserved memory area")
    }
    if offset > 0xFEFF {
        panic("0xFEFF - 0xFFFF is a reserved memory area")
    }
    for i, v := range values {
        c.putValueInRAM(offset+uint16(i), v)
    }
}
```

Writing a value into RAM is a two-phase operation (identical to the protocol the CPU uses internally): first write the address into the memory address register over the bus, then write the data value:

```go
func (c *SimpleComputer) putValueInRAM(address, value uint16) {
    c.memory.AddressRegister.Set()
    c.mainBus.SetValue(address)
    c.memory.Update()
    c.memory.AddressRegister.Unset()
    c.memory.Update()

    c.mainBus.SetValue(value)
    c.memory.Set()
    c.memory.Update()
    c.memory.Unset()
    c.memory.Update()
}
```

### The Sentinel Jump

When `Run()` starts, before the clock begins, it writes a JMP instruction to addresses `0xFEFE` and `0xFEFF`:

```go
c.putValueInRAM(0xFEFE, 0x0040)          // JMP opcode
c.putValueInRAM(0xFEFF, CODE_REGION_START) // jump target: 0x0500
```

If a program has no explicit loop and the instruction pointer advances past the last instruction, it eventually reaches 0xFEFE. The CPU decodes a JMP instruction there and jumps back to 0x0500, restarting the program. This prevents the CPU from executing garbage data at the top of address space.

### The Clock Loop

`Run()` sets the instruction address register to 0x0500, starts the screen scanner goroutine, then drives the clock:

```go
c.cpu.SetIAR(CODE_REGION_START)
go c.screenControl.Run()

for {
    select {
    case <-c.quitChannel:
        return
    case <-tickInterval:
        c.cpu.Step()
    }
}
```

`tickInterval` is a channel that delivers a value on each clock edge — in a real application, a `time.Tick` at the desired clock frequency. In tests, the channel can be a buffered Go channel that you push values into manually to advance exactly N ticks.

The `quitChannel` lets external code (the OS signal handler, a test harness) stop the loop cleanly. On quit, `Run()` returns, which also halts the screen goroutine (since they share the same `quitChannel`).

---

## Key Insights

### Construction order is a dependency order

The wiring sequence — bus, RAM, CPU, peripherals — mirrors the dependency graph. Each constructor call produces a pointer that the next constructor consumes. Getting this order wrong causes a nil pointer dereference at the first `Update()` call, not at construction, because the wiring happens lazily through shared pointers.

### The sentinel JMP is the simplest possible loop

The alternative — checking whether IAR has reached the end of memory and resetting it in software — would require the clock loop to inspect CPU state and intervene between steps. The sentinel JMP lets the CPU handle it itself. It costs two words of RAM and is transparent to the programmer.

### The screen goroutine must start before the clock loop

The `ScreenControl.Run()` goroutine reads the display RAM at ~30 frames per second and pushes frames to `screenChannel`. If it started after `Run()`'s loop began, the first frame would already be stale. Worse: if the clock loop ran for several milliseconds before the goroutine started, the display would show a burst of updates all at once instead of a smooth stream. Starting the goroutine first ensures it is ready to consume frames before any pixels are written.

### `tickInterval` decouples clock speed from everything else

The computer itself does not know how fast it runs. `Run()` just responds to whatever rate `tickInterval` delivers ticks. This is intentional: the real application sets the interval based on the target frequency; tests set it to a manual channel and push ticks one at a time. Neither needs to change `SimpleComputer`'s code.

---

## Conclusion

Phase 16 delivers `SimpleComputer`: a `computer` package that wires the bus, RAM, CPU, keyboard, and display into a single struct. `LoadToRAM` enforces the memory convention; the sentinel JMP prevents the CPU from falling off the end of useful memory; the clock loop drives execution at whatever frequency the caller chooses. The machine can now load and run any program that fits in the 0x0500–0xFEFD code region.

---

## What's Next

**Phase 17: The Assembler**

Writing programs as raw 16-bit numbers is painful. The instruction `DATA R0, 0x0042` encodes as `0x0020, 0x0042` — and if you want to jump to a label, you have to count the exact word address yourself. Phase 17 adds an assembler: a program that reads human-readable mnemonics like `ADD R0, R1` or `JMP end` and produces the `[]uint16` slice that `LoadToRAM` expects. It is the final piece that makes the machine programmable without counting bits.
