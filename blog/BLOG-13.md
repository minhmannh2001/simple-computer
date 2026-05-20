# Phase 13: IOBus & Keyboard

## Core Problem

65,536 possible memory addresses and only one 16-bit bus. How does the keyboard know the CPU is talking to *it* specifically? And how does RAM know to sit quiet during an I/O cycle?

## The Two-Bus Trick

The answer is a second, 4-wire bus running alongside the main 16-bit bus. This `IOBus` carries four control signals independently:

- **SET** — CPU is writing to a device
- **ENABLE** — CPU is reading from a device
- **MODE** — output (CPU→device) or input (device→CPU)
- **DATA_OR_ADDRESS** — is the main bus carrying an address or data payload?

RAM only responds when the IOBus is idle. Peripherals only respond when the IOBus is active. Same 16-bit wire, no ambiguity.

## Address Detection in Hardware

The keyboard lives at address `0x000F` = `0000 0000 0000 1111`. Detecting this exact pattern requires no software — just gates:

- Wires 0–3 (the four MSBs, all 0) feed through NOT gates → become 1s
- Wires 12–15 (the four LSBs, all 1) feed directly → already 1s
- An ANDGate8 combines all eight → fires only when the address matches

This is address recognition in pure hardware. No comparisons, no branching.

## The Timing Problem: Why `memoryBit` Exists

The CPU's I/O sequence takes two separate clock phases:

1. **Address phase**: CPU puts `0x000F` on the bus with SET + ADDRESS + OUTPUT
2. **Data phase**: CPU flips to ENABLE + DATA + INPUT to read the response

Between these two phases the main bus *changes state* — it no longer holds `0x000F`. Without memory, the keyboard would forget it was selected the moment the bus value shifted. `memoryBit` is a single SR latch that bridges this gap: it fires once on address detection and holds "I am selected" across the transition into the data phase.

## The Two-Phase Handshake

```
Phase 1 (CPU → keyboard): mainBus=0x000F, IOBus=SET+ADDRESS+OUTPUT
  → andGate1 (address match) = 1
  → andGate2 (SET+ADDRESS+OUTPUT) = 1
  → memoryBit latches: Get() = true

Phase 2 (keyboard → CPU): IOBus=ENABLE+DATA+INPUT
  → andGate3 (ENABLE+NOT(addr)+NOT(mode)) = 1
  → andGate4 (memoryBit AND andGate3) = 1
  → keycodeRegister drives mainBus with last keycode
```

## What Would Break Without the IOBus

If I/O were purely memory-mapped with no IOBus, addresses 0x000F and 0x0007 would be permanently reserved — no program could ever store data there. Worse, there'd be no signal to prevent RAM from responding to those same addresses, so every keyboard read would also trigger a simultaneous (incorrect) RAM access.

The IOBus is a clean separation of concerns: the control plane (IOBus) tells the data plane (main bus) what each cycle means.

## Non-Obvious: `keycodeRegister` as a Value, Initialized in Connect

The `KeyboardAdapter` holds `keycodeRegister components.Register` as a value (not a pointer), but the register needs a reference to `mainBus` — which isn't known until `Connect()` is called. The solution: create the register in `Connect()` with `*components.NewRegister(...)` and assign the dereferenced struct. The embedded pointer fields (`word`, `enabler`, `inputBus`, `outputBus`) remain valid after the copy; the value-type wires (`set`, `enable`) live directly in the struct and are modified in place by Set/Enable/Unset/Disable.
