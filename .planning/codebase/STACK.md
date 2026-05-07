# Technology Stack

**Analysis Date:** 2026-05-07

## Primary Language(s)

- **Go** — used for all simulation, assembler, and tooling code
  - Root-level module (`simple-computer`): declared as `go 1.26.1` (`go.mod` at repo root), no external dependencies — this is the in-progress re-implementation
  - Reference implementation (`github.com/djhworld/simple-computer`): declared as `go 1.12` (`simple-computer/go.mod`), has external dependencies; README states `go 1.12+` as requirement

## Frameworks & Libraries

There are no web frameworks or application frameworks. The only external dependencies are two OpenGL/windowing bindings used exclusively by the simulator binary:

- **`github.com/go-gl/gl v0.0.0-20190320180904-bf2b1f2f34d7`** — Go bindings for OpenGL 3.2 (compatibility profile). Used in `simple-computer/cmd/simulator/glfw_io.go` to draw pixels directly via `gl.Begin(gl.POINTS)` / `gl.Vertex2i`. This is the only rendering mechanism; there is no sprite library or 2D framework on top of it.
- **`github.com/go-gl/glfw v0.0.0-20190409004039-e6da0acd62b1`** — Go bindings for GLFW 3.2. Used in `simple-computer/cmd/simulator/glfw_io.go` to create the 240×160 window, poll key events, and swap buffers at ~30fps (33ms tick).

All other imports throughout the codebase are Go standard library only: `encoding/binary`, `flag`, `fmt`, `log`, `os`, `runtime`, `strings`, `time`, `io`.

## Build & Tooling

**Build system:**
- `make` via `simple-computer/Makefile`
- Three build targets:
  ```
  go build -o bin/simulator github.com/djhworld/simple-computer/cmd/simulator
  go build -o bin/assembler github.com/djhworld/simple-computer/cmd/assembler
  go build -o bin/generator github.com/djhworld/simple-computer/cmd/generator
  ```
- Programs pipeline: `make` in `simple-computer/_programs/` calls `bin/generator <name>` piped through `bin/assembler` to produce `.bin` files

**Test runner:**
- Go's built-in test runner: `go test ./...` (run via `make test`)
- No external test libraries — all tests use `testing` package from stdlib

**Linter / Formatter:**
- No `.eslintrc`, `biome.json`, `.golangci.yml`, or any linter config detected
- Formatting assumed to be `gofmt` (Go standard); the root-level re-implementation code uses single-line struct style that differs from the reference implementation (sign of deliberate reformatting)

**Lockfile:**
- `simple-computer/go.sum` is present and committed — dependency versions are pinned

## Runtime & Infrastructure

**Runtime:** Native binary — compiled Go executables, no VM or interpreter
**OS requirement:** Desktop OS with GLFW 3.2+ system library installed (`libglfw3`)
**Main thread pinning:** `runtime.LockOSThread()` is called in `simple-computer/cmd/simulator/main.go` `init()` to satisfy GLFW/OpenGL thread requirements
**Concurrency model:** Three goroutines at runtime:
  1. `keyboard.Run()` — reads from `keyPressChannel`
  2. `comp.Run(time.Tick(1*time.Nanosecond), ...)` — CPU simulation loop, clocked at ~1ns tick
  3. `screenControl.Run()` — renders at 33ms (≈30fps), reads display RAM and sends frames on `screenChannel`
  4. Main goroutine runs `glfw.Run()` on the locked OS thread

**Output binaries location:** `simple-computer/bin/` (gitignored)

## Key Dependencies

| Dependency | Version | Why it matters |
|---|---|---|
| `github.com/go-gl/glfw` | `v0.0.0-20190409004039-e6da0acd62b1` | Window creation, keyboard event polling, buffer swap for the simulator UI |
| `github.com/go-gl/gl` | `v0.0.0-20190320180904-bf2b1f2f34d7` | Pixel-level rendering of the 240×160 simulated display using OpenGL points |

Both dependencies are CGO-based and require system libraries (`libglfw3`, OpenGL drivers). The assembler and generator binaries have **zero** external dependencies and compile with pure Go.

## Module Structure

There are two independent Go modules in this repository:

1. **Root module** (`simple-computer`, `go.mod` at `/`):
   - Contains: `circuit/gates.go`, `circuit/wires.go`, and their tests
   - This is the **new, in-progress re-implementation** (only gates and wires exist so far)
   - No external dependencies

2. **Reference implementation** (`github.com/djhworld/simple-computer`, `go.mod` at `simple-computer/`):
   - Contains: the full working simulation — `circuit/`, `components/`, `alu/`, `cpu/`, `memory/`, `io/`, `computer/`, `asm/`, `utils/`, and all three `cmd/` binaries
   - This is the **complete, runnable system** (the upstream reference)
   - Listed in `.gitignore` — the `simple-computer/` directory is tracked but `bin/` is not

---

*Stack analysis: 2026-05-07*
