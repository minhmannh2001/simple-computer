# Codebase Concerns

**Analysis Date:** 2026-05-07

---

## Critical Issues

**The entire `simple-computer/` subdirectory is gitignored and untracked.**
- The `.gitignore` at repo root contains `simple-computer/` on line 2, which means all production code — the CPU, memory, ALU, assembler, display, components, IO — lives only on disk and is never committed to version control.
- Only the root `circuit/` package (gates + wires) and `go.mod` are actually tracked in git.
- Files: `simple-computer/` (entire subtree), `/Users/minhmannh2001/Documents/codespace/minhmannh2001/simple-computer/.gitignore`
- Impact: Any disk failure, accidental deletion, or new clone of the repo loses all ~14,000 lines of the real implementation.
- Fix: Remove `simple-computer/` from `.gitignore`, add and commit the directory.

**Duplicate `circuit/` package at two different module roots.**
- There are two separate, diverging copies of the same gate/wire primitives:
  - `/circuit/` — in the root module `simple-computer` (go 1.26.1)
  - `/simple-computer/circuit/` — in the nested module `github.com/djhworld/simple-computer` (go 1.12)
- The root copy is the only one tracked in git. The nested copy (used by all production code) differs in formatting and method ordering, and the `NORGate.Output()` method is missing from the tracked copy.
- Files: `circuit/gates.go`, `simple-computer/circuit/gates.go`
- Impact: Any future changes to the tracked copy diverge silently from the implementation actually used. The root module (`go.mod` declares `module simple-computer`) has no consumers.
- Fix: Decide on one canonical module. Either consolidate both into a single module or remove the unused root package.

**`computer.LoadToRAM` uses `panic` for invalid address ranges.**
- Files: `simple-computer/computer/computer.go:60,63`
- Impact: An invalid program load crashes the entire simulator with no recovery path, stack trace, or user-friendly message. Calling code has no way to catch this.
- Fix: Return an `error` instead of panicking.

---

## Technical Debt

**Three separate 1,000+ line key-press lookup tables defined twice.**
- `up_key_presses` (1,024 entries) is defined identically in two files:
  - `simple-computer/io/keyboard.go:158`
  - `simple-computer/cmd/simulator/glfw_io.go:162`
- `down_key_presses` (1,024 entries) is defined only in `simple-computer/cmd/simulator/glfw_io.go:1189`.
- These tables are pure data — every entry is a hand-written struct literal. If the key range or struct fields change, both copies must be updated.
- These tables account for ~2,000 of the 2,214 lines in `glfw_io.go`.
- Fix: Move key-press tables into a shared package; generate them programmatically (e.g., `for i := 0; i < 1024; i++`), or at minimum deduplicate.

**`LeftShifter.Update` and `RightShifter.Update` are fully unrolled loops.**
- Files: `simple-computer/components/components.go:86-104`, `simple-computer/components/components.go:136-153`
- 16 lines of identical `l.outputs[N].Update(l.inputs[N+1].Get())` statements instead of a single loop.
- The same unrolling is repeated for `BusOne.Update` (lines 547-579).
- Impact: Any change to `BUS_WIDTH` requires manually updating these blocks.
- Fix: Replace with `for i := 0; i < BUS_WIDTH-1; i++` loops.

**Commented-out `parseDataInstruction` function left in parser.**
- Files: `simple-computer/asm/parser.go:272-299`
- An entire prior implementation of `parseDataInstruction` is commented out (28 lines). The replacement immediately follows.
- Fix: Remove the dead code.

**`// TODO not sure if this is exactly how this should look...` on `LeftShifter`.**
- Files: `simple-computer/components/components.go:57`
- Design uncertainty is flagged but left unresolved.

**`//TODO update these to use the helper methods` in keyboard adapter.**
- Files: `simple-computer/io/keyboard.go:88`
- `andGate2` reads raw `ioBus.GetOutputWire()` indices directly instead of calling `ioBus.IsSet()` / `ioBus.IsDataMode()` etc., while the rest of the adapter uses the helper methods. Inconsistent and fragile.

**`//TODO get value from symbols map....` in DATA instruction emit.**
- Files: `simple-computer/asm/instructions.go:64`
- The code already does resolve the symbol; the comment is stale but creates confusion about whether the current logic is correct.

**`go 1.12` module declaration is severely outdated.**
- Files: `simple-computer/go.mod:3`
- The module declares `go 1.12` (released 2018). Current language features (generic `for range`, improved type inference) and standard library APIs are unavailable. The `for i, _ := range` idiom appears 37 times throughout the codebase — modern Go (1.22+) allows `for i := range` directly.

---

## Security Concerns

**No input validation on binary file loading.**
- Files: `simple-computer/cmd/simulator/main.go:65-85`
- The simulator reads a raw binary file and loads it directly into RAM. There is no validation of file size beyond ensuring it is even-byte. A malformed or adversarial `.bin` file can fill the entire 64K memory arbitrarily.
- This is a local tool so the risk is low, but worth noting if the simulator is ever exposed to untrusted input.

**Key-press channel is unbuffered — GLFW keyboard callback can block the main thread.**
- Files: `simple-computer/cmd/simulator/main.go:44`, `simple-computer/cmd/simulator/glfw_io.go:59,64,66`
- `keyPressChannel := make(chan *io.KeyPress)` — unbuffered. The GLFW key callback sends on this channel synchronously from the main OS thread. If the keyboard goroutine is slow to read, the callback blocks, which can stall GLFW event processing and cause the window to freeze.
- Fix: Buffer the channel (`make(chan *io.KeyPress, 64)`) or use a non-blocking send.

---

## Performance Concerns

**`memory.Memory64K.String()` iterates all 65,536 cells every call.**
- Files: `simple-computer/memory/memory.go:117-143`
- Formats all 256×256 = 65,536 memory cells into a string on every invocation. This is called indirectly when debug state printing is enabled. At high step rates (`-print-state-every 512`) this produces enormous output and is extremely slow.
- Impact: Enabling `--print-state` degrades performance severely.

**`ScreenControl.Update()` touches all 4,800 display RAM addresses per frame.**
- Files: `simple-computer/io/display.go:189-226`
- Each frame scans 160×30 = 4,800 addresses sequentially, calling `setOutputRAMAddress` (4 memory register operations) and `renderPixelsFromRAM` (enable + update + disable) per address. At 30fps this is ~144,000 register-level operations per second just for display readout, all on the screen control goroutine.
- No dirty tracking or partial updates exist.

**Entire ALU runs all sub-operations every clock cycle regardless of opcode.**
- Files: `simple-computer/alu/alu.go:205-258`
- `updateComparator()` always runs. The `switch` selects which unit feeds the output, but the comparator is unconditionally computed every cycle regardless of whether a CMP instruction is executing.
- Low impact in simulation, but reflects a pattern of over-eagerness in the update loop.

**`time.Tick` used in four places, creating unrecoverable goroutine leaks.**
- Files: `simple-computer/io/display.go:169`, `simple-computer/io/keyboard.go:143`, `simple-computer/cmd/simulator/glfw_io.go:36`, `simple-computer/cmd/simulator/main.go:60`
- `time.Tick` returns a channel backed by a goroutine that is never garbage-collected. In this application all four are created at startup and run for the process lifetime, so it does not cause observable leaks, but it is against Go best practices (`time.NewTicker` with `defer ticker.Stop()` is the correct form).

---

## Scalability Concerns

**`BUS_WIDTH = 16` is a hardcoded constant duplicated across four packages.**
- Defined independently in: `simple-computer/components/components.go:8`, `simple-computer/cpu/cpu.go:259`, `simple-computer/alu/alu.go` (implied), `simple-computer/memory/memory.go:11`, `simple-computer/io/keyboard.go:11`
- Every component assumes 16-bit wires. Changing the bus width requires updating every package separately with no compile-time enforcement that they match.

**Only 4 general-purpose registers; ISA has no mechanism to extend.**
- Files: `simple-computer/cpu/cpu.go:299-308`
- R0–R3 are the only addressable registers. The instruction format encodes register selection in 2 bits. Any extension requires a redesigned ISA.

**Memory is hard-coded as 64K words (65,536 cells).**
- Files: `simple-computer/memory/memory.go:46-68`
- The 256×256 cell array is statically allocated at construction. The display RAM (`simple-computer/io/display_ram.go`) also allocates a full 256×256 array, for a combined ~131,072 cell objects at startup.

---

## Missing Infrastructure

**No CI/CD pipeline.**
- No `.github/workflows/`, no `Dockerfile`, no build automation beyond the `Makefile`.
- The `Makefile` uses `@@go build` (double `@` suppresses echo), which is non-standard and will silently do nothing if `go` is not on PATH.
- Files: `simple-computer/Makefile`
- Fix: Add a GitHub Actions workflow that runs `go test ./...` on push.

**No test coverage enforcement or coverage reporting.**
- `go test ./...` is available via `make test` but no coverage threshold is defined.

**No structured logging — mix of `log` and `fmt.Println`.**
- `log.Println` / `log.Printf` used in IO and computer packages for runtime events; `fmt.Println` used in the CPU state dump path (`computer/computer.go:108-112`) and assembler debug path (`asm/assembler.go:35`).
- There is no log level, no way to silence output in tests, and no machine-readable log format.

**Debug `fmt.Println(a.symbols)` left in production assembler code.**
- Files: `simple-computer/asm/assembler.go:35`
- When symbol resolution fails, the entire symbol map is dumped to stdout. This is clearly a debug artifact.

---

## Code Quality Issues

**`NORGate` is implemented and tested but never used in any circuit.**
- The gate exists in `simple-computer/circuit/gates.go:93-106` and is tested in `simple-computer/circuit/gates_test.go:95-118`, but no component in `components/`, `alu/`, `cpu/`, `io/`, or `computer/` instantiates it.
- This is dead code — either it should be removed or its absence in the design should be documented.

**`IsZero.Update()` traverses the ORer outputs with an early-return shortcut that bypasses the zero path.**
- Files: `simple-computer/components/components.go:188-207`
- When any ORer output is true, it sets `notGate(true)` and returns. When all are false, it sets `notGate(false)` in the loop body and then calls `z.output.Update(z.notGate.Output())`. The logic is correct but the structure — mixing a loop with early return and post-loop update — is non-obvious and fragile to maintain.

**`getNextExecutableInstructionLoc` has a broken loop that only ever counts the current instruction.**
- Files: `simple-computer/asm/assembler.go:167-191`
- The loop iterates from `currentInstrIndex` forward, but the inner `else` block only adds `instruction.Size()` when `currentInstrIndex == i` (i.e., the first iteration), then immediately `break`s on the second non-label instruction. The function therefore never correctly computes the location of the *next* instruction when labels interleave — it always returns `currentOffset + currentInstruction.Size()`, which happens to be correct for the common case but is misleadingly complex code.

**`Decoder4x16.Index()` accumulates with `+=` instead of `=`.**
- Files: `simple-computer/components/decoders.go:171-176`
- In `Update()`, when multiple outputs are set (which should not happen in a well-formed decoder), `d.index += i` accumulates rather than assigning. In practice only one output should ever be hot, so the bug is latent, but `d.index = i` would be the correct and clearer form.

**`IOBus.IsEnable()` is inconsistently named — all other peripherals use `IsEnabled`-style names.**
- Files: `simple-computer/components/iobus.go:48`
- The method is `IsEnable()` while the corresponding state-check methods are `IsSet()`, `IsInputMode()`, `IsOutputMode()`, `IsDataMode()`, `IsAddressMode()`. The keyboard adapter does not use this method at all (reads raw wire index instead).

**`Wire.Name` is public but never used outside the `circuit` package.**
- Files: `simple-computer/circuit/wires.go:4`
- `Name string` is exported, but no code outside the package reads or sets names on wires after construction. The `NewWire("O", false)` calls throughout `gates.go` use a hardcoded name. This field is effectively dead public API.

---

## Risk Areas

**CPU `step()` calling order is load-bearing and undocumented.**
- Files: `simple-computer/cpu/cpu.go:564-586`, `simple-computer/cpu/cpu.go:594-637`
- The `updateStates()` method calls 13 component updates in a fixed sequence (IAR → MAR → IR → RAM → TMP → FLAGS → BUS1 → ALU → ACC → R0–R3). This ordering encodes the data-flow graph of the CPU. There are no comments explaining why each component appears where it does. Inserting a new peripheral or reordering calls during refactoring will silently break instruction execution.

**`DisplayAdapter` maintains toggle state (`writeToRAM` bit) that depends on paired write calls.**
- Files: `simple-computer/io/display.go:104-148`
- The adapter alternates between writing to the address register and writing to display RAM using a flip-flop bit. If the CPU sends an odd number of writes (e.g., due to a bug in a user program or an interrupt), the toggle gets out of phase permanently, corrupting all subsequent display output with no error or recovery.

**Assembler's `CALL` instruction resolves to a label jump but has no matching `RET`.**
- Files: `simple-computer/asm/instructions.go` (CALL type), `simple-computer/asm/parser.go:131`
- `CALL` is parsed and emits as a labelled jump, but there is no corresponding `RET` instruction in the ISA. User programs must manually manage return addresses in registers, making call depth practically limited to 1 level without significant bookkeeping.

**Channel-based quit mechanism allows multiple goroutines to receive a single quit signal.**
- Files: `simple-computer/cmd/simulator/main.go:46`, `simple-computer/computer/computer.go:92-116`, `simple-computer/io/display.go:175-187`, `simple-computer/io/keyboard.go:142-156`
- `quitChannel` is buffered with size 10. The screen control, keyboard, and GLFW loop all select on it. On close, GLFW calls `close(quitChannel)` which broadcasts to all receivers. This is the correct Go pattern, but the `Run()` loop in `computer.go` does not select on `quitChannel` at all — it only drains `tickInterval`. The computer goroutine never terminates on quit; it blocks waiting for the next tick.

---

*Concerns audit: 2026-05-07*
