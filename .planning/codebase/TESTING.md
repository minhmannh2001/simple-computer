# Testing

**Analysis Date:** 2026-05-07

## Test Strategy

**Unit tests only.** No integration or end-to-end tests exist. Every test exercises a single component or gate type in isolation, constructing the component directly and feeding it inputs bit by bit.

Tests follow a bottom-up progression matching the hardware stack: primitives (wires, gates) → composite components (shifters, adders, decoders) → subsystems (ALU, memory, registers) → the CPU fetch-decode-execute cycle.

The `computer` package (`simple-computer/computer/computer.go`) has no test file — the full assembled computer is untested.

## Test Frameworks

**Runner:** Go standard library `testing` package
- No third-party assertion libraries (no `testify`, `gomega`, etc.)
- All assertions written with raw `if` comparisons plus `t.Fail()`, `t.FailNow()`, `t.Logf()`, `t.Log()`, `t.Errorf()`

**Run commands:**
```bash
go test ./...          # run all tests (from simple-computer/ directory)
make test              # equivalent, via Makefile target in simple-computer/Makefile
go test ./circuit/...  # run only circuit package tests
```

No coverage flag is configured in the Makefile. Coverage must be run manually:
```bash
go test -cover ./...
```

## Test File Organization

**Location:** Co-located with source files. Every `_test.go` file sits in the same directory and package as the file it tests.

**Naming:** `<source_file>_test.go` — e.g. `gates_test.go` tests `gates.go`, `alu_test.go` tests `alu.go`.

**Package:** Tests use the same package as the source (`package circuit`, `package components`, `package alu`, etc.), giving full access to unexported fields. There is no `_test` package suffix used anywhere.

**Structure:**
```
simple-computer/
├── circuit/
│   ├── gates.go
│   └── gates_test.go          # tests all 6 gate types
├── components/
│   ├── big_gates.go
│   ├── big_gates_test.go      # truth-table tests for multi-input gates
│   ├── components.go
│   ├── components_test.go     # DummyComponent + Enabler/Shifter/Comparator tests
│   ├── adder.go
│   ├── adder_test.go
│   ├── decoders.go
│   ├── decoders_test.go
│   ├── register.go
│   ├── register_test.go
│   ├── stepper.go
│   ├── stepper_test.go
│   ├── storage.go
│   └── storage_test.go
├── alu/
│   ├── alu.go
│   └── alu_test.go            # comprehensive op coverage with loops
├── memory/
│   ├── memory.go
│   └── memory_test.go
├── asm/
│   ├── assembler.go           # no direct test (tested via parser/instructions)
│   ├── instructions.go
│   ├── instructions_test.go   # String() and Emit() for all instruction types
│   ├── parser.go
│   └── parse_test.go
├── io/
│   ├── keyboard.go
│   ├── keyboard_test.go       # single scenario test
│   ├── display.go             # no test file
│   ├── display_ram.go         # no test file
│   └── peripheral.go          # no test file
├── cpu/
│   ├── cpu.go
│   └── cpu_test.go            # instruction-level CPU integration tests
└── computer/
    └── computer.go            # no test file
```

## Coverage Approach

**What is tested:**
- All 6 primitive gates (`NANDGate`, `ANDGate`, `NOTGate`, `ORGate`, `XORGate`, `NORGate`) — full truth tables
- All multi-input AND/OR gate variants (`ANDGate3`–`ANDGate8`, `ORGate3`–`ORGate6`) — full truth tables
- `Wire` get/set/update behaviour
- `LeftShifter`, `RightShifter` — selected fixed inputs plus power-of-two sweeps
- `Enabler`, `NOTer`, `ANDer`, `ORer`, `XORer`, `IsZero`, `BusOne`, `Comparator` — fixed scenarios
- `Bit`, `Word` — basic set/hold/propagate
- `Register` — set, enable, disable, bus propagation
- `Stepper` — 0–6 cycles
- `Decoder2x4`, `Decoder3x8`, `Decoder4x16`, `Decoder8x256` — all input combinations
- `Adder`, `Add2` — selected values plus carry-in
- All ALU operations (`ADD`, `SHR`, `SHL`, `NOT`, `AND`, `OR`, `XOR`, `CMP`) — fixed cases + full 16-bit sweep loops for `ADD`, `AND`, `OR`, `XOR`, `CMP`
- `Memory64K` — write/read all 65535 addresses, flag-off write guard
- CPU instruction execution — flags register, LD/ST, JMP, DATA, arithmetic, ALU ops, IO, register moves
- `KeyboardAdapter` — single connect/set/enable scenario
- Assembler parser — label definitions, symbol definitions, instructions (String and Emit)

**No coverage targets configured.** No `go test -coverprofile` in Makefile.

**What is not tested:**
- `computer/computer.go` — full system integration (no test file)
- `io/display.go` and `io/display_ram.go` — display output pipeline (no test file)
- `io/peripheral.go` — peripheral interface (no test file)
- `asm/assembler.go` directly — `Assembler.Process()` and `Assembler.ToString()` have no dedicated tests; they are exercised only indirectly through the instruction tests
- `cmd/` entry points (`simulator`, `assembler`, `generator`) — no tests
- `utils/common.go` — `ValueToString` has no test

## Test Patterns

**Truth-table pattern (gates and multi-input gates):**
Used in `circuit/gates_test.go` (root) and `components/big_gates_test.go`. A `[][]bool` slice encodes all input combinations with the expected output as the last element. A `for` loop iterates, calls `Update()`, and checks the output.

Root `circuit/` version:
```go
// circuit/gates_test.go
func TestNANDGate(t *testing.T) {
    combinations := [][]bool{
        {false, false, true},
        {true, false, true},
        {false, true, true},
        {true, true, false},
    }
    for _, combination := range combinations {
        gate1 := NewNANDGate()
        gate1.Update(combination[0], combination[1])
        if gate1.output.Get() != combination[2] {
            t.Fail()
        }
    }
}
```

Refactored `simple-computer/circuit/` version uses a shared helper:
```go
// simple-computer/circuit/gates_test.go
func testTwoInputGate(t *testing.T, name string, update func(a, b bool), output func() bool, cases [][3]bool) {
    t.Helper()
    for _, c := range cases {
        update(c[0], c[1])
        if output() != c[2] {
            t.Errorf("%s(%v,%v): got %v, want %v", name, c[0], c[1], output(), c[2])
        }
    }
}

func TestNANDGate(t *testing.T) {
    g := NewNANDGate()
    testTwoInputGate(t, "NAND", g.Update, g.Output, [][3]bool{
        {false, false, true},
        {true, false, true},
        {false, true, true},
        {true, true, false},
    })
}
```

**Parametrised helper pattern:**
A private `test<Component>` function accepts inputs, expected outputs, and `*testing.T`. The public test function is a list of calls to this helper. Used in `components/adder_test.go`, `components/components_test.go`, `alu/alu_test.go`, `cpu/cpu_test.go`.

```go
// alu/alu_test.go
func testOp(alu *ALU, op uint16, inputA, inputB uint16, CarryIn bool,
    expectedOutput uint16, expectedEqual, expectedIsLarger, expectedCarry, expectedZero bool, t *testing.T) {
    inputABus.SetValue(inputA)
    inputBBus.SetValue(inputB)
    setOp(alu, op)
    alu.CarryIn.Update(CarryIn)
    alu.Update()
    // assertions with t.Logf + t.FailNow
}

func TestAluAdd(t *testing.T) {
    alu := NewALU(inputABus, inputBBus, outputBus, flagsBus)
    testOp(alu, ADD, 0x0000, 0x0000, false, 0x0000, true, false, false, true, t)
    // many more calls...
}
```

**Exhaustive sweep pattern:**
For commutative operations, the ALU tests sweep all 65535×65535 combinations using nested loop counters, verifying the Go operator matches the simulated hardware result:

```go
// alu/alu_test.go
j := uint16(0xFFFF)
for i := uint16(1); i < 65535; i++ {
    testOp(alu, AND, i, j, false, i&j, i==j, i>j, false, i&j==0, t)
    testOp(alu, AND, j, i, false, j&i, i==j, j>i, false, j&i==0, t)
    j--
}
```

**Map-based string table pattern:**
`asm/instructions_test.go` uses `map[Instruction]string` to verify `String()` output for all instruction variants:

```go
var TABLE map[Instruction]string = map[Instruction]string{
    LOAD{REG0, REG0}: "LD R0, R0",
    LOAD{REG0, REG1}: "LD R0, R1",
    // all 16 LOAD combinations, all ADDs, all ANDs...
}
for instruction, expected := range TABLE {
    if instruction.String() != expected {
        t.Errorf(...)
    }
}
```

**DummyComponent fixture:**
`components/components_test.go` defines a `DummyComponent` struct implementing `Component` to feed wire values into components under test. Reused across `TestEnabler`, `TestWord`, `TestWordWithSetOn`.

```go
type DummyComponent struct {
    wires [BUS_WIDTH]circuit.Wire
    next  Component
}
```

**CPU test setup helpers:**
`cpu/cpu_test.go` provides `SetUpCPU()`, `ClearMem()`, `setMemoryLocation()`, `setRegisters()`, `doFetchDecodeExecute()`, `checkFlagsRegister()`, `checkIAR()`, `checkRegisters()` — a suite of helpers that orchestrate full instruction-level scenarios.

**Assertion style:**
- `t.FailNow()` — used when no message is needed (most gate/component tests)
- `t.Fail()` — non-fatal variant used in `register_test.go` and `big_gates_test.go`
- `t.Logf(...)` + `t.FailNow()` — used when a hex value mismatch needs a message (`adder_test.go`, `alu_test.go`, `components_test.go`)
- `t.Errorf(...)` — used only in `simple-computer/circuit/gates_test.go` (the more modern refactored version) and `asm/instructions_test.go`
- No `t.Fatal(...)` or `t.Error(...)` without format strings observed

## Notable Gaps

**`computer/computer.go` has no tests.**
The `SimpleComputer` type that wires CPU + memory + display + keyboard together is never exercised by the test suite. Full-system behaviour can only be verified by running the simulator binary.

**`io/display.go` and `io/display_ram.go` have no tests.**
The display adapter and display RAM components are untested. These are complex components with address decoding and pixel output logic.

**`asm/assembler.go` has no direct test.**
`Assembler.Process()` is the critical path for the assembler tool but is not called directly from any test. The instruction `Emit()` methods are tested, but the full label-resolution and two-pass assembly pipeline is not.

**Shared package-level state in `alu/alu_test.go` and `cpu/cpu_test.go`.**
Global `var inputABus`, `outputBus`, `flagsBus` (ALU) and `var BUS`, `MEMORY` (CPU) are shared across all tests in those packages. `cpu/cpu_test.go` compensates by calling `ClearMem()` at the start of each test, but `alu_test.go` does not reset bus values between tests, creating a potential ordering dependency.

**`memory/memory_test.go` does not call `t.Fail` on bus mismatch.**
`checkBus()` in `memory/memory_test.go` (line 94) returns a `bool` but the result is never checked inside `TestMemory64KDoesNotUpdateWhenSetFlagIsOff`. The assertion silently does nothing if the bus value is wrong.

**No benchmarks.**
No `Benchmark*` functions exist anywhere. Given the exhaustive sweep tests in `alu_test.go` run hundreds of thousands of iterations, performance characteristics are not formally tracked.

**`utils/common.go` has no test.**
`ValueToString` is used everywhere for debug output but is never tested directly.

**Root `circuit/` package tests are a diverged copy.**
The root `circuit/gates_test.go` and `circuit/wires_test.go` test the root-module copy of the circuit package, not the `simple-computer/` module version. They are maintained separately and have diverged in style (no `testTwoInputGate` helper in root version).

---

*Testing analysis: 2026-05-07*
