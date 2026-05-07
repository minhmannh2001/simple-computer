# Integrations

**Analysis Date:** 2026-05-07

## External Services

None. This project has no network calls, no cloud services, no SaaS APIs, and no HTTP clients anywhere in the codebase. It is a fully self-contained desktop simulation.

## Databases & Storage

**No databases.** State is held entirely in-process Go memory:
- 64K RAM is simulated as a Go data structure in `simple-computer/memory/memory.go`
- Display RAM (4800 bytes covering a 240×160 framebuffer at 1 bit/pixel, packed into 16-bit words) is in `simple-computer/io/display_ram.go`
- No SQLite, no embedded KV store, no file-backed persistence

**File I/O (binary programs):**
- The simulator reads a `.bin` file (little-endian `uint16` array) from disk at startup via `simple-computer/cmd/simulator/main.go`
- The assembler reads `.asm` text from stdin or a file (`-i` flag) and writes `.bin` to stdout or a file (`-o` flag) via `simple-computer/cmd/assembler/main.go`
- The generator writes `.asm` text to stdout; piped through the assembler to produce `.bin`
- No database, cache, or object store involved at any point

**Example program binaries** are committed to the repository under `simple-computer/_programs/`:
- `ascii.bin`, `brush.bin`, `me.bin`, `text-writer.bin`

## Internal Service Dependencies

There are no microservices, no RPC calls, and no inter-process communication other than the following **intra-process** Go channels used to wire simulator components together:

| Channel | Type | Direction | Purpose |
|---|---|---|---|
| `screenChannel` | `chan *[160][240]byte` | `computer → glfw_io` | Carries rendered frames from the screen control goroutine to the GLFW display goroutine at ~30fps |
| `keyPressChannel` | `chan *io.KeyPress` | `glfw_io → keyboard` | Carries GLFW key events into the simulated keyboard adapter |
| `quitChannel` | `chan bool` (buffered, cap 10) | broadcast | Signals all goroutines to shut down when the GLFW window is closed |

These channels are defined in `simple-computer/cmd/simulator/main.go` and passed into `computer.NewComputer`, `io.NewKeyboard`, and `NewGlfwIO`.

## Integration Patterns

**System I/O (IN/OUT instructions):**
The simulated CPU communicates with peripherals via a dedicated `IOBus` (in `simple-computer/components/iobus.go`). Peripheral adapters are registered with the CPU via `cpu.ConnectPeripheral()`. Two peripherals exist:

- **Keyboard adapter** (`simple-computer/io/keyboard.go`, `simple-computer/io/peripheral.go`) — address `0x000F`
- **Display adapter** (`simple-computer/io/display.go`) — address `0x0007`

CPU assembly programs select a peripheral using `OUT Addr, Rx` (puts the peripheral address on the IOBus) then transfer data using `OUT Data, Rx` or `IN Data, Rx`. This is the only "integration pattern" in the system — all communication is memory-mapped bus operations, not network protocols.

**File format for programs:**
- Input: plain-text assembly (`.asm`) — custom format parsed by `simple-computer/asm/parser.go`
- Output: raw little-endian binary (`.bin`) — `uint16` array written with `encoding/binary.Write`
- No standard formats (ELF, WASM, etc.) are used

## Notable Absences

The following integrations are absent, some intentionally documented in the README:

- **No networking** — the computer is explicitly noted as "incapable of accessing the internet"; no network I/O adapter exists
- **No interrupts** — the README flags this: polling-only keyboard/IO; interrupt controller would require additional bus wiring
- **No persistent storage / hard drive** — explicitly listed as a missing feature
- **No CI/CD pipeline** — no `.github/workflows/`, no Makefile `ci` target, no test coverage enforcement
- **No error tracking / observability** — only `log.Println` / `fmt.Println` to stdout; no structured logging, no metrics, no tracing
- **No environment variable configuration** — no `.env` file, no `os.Getenv` calls anywhere; all configuration is via CLI flags (`-bin`, `-print-state`, `-print-state-every`)
- **No container support** — no `Dockerfile`, no `docker-compose.yml`
- **No package registry publishing** — the module path `github.com/djhworld/simple-computer` is a VCS path, not published to a registry

---

*Integration audit: 2026-05-07*
