# Building a Computer from Scratch — Part 1: Wires and Logic Gates

_This is the first post in a series where I rebuild a complete computer in Go, starting from the smallest possible primitive and working upward — one phase at a time. The computer is the one described in **"But How Do It Know?"** by J. Clark Scott. If you've ever wondered what's actually happening inside a computer beyond the "transistors flip bits" handwave, this series is for you._

---

## The Problem

Before you can simulate any computer, you need to answer a deceptively simple question: **how do you represent the state of a wire?**

A real wire is either carrying voltage (on) or not (off). That's one bit. Everything a computer does — adding numbers, storing data, running programs — is ultimately just a massive pile of wires being turned on and off in the right sequence.

So Phase 1 is about building two things:

1. **Something to hold a single on/off state** — a `Wire`
2. **Something to transform one or two states into a new state** — a `Gate`

Without these two primitives, nothing else in the system can exist. No storage cells, no ALU, no CPU. We start here.

---

## Context

The book _But How Do It Know?_ starts by introducing the **NAND gate** as the single physical building block from which all computation emerges. A NAND gate takes two inputs and produces one output: the output is **off only when both inputs are on**. Everything else — AND, NOT, OR, XOR, memory, arithmetic — can be derived from NAND alone.

Here's the NAND truth table:

| A | B | Output |
|---|---|--------|
| 0 | 0 | 1 |
| 1 | 0 | 1 |
| 0 | 1 | 1 |
| 1 | 1 | 0 |

From NAND, you get:
- **NOT**: connect both inputs together → NAND with itself
- **AND**: NAND followed by NOT
- **OR**: NOT each input, then NAND them (De Morgan's law)
- **XOR**: "are the inputs different?" — derivable from combinations of NAND
- **NOR**: NOT followed by OR

Phase 1 implements all six gate types upfront, even though the book introduces OR and XOR later. The reason: Phase 3 (storage) needs NOR, and every phase after that needs OR and XOR. Better to have them all now.

---

## Building It

### The Wire

```go
type Wire struct {
    Name  string
    value bool
}

func NewWire(name string, value bool) *Wire
func (w *Wire) Update(value bool)
func (w *Wire) Get() bool
```

A `Wire` is not just a `bool`. It's a **named, mutable holder**. The name makes debugging easier. The mutability is the key design decision — explained below.

### The Gates

Each gate follows the same shape:

```go
type NANDGate struct{ output Wire }

func (g *NANDGate) Update(a, b bool) { g.output.Update(!(a && b)) }
func (g *NANDGate) Output() bool     { return g.output.Get() }
```

The pattern is always: inputs arrive via `Update()`, output is read via `Output()`. The gate stores its result internally.

Derived gates build on the same idea:

```go
// AND: NAND followed by NOT
func (g *ANDGate) Update(a, b bool) { g.output.Update(a && b) }

// OR via De Morgan: !(NOT a AND NOT b)
func (g *ORGate) Update(a, b bool) { g.output.Update(!(!a && !b)) }

// XOR: NAND-derivable form
func (g *XORGate) Update(a, b bool) { g.output.Update(!((!a && !b) || (a && b))) }
```

### The Tests

For each gate, every combination of inputs is tested:

```go
func TestNANDGate(t *testing.T) {
    cases := []struct{ a, b, want bool }{
        {false, false, true},
        {true, false, true},
        {false, true, true},
        {true, true, false},
    }
    for _, c := range cases {
        g := NewNANDGate()
        g.Update(c.a, c.b)
        if got := g.Output(); got != c.want {
            t.Errorf("NAND(%v,%v) = %v, want %v", c.a, c.b, got, c.want)
        }
    }
}
```

Same pattern for all six gates and for `Wire`. Every truth table row is verified.

---

## Key Insights

### The chain-of-ownership pattern

The most important design decision in Phase 1 is not which gates to implement — it's **how they store their output**.

Each gate holds its result in an internal `Wire`. Callers always read via `.Output()`, never compute the result themselves. This creates a rule: **every component owns its own output**. The layer above only asks; it never computes on behalf of the layer below.

```go
// A caller composing two gates:
not.Update(a)
and.Update(not.Output(), b)
result := and.Output() // reads committed snapshot, not a live formula
```

This might seem like a minor style choice. It's not. The consequence shows up in Phase 3.

### Why internal state is necessary for feedback

Storage cells (like the SR latch we'll build in Phase 3) work by feeding an output back into an input:

```
Q = NAND(S, Q_previous)
```

If `Output()` computed its result on the fly from current inputs, this would be circular: to compute `Q`, you need `Q`, which requires computing `Q` again, infinitely. The program would hang or produce garbage.

With an internal `Wire`, `Output()` returns the **last committed value** — the snapshot from the previous `Update()`. The feedback reads "what was the output last time" rather than "what would the output be right now if we computed it". The circuit settles to a stable state.

```go
// Stable: reads committed snapshot from previous Update()
or.Update(s, or.Output())

// Would be unstable without internal state:
// or.Output() tries to compute from current inputs
// but current inputs haven't settled yet
```

This design decision made in Phase 1 is what makes storage possible in Phase 3.

### The non-obvious XOR

XOR is implemented as:

```go
g.output.Update(!((!a && !b) || (a && b)))
```

Not the more intuitive:

```go
g.output.Update((a || b) && !(a && b))
```

Both produce identical truth tables. The chosen form asks "are the inputs the same?" and inverts — which maps cleanly to physical NAND-gate combinations on real silicon. It's closer to how the hardware actually works.

---

## Conclusion

After Phase 1, we have:

- A `Wire` type that carries a single bit with a name
- Six gate types: `NANDGate`, `ANDGate`, `NOTGate`, `ORGate`, `XORGate`, `NORGate`
- Full truth-table tests proving every input combination
- A consistent pattern — `Update()` in, `Output()` out — that every future component will follow

These 46 lines of implementation code and their tests are the vocabulary the entire system speaks. Every register, every ALU operation, every CPU control signal is eventually composed from these primitives.

---

## What's Next

**Phase 2: Multi-Input Gates**

AND and OR gates take exactly two inputs. But building an 8-bit storage word requires an AND gate with 8 inputs, and a decoder needs one with 16. Phase 2 builds `ANDGate3`, `ANDGate4`, `ANDGate5`, `ANDGate8`, and their OR counterparts — by chaining the two-input gates we just built.

It's a short phase, but it introduces a new package (`components`) and sets up the foundation for storage in Phase 3.
