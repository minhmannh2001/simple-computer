# Coding Conventions

**Analysis Date:** 2026-05-07

## Naming Conventions

**Types (structs):**
- PascalCase for all types: `Wire`, `NANDGate`, `ANDGate`, `ORGate3`, `ANDGate8`, `Comparator`, `BusOne`, `IsZero`, `Memory64K`, `SimpleComputer`
- Multi-word type names use no separator: `LeftShifter`, `RightShifter`, `NOTer`, `ANDer`, `ORer`, `XORer`
- Multi-input gate variants use a number suffix on the type name: `ANDGate3`, `ANDGate4`, `ANDGate5`, `ANDGate8`, `ORGate3`, `ORGate4`, `ORGate5`, `ORGate6`

**Constructor functions:**
- Always prefixed with `New`: `NewWire`, `NewNANDGate`, `NewANDGate3`, `NewBusOne`, `NewMemory64K`, `NewCPU`, `NewALU`
- Pattern is strictly `New<TypeName>()`

**Methods:**
- `Update(...)` — the universal method to compute and propagate new state through a component; takes current inputs as arguments or reads from wires
- `Output() bool` — single-output read method on gate types
- `Get() bool` / `Set()` / `Unset()` — state accessors on `Wire` and register types
- `Enable()` / `Disable()` — enable control on `Register`, `Enabler`, `BusOne`, `Memory64K`, `IOBus`
- `GetOutputWire(int) bool` / `SetInputWire(int, bool)` — index-based wire I/O for `Component` interface types
- `ConnectOutput(Component)` — chains components together

**Fields:**
- `camelCase` for all struct fields: `output`, `inputA`, `inputB`, `shiftIn`, `shiftOut`, `equalIn`, `equalOut`
- Numbered gate fields use single-letter suffix: `andA`, `andB`, `orA`, `orB`, `xor1`, `not1`
- Arrays of gates named by gate type in plural: `gates [4]circuit.NANDGate`, `andGates [3]circuit.ANDGate`, `notGates`
- Bus and wire arrays: `inputs`, `outputs`, `wires`

**Files:**
- `snake_case` filenames: `big_gates.go`, `display_ram.go`, `glfw_io.go`
- Test files co-located with source: `gates_test.go` alongside `gates.go`
- Command entry points under `cmd/<name>/main.go`

**Constants:**
- `SCREAMING_SNAKE_CASE` for package-level constants: `BUS_WIDTH`, `ADD`, `SHR`, `NOT`, `FLAGS_BUS_CARRY`, `CODE_REGION_START`, `CURRENTINSTRUCTION`, `NEXTINSTRUCTION`
- `iota` used for operation code enumerations in `alu/alu.go` and register names

**Variables in tests:**
- `SCREAMING_SNAKE_CASE` for package-level test vars: `BUS`, `MEMORY`, `inputABus`, `outputBus`, `flagsBus` (inconsistent — see Anti-patterns)

## Code Style

**Formatting:**
- Standard `gofmt` style — inferred from consistent brace placement and indentation
- Tabs for indentation throughout
- No formatter config file present in repo; standard Go toolchain `gofmt` is implied

**Struct declarations:**
- Two styles exist — one-liner (root `circuit/` package) vs. multi-line (main `simple-computer/` package):
  - Root: `type NANDGate struct{ output Wire }` with same-line braces
  - Main: full multi-line struct blocks with fields on separate lines
- Example of root compact style (`circuit/gates.go` line 3): `type NANDGate struct{ output Wire }`
- Example of main expanded style (`simple-computer/circuit/gates.go` lines 3–5): fields on their own lines

**Method chaining style:**
- Method bodies are always separate from the struct declaration
- Simple one-liner methods occur only in the root `circuit/` package: `func (g *NANDGate) Output() bool { return g.output.Get() }`

**Explicit unrolling over loops:**
- Large fixed-width operations are written as explicit sequential assignments rather than loops, e.g. `LeftShifter.Update` and `BusOne.Update` in `simple-computer/components/components.go` each explicitly index all 16 positions instead of iterating

**Range idiom:**
- Mixed: older code uses `for i, _ := range` (deprecated blank identifier form); newer code uses `for i := range`
- Examples of `for i, _ := range` in `components/stepper.go`, `components/decoders.go`, `components/bus.go`
- Examples of `for i := range` in `cpu/cpu.go`, `io/keyboard.go`, `io/display.go`

**Import grouping:**
- Standard library first, then internal packages — no explicit blank-line separation enforced
- Internal packages referenced by full module path: `github.com/djhworld/simple-computer/circuit`

## Error Handling

**Approach:** Errors are returned as `(value, error)` tuples where they can occur. No panic use in library code.

**Patterns observed:**
- `func (a *Assembler) Process(...) ([]uint16, error)` — returns nil slice + wrapped `fmt.Errorf` on failure
- `func (a *Assembler) ResolveLabel(...) (uint16, error)` — canonical Go error-return pattern
- All low-level circuit/component code is error-free by design: gates and wires never fail, so no error returns exist below the assembler layer
- No custom error types; all errors are plain `fmt.Errorf` strings

**No error handling in tests:**
- Test helpers call `t.FailNow()` or `t.Fail()` directly; errors from `Emit()` or `Parse()` in test code are checked with `if err != nil { t.FailNow() }`

## Comment Style

**Inline comments:**
- Used sparingly to explain non-obvious circuit derivations:
  - `// OR via De Morgan: !(NOT a AND NOT b)` in `circuit/gates.go`
  - `// XOR: true only when inputs differ — implemented in NAND-derivable form` in `circuit/gates.go`
  - `// these start out as 1 and 0 respectively` in `components/components.go`
- No doc comments (`//` block above exported types or functions) anywhere in the codebase

**Block comments at file top:**
- `cpu/cpu.go` contains an extensive opcode reference table as a block of `//` comments (lines 13–258) — the only large comment block in the project

**TODO comments:**
- Three TODO comments exist without associated issues:
  - `simple-computer/components/components.go:57` — `// TODO not sure if this is exactly how this should look...` (on `LeftShifter`)
  - `simple-computer/asm/instructions.go:64` — `//TODO get value from symbols map....`
  - `simple-computer/io/keyboard.go:88` — `//TODO update these to use the helper methods`

**No godoc:**
- No package-level doc comments. No exported function or type has a doc comment.

## Patterns Used

**Constructor pattern:**
Every type has a `New<Type>()` factory function that allocates, initialises all sub-gates/wires via `*circuit.NewXxx()` dereferences, and returns a pointer. Sub-components are stored by value inside the parent struct (not as pointers), then dereferenced at construction time.

Example from `simple-computer/components/big_gates.go`:
```go
func NewANDGate3() *ANDGate3 {
    a := new(ANDGate3)
    a.andA = *circuit.NewANDGate()
    a.andB = *circuit.NewANDGate()
    return a
}
```

**Component interface:**
The `Component` interface (`components/components.go`) with three methods — `ConnectOutput`, `SetInputWire`, `GetOutputWire` — is implemented by every bus-connected component. It allows test helpers like `setWireOnComponent16` to operate generically.

**Table-driven tests (manual):**
Truth tables encoded as `[][]bool` slices where the last element is the expected output. Each row is a complete input/output combination. Used throughout `circuit/gates_test.go`, `components/big_gates_test.go`, `components/decoders_test.go`.

**Delegation helper pattern:**
Tests define shared helper functions (`testOp`, `testAdderReturnsCorrectResult`, `testLeftShifter`, `testComparatorReturnsCorrectResult`) that accept expected values and call `t.FailNow()` with a descriptive log. The top-level test function is a list of calls to the helper.

**Dual-phase Enable/Set model:**
All stateful components (registers, memory, IOBus) expose `Enable()`/`Disable()` and `Set()`/`Unset()` separately from `Update()`. The `CPU` orchestrates these in `runEnable()` then `runSet()` each clock half-cycle.

**String method for debug output:**
Most complex components implement `String() string` for human-readable hex debug output (e.g. `Register.String()`, `ALU.String()`, `BusOne.String()`, `Memory64K.String()`), using `utils.ValueToString(uint16)` from `simple-computer/utils/common.go`.

## Anti-patterns Observed

**Duplicate codebase (root vs. simple-computer/):**
The root of the repo contains a separate `circuit/` package (`circuit/gates.go`, `circuit/wires.go`, `circuit/gates_test.go`, `circuit/wires_test.go`) that duplicates the `simple-computer/circuit/` package. These are diverged copies — the root versions use compact one-liner style; the `simple-computer/` versions use expanded style. The root `go.mod` (`go.mod` at repo root) is a separate module with no relation to `simple-computer/go.mod`.

**Deprecated blank identifier in range:**
`for i, _ := range` is used throughout older files (`stepper.go`, `decoders.go`, `bus.go`, `storage_test.go`, `components_test.go`) instead of the idiomatic `for i := range`.

**Shared mutable state in tests:**
`alu/alu_test.go` declares package-level `var inputABus`, `outputBus`, `flagsBus` that are shared across all test functions. Similarly `cpu/cpu_test.go` declares `var BUS`, `var MEMORY`. This means test order can affect results if state is not explicitly reset.

**Explicit 16-step unrolling instead of loops:**
`LeftShifter.Update` and `RightShifter.Update` in `components/components.go` (lines 89–104 and 137–153) and `BusOne.Update` (lines 547–579) write out all 16 index assignments individually rather than using a loop. This is repetitive and fragile if `BUS_WIDTH` ever changes.

**`BUS_WIDTH` redefined per package:**
The constant `BUS_WIDTH = 16` is defined separately in `components/components.go`, `alu/alu.go`, `memory/memory.go`, and `cpu/cpu.go`. There is no shared definition.

**Commented-out code left in production files:**
`cpu/cpu.go` in `refreshFlagStateGates()` (lines 812–819) contains a block of commented-out code that is identical to the live code immediately below it, with no explanation of why it was retained.

**Inconsistent SCREAMING_SNAKE_CASE for test vars:**
`cpu/cpu_test.go` uses `BUS` and `MEMORY` (all-caps), while `alu/alu_test.go` uses `inputABus`, `outputBus`, `flagsBus` (camelCase). Go convention reserves all-caps for exported constants, not variables.

## Enforced Standards

- No linter configuration files detected (no `.golangci.yml`, no `golint` config)
- No pre-commit hooks in `.git/hooks/` or `.claude/` skills
- No CI pipeline detected (no `.github/workflows/`)
- Build and test are run via `Makefile` with `make` (builds) and `make test` (runs `go test ./...`)
- Only enforcement is standard `go build` and `go test` — no static analysis tooling configured

---

*Convention analysis: 2026-05-07*
