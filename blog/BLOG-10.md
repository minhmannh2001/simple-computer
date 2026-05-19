# Building a Computer from Scratch — Part 10: The Stepper

_In Parts 1–9, we built wires, gates, storage cells, a shared bus, bitwise operations, comparison logic, decoders, a 16-bit adder, and a register that connects storage to the bus. This phase introduces the Stepper — a component that counts through six positions and tells every other component which moment in time it is._

---

## The Problem

Everything we've built so far is stateless with respect to time. Give the adder two numbers and it adds them. Give the register a value and it stores it. But there's no concept of "do this *now*, then do *that* next." Nothing controls the order of operations.

Imagine an assembly line where ten workers stand side by side. Part 1 welds, Part 2 paints, Part 3 inspects — but with no signal saying whose turn it is, all ten start working at once on whatever is in front of them. The result is chaos. What the line needs is a rotating green light — only the worker whose station is currently lit does their job. Everyone else waits.

The Stepper is that rotating green light. It cycles through six stations (step 0 through step 5), activating exactly one at a time. Every other component in the system watches these step wires and acts only during its assigned step.

Without the Stepper, there is no "first", "then", or "after." The computer cannot execute anything in order.

---

## Context

The book _But How Do It Know?_ introduces this across three chapters: "The Clock", "Step by Step", and "Doing Something Useful."

The clock is a wire that flips between off and on at a regular rhythm — like a heartbeat. Each full beat (off → on → off) is one clock cycle. The stepper uses this rhythm to move a token forward: on one phase of the clock, the first storage cell in a pair latches a new value; on the other phase, the second cell copies it. Together these two cells form a single "step" in the sequence.

Chaining six such pairs produces a shift register: a token that moves one position forward each full clock cycle, loops back to the start after six steps, and ensures only one position is active at any moment.

---

## Building It

### The Structure

```go
type Stepper struct {
    bits           [12]Bit
    reset          circuit.Wire
    resetNotGate   circuit.NOTGate
    clockIn        circuit.Wire
    clockInNotGate circuit.NOTGate
    inputOrGates   [2]circuit.ORGate
    outputs        [7]circuit.Wire
    outputAndGates [5]circuit.ANDGate
    outputOrGate   circuit.ORGate
    outputNotGates [6]circuit.NOTGate
}
```

Twelve `Bit` latches, organized as six pairs. Each pair is one step. Two OR gates control *when* each half of a pair updates. Six NOT gates and five AND gates derive the six visible output signals from the internal state.

### The Clock-Gated Enable Signal

The key to the shift register is that the two `Bit`s in each pair update at *different* moments:

```go
s.inputOrGates[0].Update(s.reset.Get(), s.clockInNotGate.Output()) // high when clock LOW or reset
s.inputOrGates[1].Update(s.reset.Get(), s.clockIn.Get())           // high when clock HIGH or reset
```

- When the clock is **LOW**: `inputOrGates[0]` goes high — all **even** bits (the "masters") can latch new values. The odd bits hold.
- When the clock is **HIGH**: `inputOrGates[1]` goes high — all **odd** bits (the "slaves") copy from their masters. The even bits hold.

This is a master-slave arrangement. A value travels through one pair like this:
1. Clock LOW: the master latches what came before it.
2. Clock HIGH: the slave copies the master.

The slave's output is stable for the full high half-cycle before the next master latches. There's no race condition, no momentary overlap.

### Propagating the Token

```go
s.bits[0].Update(s.resetNotGate.Output(), s.inputOrGates[0].Output())
s.bits[1].Update(s.bits[0].Get(), s.inputOrGates[1].Output())
for i := 1; i < 6; i++ {
    s.bits[2*i].Update(s.bits[2*i-1].Get(), s.inputOrGates[0].Output())
    s.bits[2*i+1].Update(s.bits[2*i].Get(), s.inputOrGates[1].Output())
}
```

Pair 0's master (`bits[0]`) has a special input: `NOT(reset)`. When no reset is happening, `NOT(false) = true` — so this master always sees a `true` input. It latches `true` on every clock-LOW phase, keeping a constant `true` value at the head of the chain.

Each subsequent master reads from the slave of the previous pair. The `true` value flows like a bucket brigade: master[0] → slave[0] → master[1] → slave[1] → … → slave[5].

### Computing the Step Outputs

The outputs are not the raw bit values. They're derived from combinations that enforce mutual exclusivity:

```go
for i := range 6 {
    s.outputNotGates[i].Update(s.bits[2*i+1].Get()) // NOT(slave[i])
}

s.outputOrGate.Update(s.reset.Get(), s.outputNotGates[0].Output()) // step 0
s.outputs[0].Update(s.outputOrGate.Output())

for i := range 5 {
    s.outputAndGates[i].Update(s.bits[2*i+1].Get(), s.outputNotGates[i+1].Output())
    s.outputs[i+1].Update(s.outputAndGates[i].Output())
}
```

- **Steps 1–5**: `AND(slave[N], NOT(slave[N+1]))`. Step N is active when its own slave is true *and the next slave is not yet true*. The moment the token advances to step N+1, step N goes dark.
- **Step 0**: `OR(reset, NOT(slave[0]))`. This reads "active when reset is firing, OR when slave[0] hasn't been claimed yet." Since `slave[0]` starts false at power-on, step 0 is immediately active — no initialization needed.

### The Automatic Reset

The seventh output wire is the sentinel:

```go
s.outputs[6].Update(s.bits[11].Get()) // fires when token reaches last slave
```

When `bits[11]` (the slave of pair 5) goes true, the sentinel fires. The `Update` function catches this and immediately clears the entire shift register by calling `step()` a second time with `reset=true`:

```go
func (s *Stepper) Update(clockIn bool) {
    s.clockIn.Update(clockIn)
    s.reset.Update(s.outputs[6].Get())
    s.step()
    if s.outputs[6].Get() {
        s.reset.Update(true)
        s.step()
    }
}
```

With `reset=true`, both OR gates go high simultaneously — all twelve bits are enabled at once. And because `NOT(reset) = false` is the input fed to `bits[0]`, every bit latches `false`, clearing the entire chain. After clearing, `outputs[0] = OR(reset=true, NOT(slave[0]=false)) = true`: step 0 is live before `Update` returns.

### Constructor

```go
func NewStepper() *Stepper {
    s := &Stepper{}
    for i := range 12 {
        s.bits[i].gates[3].Update(false, false) // bring each latch to stable hold-false state
    }
    s.step()
    return s
}
```

The `Bit` value-types embedded in the array must be bootstrapped to a stable initial state — the same reason `NewBit()` primes `gates[3]` in Phase 3. Since we're in the same package, we can access this field directly. The initial `step()` call with `clockIn=false` and `reset=false` has `NOT(reset)=true` as input to `bits[0]`, so the master latches `true` and the outputs are computed — step 0 active, all others dark.

---

## Key Insights

### Step 0 is the natural resting state

Steps 1–5 are active via AND gates that require their slave to be `true`. Step 0 uses an OR gate that activates when slave[0] is `false`. At power-on, all slaves are false, so step 0 is immediately active without any special startup sequence. It's the default, not the exception.

### Without the double step(), the stepper would take 7 cycles

When the sentinel fires and sets `reset=true`, a naïve implementation would let that state persist until the *next* `Update` call — effectively spending one full clock cycle stuck in the "resetting" position. The consumer of the stepper would see a 7th ghost step. The fix is to call `step()` again within the same `Update` call. The reset completes instantly and step 0 is live before the function returns. From the outside, every full cycle is exactly six steps — no ghost.

### The OR gate makes reset win unconditionally

During normal operation, step 0's OR formula is `OR(false, NOT(slave[0]))`. When slave[0] is true (step 1 is active), this OR evaluates to false — step 0 is correctly off. But during a reset, the left operand flips to `OR(true, …) = true`, overriding whatever slave[0] says. The step 0 output is restored without needing to wait for slave[0] to clear. Reset takes priority structurally, not by conditional logic.

### Exactly one step is active at all times

The AND formula for steps 1–5 (`AND(slave[N], NOT(slave[N+1]))`) ensures that when slave[N+1] goes true (advancing the token), slave[N] is immediately ANDed with `false` — it turns off in the same `step()` call that turns step N+1 on. There's no window where two outputs are active simultaneously. The NOT-of-next-slave is what enforces the exclusivity.

---

## Conclusion

Phase 10 delivers `Stepper` — a six-position shift register built entirely from the `Bit` latches in Phase 3, arranged into master-slave pairs and clocked by OR-gated enable signals. Exactly one of its six outputs is active at any moment. After six clock cycles, it resets to step 0 automatically and instantly. Every component that needs to do different things at different moments in time will AND its control signals with one of these step outputs — giving the whole machine a shared sense of "now."

---

## What's Next

**Phase 11: The ALU**

With the Stepper providing a sense of time, the next piece is a unit that performs all the operations the computer actually needs: add, shift left, shift right, NOT, AND, OR, XOR, and compare. Phase 11 builds the Arithmetic Logic Unit — eight operations selectable by three control wires, with four flag outputs (carry, larger, equal, zero) that capture properties of the result. It draws on the adder from Phase 8, the bitwise operations from Phase 5, and the comparator from Phase 6, combining them behind a single Update interface.
