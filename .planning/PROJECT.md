# PROJECT.md — Simple Computer Rebuild

## What This Is

A ground-up reimplementation of the computer described in **_But How Do It Know?_** by J. Clark Scott (2009) — built in Go, one small phase at a time.

The project starts from the smallest possible primitive (a `Wire` carrying a single bit) and builds upward through logic gates, storage cells, a bus, an ALU, memory, I/O, a full CPU, and finally an assembler. Every layer is composed purely from the layer below it, mirroring how real hardware works.

This is **not** a port of the existing `simple-computer/` reference implementation. It is a clean rebuild written from understanding, following the book chapter by chapter, with tests proving each component before moving on.

## Why

To genuinely understand how a computer works — not at the "transistors flip bits" handwave level, but at the level where you can trace a single clock pulse from the IAR through the fetch-decode-execute cycle and explain what every wire is doing at every sub-step.

The secondary goal is to document that journey in writing, so others following the same book can use it as a companion.

## How It's Built

- **Language:** Go
- **Structure:** Each phase produces one or more Go packages, fully tested before the next phase begins
- **Book:** Follows _But How Do It Know?_ chapter order (see PHASES.md for chapter-to-phase mapping)
- **Difference from book:** The book uses 8-bit bytes and 256-byte RAM. This implementation uses 16-bit words and 65,536-word RAM — matching the reference Go simulator. Every "byte" in the book is a "word" here; all concepts are identical
- **Tests:** Table-driven unit tests per gate / component; tests are written before the implementation (TDD)
- **Blog:** One English-language blog post per phase, stored in `blog/`, written after each phase completes

## Blog Contract

Every phase produces one blog post in `blog/BLOG-NN.md` with this structure:

1. **The Problem** — what challenge this phase solves; why it has to exist before anything else
2. **Context** — what book chapters to read; what we're building and why it matters
3. **Building It** — step-by-step walkthrough of the implementation with key code excerpts
4. **Key Insights** — non-obvious decisions, surprising moments, things that would break if done differently
5. **Conclusion** — what we now have; what this phase unlocks
6. **What's Next** — brief intro to the next phase

## Constraints

- No logic copied from the reference `simple-computer/` — implementations are written from scratch
- Each phase is committed before the next begins
- Blog posts are written in English, stored in `blog/`, not published externally
- The `simple-computer/` reference folder stays gitignored — it is a reference to read, not to copy

## Team

Solo project — one developer learning by building.
