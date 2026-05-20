# Building a Computer from Scratch — Part 13: I/O Bus & Keyboard

_In Parts 1–12, we built wires, gates, storage primitives, a shared bus, bitwise operations, comparison logic, decoders, a 16-bit adder, a register, a stepper, an ALU, and 64K of RAM. This phase connects the computer to the outside world: a 4-wire control bus that tells devices when the CPU is talking to them, and a keyboard that can inject a keycode onto the main bus on demand._

---

## The Problem

The CPU and RAM communicate through the same 16-bit bus. But RAM alone isn't a computer — a computer also needs to accept input from the outside world. A keyboard, for instance, needs a way to put a keycode onto the bus when the CPU asks for it.

The obvious first instinct is to treat the keyboard like a RAM cell: give it an address, and whenever the CPU wants a keycode, it reads from that address. But this runs into two problems.

**Problem 1: ambiguity.** RAM responds to every address on the bus on every cycle. If the keyboard occupies address `0x000F`, RAM at that address would respond too — both trying to drive the bus at the same time. There needs to be a way to say "this cycle is I/O, not RAM" so the RAM stays quiet.

**Problem 2: timing.** The I/O exchange takes two separate steps: first the CPU announces *which device* it wants to talk to, then it switches modes to actually receive the data. Between those two steps, the device must remember it was selected — but the main bus will have moved on by then.

---

## Context

The book _But How Do It Know?_ covers these ideas in "The Outside World" and "The Keyboard."

"The Outside World" introduces memory-mapped I/O: instead of dedicated port instructions, devices are assigned addresses on the same bus as RAM. But to keep RAM out of I/O cycles, a separate set of control wires — the I/O bus — signals "this cycle is I/O, not memory." The chapter describes four control wires: set, enable, mode (input/output), and data-or-address.

"The Keyboard" shows how a keyboard peripheral uses the address-recognition pattern to detect when the CPU is addressing it, then drives the bus with the last pressed key.

This implementation adds the `IOBus` to the `components` package, and a new `io` package with a `Peripheral` interface, a `KeyboardAdapter`, and a `Keyboard` goroutine.

---

## Building It

### The IOBus

The `IOBus` is four wires bundled into a struct. Each wire has a name:

```go
const (
    CLOCK_SET       = 0
    CLOCK_ENABLE    = 1
    MODE            = 2 // false=input, true=output
    DATA_OR_ADDRESS = 3 // false=data, true=address
)

type IOBus struct {
    wires [4]circuit.Wire
}
```

`Set()`/`Unset()` drive the CLOCK_SET wire. `Enable()`/`Disable()` drive CLOCK_ENABLE. `Update(mode, dataOrAddress bool)` sets MODE and DATA_OR_ADDRESS together. Query methods like `IsOutputMode()` and `IsDataMode()` just read the appropriate wire.

The RAM ignores the IOBus entirely. Peripherals only respond when the IOBus is active.

### Address Detection in Hardware

The keyboard lives at address `0x000F` = `0000 0000 0000 1111` on the 16-bit bus (wire 0 = MSB):

- Wires 0–3 (MSBs, all zero) feed through NOT gates → become 1
- Wires 12–15 (LSBs, all one) feed directly → already 1
- An `ANDGate8` combines all eight → fires only when the bus carries exactly `0x000F`

No comparisons, no branching — pure gates.

### The `memoryBit` Latch

The CPU's I/O exchange is two-phase:

1. **Address phase:** put `0x000F` on the bus, IOBus = SET + ADDRESS + OUTPUT. The keyboard detects its address and latches "I am selected" into a single `Bit`.
2. **Data phase:** IOBus = ENABLE + DATA + INPUT. The keyboard sees the latch is set and drives the bus with the stored keycode.

Between phases 1 and 2, the main bus changes — it no longer carries `0x000F`. The `memoryBit` bridges this timing gap. It's an SR latch, exactly like the ones in `Word` — but used here as a single-bit state holder across two distinct IOBus states.

### The Gate Logic

```
memoryBit set when:
  andGate1 (address match on mainBus)
  AND andGate2 (SET + DATA_OR_ADDRESS + MODE)

Keycode delivered when:
  andGate4 = memoryBit.Get()
  AND andGate3 (ENABLE + NOT(DATA_OR_ADDRESS) + NOT(MODE))
```

The `notGatesForAndGate3` pair inverts DATA_OR_ADDRESS and MODE so that `andGate3` fires only during the input-data phase — not the output-address phase.

### KeyboardAdapter.Update()

```go
func (k *KeyboardAdapter) Update() {
    // Detect address 0x000F
    for i := 0; i < 4; i++ {
        k.notGatesForAndGate1[i].Update(k.mainBus.GetOutputWire(i))
    }
    k.andGate1.Update(
        k.notGatesForAndGate1[0].Output(), ..., // wires 0-3 inverted
        k.mainBus.GetOutputWire(12), ...,        // wires 12-15 direct
    )

    // Latch trigger: SET + ADDRESS + OUTPUT
    k.andGate2.Update(
        k.ioBus.GetOutputWire(CLOCK_SET),
        k.ioBus.GetOutputWire(DATA_OR_ADDRESS),
        k.ioBus.GetOutputWire(MODE),
    )
    k.memoryBit.Update(k.andGate1.Output(), k.andGate2.Output())

    // Read trigger: ENABLE + NOT(ADDRESS) + NOT(OUTPUT)
    k.notGatesForAndGate3[0].Update(k.ioBus.GetOutputWire(DATA_OR_ADDRESS))
    k.notGatesForAndGate3[1].Update(k.ioBus.GetOutputWire(MODE))
    k.andGate3.Update(
        k.ioBus.GetOutputWire(CLOCK_ENABLE),
        k.notGatesForAndGate3[0].Output(),
        k.notGatesForAndGate3[1].Output(),
    )
    k.andGate4.Update(k.memoryBit.Get(), k.andGate3.Output())

    if k.andGate4.Output() {
        k.keycodeRegister.Set()
        k.keycodeRegister.Enable()
        k.keycodeRegister.Update() // latches from KeyboardInBus, drives mainBus
        k.KeyboardInBus.SetValue(0)
        k.keycodeRegister.Unset()
    } else {
        k.keycodeRegister.Disable()
        k.keycodeRegister.Update()
    }
}
```

### The Keyboard Goroutine

The `Keyboard` struct runs in a goroutine. On a 33ms tick it reads from a channel of key events. When a key-down event arrives, it writes the keycode onto `KeyboardInBus` via `SetValue`. The `KeyboardAdapter` picks it up on the next `Update()` cycle.

---

## Key Insights

### Why the IOBus exists at all

Without the IOBus, there's no way to tell RAM "stay out of this cycle." Every I/O device would share its address with a RAM cell, and both would try to respond to the same address. The IOBus is the signal that lets RAM go silent and lets peripherals wake up — they operate on different control planes even though they share the same 16-bit data bus.

### `memoryBit` bridges two different IOBus states

The keyboard's address-detection gate fires during the first IOBus state (SET+ADDRESS+OUTPUT). The keycode-delivery gate fires during the second (ENABLE+DATA+INPUT). These states are mutually exclusive — the address gate won't fire during the data phase, and vice versa.

Without `memoryBit`, the keyboard would forget it was selected the moment the CPU switched IOBus state. The latch holds "I am the device being addressed" across the transition. It's the same SR latch mechanism from Phase 3, now deployed at the I/O controller level.

### Address detection uses only 8 of 16 bits

`0x000F` has its top 12 bits all zero and bottom 4 bits all one. The ANDGate8 checks only 8 wires: the top 4 (through NOT gates) and the bottom 4 (directly). Bits 4–11 go unchecked.

This is acceptable because only two device addresses exist in this computer (keyboard at `0x000F`, display at `0x0007` in Phase 14). There's no other device that would false-positive on those same 8 bits. Hardware address decoders often use partial decoding like this to reduce gate count when the full address space is sparsely populated.

### `keycodeRegister` is initialized in Connect, not in the constructor

The `KeyboardAdapter` holds `keycodeRegister` as a value type (`components.Register`), but the register needs a pointer to `mainBus` — which isn't available until `Connect()` is called. The solution: create the register in `Connect()` and assign the dereferenced struct:

```go
func (k *KeyboardAdapter) Connect(ioBus *components.IOBus, mainBus *components.Bus) {
    k.ioBus = ioBus
    k.mainBus = mainBus
    k.keycodeRegister = *components.NewRegister("keyboard", k.KeyboardInBus, mainBus)
}
```

Copying the struct is safe because the internal pointer fields (`word`, `enabler`, `inputBus`, `outputBus`) still point to the same heap objects. The value-type wires (`set`, `enable`) live directly in the copied struct and are modified in place.

---

## Conclusion

Phase 13 adds the I/O layer: an `IOBus` with four control wires that lets the CPU signal "this is I/O, not RAM," a `Peripheral` interface for devices that connect to both buses, and a `KeyboardAdapter` that detects its address in hardware, latches the selection across two IOBus phases, and delivers the last keycode onto the main bus when the CPU reads from it. The `Keyboard` goroutine feeds keycodes asynchronously at 30fps.

---

## What's Next

**Phase 14: Display**

The keyboard is an input peripheral — it puts data *onto* the bus. The display is an output peripheral — it takes data *from* the bus and writes pixels to a screen. Phase 14 builds the `DisplayAdapter`, which responds to a different device address (`0x0007`), accepts pixel data across the same two-phase IOBus handshake, and maintains a frame buffer that a screen renderer can read from.
