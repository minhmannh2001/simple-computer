# Building a Computer from Scratch — Part 2: Multi-Input Gates

_In Part 1, we built the six fundamental gate types — each taking exactly two inputs. That's enough to prove the primitives work, but the CPU we're building toward needs to AND or OR three, four, five, even eight signals at once. This phase handles that._

---

## The Problem

The CPU's control unit is, at its core, a decision machine. Every clock sub-step it asks questions like:

> "Is step 4 active, AND is the instruction an ADD, AND is the carry flag set?"

That's three signals AND-ed together. The book's computer needs conditions that combine up to eight signals simultaneously. With only 2-input AND gates, every call site would have to manually chain them:

```go
and1.Update(step4, isADD)
and2.Update(and1.Output(), carryFlag)
// result is and2.Output()
```

Repeated across dozens of instruction conditions, this becomes noise. The solution is to give those chains names — `ANDGate3`, `ANDGate4`, all the way to `ANDGate8` — and do the same for OR.

---

## Context

The book chapter "More Gate Combinations" shows that chaining two 2-input AND gates produces a 3-input AND gate. The output of the first gate feeds into one input of the second; the result is high only when all three original inputs are high. The pattern extends to any number of inputs.

This is the only new concept in Phase 2. There is no new theory — just the mechanical application of chaining. These types live in a new `components` package that all higher layers will depend on.

> **Note:** The same chapter also covers the Decoder. That comes in Phase 7. Stop reading when you reach it.

---

## Building It

Every type follows the same structure: a fixed set of 2-input gates wired in sequence, all stored as struct fields so they hold their committed output between `Update()` calls.

### ANDGate3

```go
type ANDGate3 struct {
    and1, and2 circuit.ANDGate
}

func (g *ANDGate3) Update(a, b, c bool) {
    g.and1.Update(a, b)
    g.and2.Update(g.and1.Output(), c)
}

func (g *ANDGate3) Output() bool { return g.and2.Output() }
```

Two gates. First AND combines `a` and `b`; second AND combines that result with `c`. Output is true only when all three inputs are true.

### ANDGate8

For 8 inputs, a balanced tree is more efficient than a pure chain — it reduces the maximum gate depth from 7 to 3:

```go
func (g *ANDGate8) Update(a, b, c, d, e, f, h, i bool) {
    g.and1.Update(a, b)   // pair 1
    g.and2.Update(c, d)   // pair 2
    g.and3.Update(e, f)   // pair 3
    g.and4.Update(h, i)   // pair 4
    g.and5.Update(g.and1.Output(), g.and2.Output()) // merge pairs 1+2
    g.and6.Update(g.and3.Output(), g.and4.Output()) // merge pairs 3+4
    g.and7.Update(g.and5.Output(), g.and6.Output()) // final merge
}
```

### OR variants

ORGate3 through ORGate6 follow the identical pattern using `circuit.ORGate`:

```go
func (g *ORGate6) Update(a, b, c, d, e, f bool) {
    g.or1.Update(a, b)
    g.or2.Update(c, d)
    g.or3.Update(e, f)
    g.or4.Update(g.or1.Output(), g.or2.Output())
    g.or5.Update(g.or4.Output(), g.or3.Output())
}
```

### Tests

Every gate is tested against the key boundary conditions:

- **All false → false** (AND) / **all false → false** (OR)
- **Exactly one true → false** (AND) / **exactly one true → true** (OR)
- **All true → true** (both)
- **ANDGate8:** each of the 8 input positions set to false while the rest are true — all 8 must produce false

```
$ go test ./components/...
ok  simple-computer/components  0.617s
```

---

## Key Insights

### Chains versus variadic functions

The most tempting shortcut would be a variadic function:

```go
func ANDMany(inputs ...bool) bool {
    for _, v := range inputs {
        if !v { return false }
    }
    return true
}
```

This is simpler to write and produces the same truth table. But it breaks the hardware-simulation discipline. In real CMOS, there is no "8-input AND gate" as a single component — it is always a tree of 2-input gates on silicon. The struct-of-gates approach mirrors that reality: the committed-output property of each internal gate is what makes the whole chain stable under feedback.

### Balanced tree vs. linear chain

For ANDGate8, a balanced tree (pairs → pairs of pairs → final) is used rather than a left-to-right chain `(((a∧b)∧c)∧d)∧...`. Both are correct. The tree structure more closely mirrors how hardware designers lay out gates to minimize propagation depth, and it's easier to read when you're reasoning about which pair of signals gets combined first.

### The struct fields are wiring documentation

The internal gate fields (`and1`, `and2`, etc.) don't just hold state — they document the wiring topology. Reading the struct definition tells you exactly how many intermediate stages exist and in what order signals combine. A variadic slice would hide that.

---

## Conclusion

Phase 2 delivers 8 named types — `ANDGate3/4/5/8` and `ORGate3/4/5/6` — living in the new `components` package. They're short and mechanical, but they give higher layers a clean vocabulary for expressing multi-signal conditions without manual chaining at every call site.

The `circuit` package is now the foundation of `components`. Every subsequent package — storage, bus, ALU, CPU — will import `components`, not `circuit` directly.

---

## What's Next

**Phase 3: Storage Primitives**

Gates combine signals. But a computer also needs to *remember* signals — to hold a value even after the inputs change. Phase 3 builds `Bit`, the smallest possible memory element, using a feedback loop through two NAND gates (the SR latch from the book's "Remember When" chapter). Then it stacks 16 `Bit`s into a `Word`. These two types are the atomic unit of every register, every RAM cell, and the instruction register in the CPU.
