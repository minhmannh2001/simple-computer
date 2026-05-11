# /implement-phase

Implement a phase of the simple-computer project end-to-end: read spec → TDD → implement → test → update state → write blog → commit.

## Usage

```
/implement-phase <N>
```

Example: `/implement-phase 3`

---

## What to do

### Step 1 — Read context

Read these files before writing a single line of code:

1. `PHASE-NN-<name>.md` for the requested phase number (e.g. `PHASE-03-storage-primitives.md`)
2. `.planning/STATE.md` — understand what's already been built and what decisions were made
3. `.planning/ROADMAP.md` — confirm the phase name and package
4. All existing source files in packages this phase depends on (e.g. `circuit/` for phase 2+, `components/` for phase 3+)
5. `blog/BLOG-01.md` through `blog/BLOG-NN.md` for all completed phases — **read every blog post** to know exactly which concepts have already been introduced to the reader

### Step 2 — Write tests first (TDD)

Create the test file before the implementation file.

- File location: `<package>/` directory specified in the phase spec
- File name: match the source file name with `_test.go` suffix
- Cover every case listed in the "Tests to Write" section of the phase spec
- Tests will fail at compile time until the implementation exists — that's correct

### Step 3 — Implement

Write the implementation to make the tests pass.

- Follow the exact API (types, function signatures) defined in the phase spec's "What to Implement" section
- Use types from the packages already built — do not import Go standard library beyond what's necessary
- No logic copied from `simple-computer/` reference folder — write from understanding
- No comments explaining what the code does — only add a comment if the WHY is non-obvious (e.g. a surprising derivation like XOR's NAND form)

### Step 4 — Run tests

```bash
go test ./<package>/...
```

All tests must pass before continuing. If any fail, fix the implementation — do not modify tests to match wrong output.

### Step 5 — Write the blog post

Create `blog/BLOG-NN.md` using this exact structure:

```
# Building a Computer from Scratch — Part N: <Phase Name>

_One-sentence recap of what previous phases built, then what this phase adds._

---

## The Problem

[Why does this phase have to exist? What breaks or becomes impossible without it?
Use a concrete, everyday analogy — not a computer science term.
Do NOT reference any concept from a future phase (CPU, clock, ALU, register, etc.)
Only use concepts the reader already knows from previous blog posts.]

---

## Context

[What book chapter covers this. What we're building and why it matters for the layer above.
Still no future-phase jargon.]

---

## Building It

[Step-by-step walkthrough. Show the key code. Explain the shape/pattern.
Every concept referenced must be one the reader already saw in a previous phase or this one.]

---

## Key Insights

[2-4 non-obvious observations:
- What would break if done differently
- A surprising implementation choice
- Why the simpler alternative doesn't work
Keep grounded — no references to components that haven't been built yet.]

---

## Conclusion

[What we now have. What this phase unlocks for the next one. One short paragraph.]

---

## What's Next

[One paragraph preview of the next phase. Name it, describe the problem it solves,
but only in terms of what the reader now knows.]
```

**Blog writing rules — enforce strictly:**

- Read every existing blog post (`blog/BLOG-01.md` through the latest) before writing
- Build a mental list of every concept introduced so far across all posts
- When writing the new post, only use concepts from that list PLUS the new concepts introduced in this phase
- If you find yourself about to write a word like "CPU", "ALU", "register", "clock", "instruction", "stepper", "memory" — stop and ask: has this been introduced in a previous blog? If not, do not use it. Find an analogy or simpler language instead
- The reader knows: wires, on/off state, gates (AND/OR/NOT/XOR/NAND/NOR), the Update/Output pattern, chaining gates, and everything from previous completed phases
- Concrete everyday analogies are always better than jargon

### Step 6 — Update STATE.md

Edit `.planning/STATE.md`:

- Change "Active phase" to the next phase number and name
- Change "Last completed" to this phase
- Increment "Overall progress" count
- Add a new "### Phase NN — name ✅" section under "Completed Phases" with:
  - Commit message
  - Package name
  - What was delivered (bullet list)
  - Key insight discovered during implementation
  - Blog link

### Step 7 — Update ROADMAP.md

Edit `.planning/ROADMAP.md`:

- Change the phase row from `⬜ Not started` to `✅ Done`
- Add `✅ blog/BLOG-NN.md` in the Blog column

### Step 8 — Commit

Two commits:

```bash
# First: the implementation
git add <package>/ && git commit -m "phase-NN: <phase-name>"

# Second: blog + state updates
git add blog/BLOG-NN.md .planning/STATE.md .planning/ROADMAP.md
git commit -m "blog: phase-NN blog post + state update"
```

---

## Hard constraints

- Never reference concepts from future phases in the blog
- Never copy code from `simple-computer/` reference folder
- Tests must pass before committing
- Do not skip any step
- If the phase spec says "read these chapters first" — note that in the blog Context section
