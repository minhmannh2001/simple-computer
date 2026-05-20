package io

import (
	"simple-computer/circuit"
	"simple-computer/components"
	"time"
)

// KeyboardAdapter sits on the main bus and IOBus. It detects when the CPU addresses 0x000F,
// latches that selection in memoryBit, then on the subsequent ENABLE+INPUT cycle drives the
// last keycode onto the main bus.
type KeyboardAdapter struct {
	KeyboardInBus *components.Bus // receives raw keycodes from the Keyboard goroutine

	ioBus           *components.IOBus
	mainBus         *components.Bus
	memoryBit       *components.Bit
	keycodeRegister components.Register

	// Address detection for 0x000F: NOT gates on wires 0-3 (bits 15-12, all 0),
	// direct feed for wires 12-15 (bits 3-0, all 1).
	andGate1            components.ANDGate8
	notGatesForAndGate1 [4]circuit.NOTGate

	// andGate2: SET + DATA_OR_ADDRESS + MODE → latch trigger
	andGate2 components.ANDGate3

	// andGate3: ENABLE + NOT(DATA_OR_ADDRESS) + NOT(MODE) → read trigger
	andGate3            components.ANDGate3
	notGatesForAndGate3 [2]circuit.NOTGate

	// andGate4: memoryBit AND andGate3 → deliver keycode
	andGate4 circuit.ANDGate
}

func NewKeyboardAdapter() *KeyboardAdapter {
	return &KeyboardAdapter{
		KeyboardInBus: components.NewBus(components.BUS_WIDTH),
		memoryBit:     components.NewBit(),
	}
}

func (k *KeyboardAdapter) Connect(ioBus *components.IOBus, mainBus *components.Bus) {
	k.ioBus = ioBus
	k.mainBus = mainBus
	k.keycodeRegister = *components.NewRegister("keyboard", k.KeyboardInBus, mainBus)
}

func (k *KeyboardAdapter) Update() {
	// Step 1: detect address 0x000F on mainBus.
	// Wires 0-3 (bits 15-12) must be LOW; wires 12-15 (bits 3-0) must be HIGH.
	for i := 0; i < 4; i++ {
		k.notGatesForAndGate1[i].Update(k.mainBus.GetOutputWire(i))
	}
	k.andGate1.Update(
		k.notGatesForAndGate1[0].Output(),
		k.notGatesForAndGate1[1].Output(),
		k.notGatesForAndGate1[2].Output(),
		k.notGatesForAndGate1[3].Output(),
		k.mainBus.GetOutputWire(12),
		k.mainBus.GetOutputWire(13),
		k.mainBus.GetOutputWire(14),
		k.mainBus.GetOutputWire(15),
	)

	// Step 2: andGate2 = SET + DATA_OR_ADDRESS + MODE (output-address phase).
	k.andGate2.Update(
		k.ioBus.GetOutputWire(components.CLOCK_SET),
		k.ioBus.GetOutputWire(components.DATA_OR_ADDRESS),
		k.ioBus.GetOutputWire(components.MODE),
	)

	// Step 3: latch "this device is selected" while address+set phase is active.
	k.memoryBit.Update(k.andGate1.Output(), k.andGate2.Output())

	// Step 4: andGate3 = ENABLE + NOT(DATA_OR_ADDRESS) + NOT(MODE) (input-data phase).
	k.notGatesForAndGate3[0].Update(k.ioBus.GetOutputWire(components.DATA_OR_ADDRESS))
	k.notGatesForAndGate3[1].Update(k.ioBus.GetOutputWire(components.MODE))
	k.andGate3.Update(
		k.ioBus.GetOutputWire(components.CLOCK_ENABLE),
		k.notGatesForAndGate3[0].Output(),
		k.notGatesForAndGate3[1].Output(),
	)

	// Step 5: fire only when selected AND in enable-input phase.
	k.andGate4.Update(k.memoryBit.Get(), k.andGate3.Output())

	// Step 6: deliver keycode to mainBus.
	if k.andGate4.Output() {
		k.keycodeRegister.Set()
		k.keycodeRegister.Enable()
		k.keycodeRegister.Update()
		k.KeyboardInBus.SetValue(0)
		k.keycodeRegister.Unset()
	} else {
		k.keycodeRegister.Disable()
		k.keycodeRegister.Update()
	}
}

// KeyPress represents a single key event from the OS.
type KeyPress struct {
	Value  int
	IsDown bool
}

// Keyboard reads from keyPressChannel and writes keycodes to its output bus at ~30fps.
type Keyboard struct {
	outBus          *components.Bus
	keyPressChannel chan *KeyPress
	quit            chan bool
}

func NewKeyboard(keyPressChannel chan *KeyPress, quit chan bool) *Keyboard {
	return &Keyboard{
		keyPressChannel: keyPressChannel,
		quit:            quit,
	}
}

func (k *Keyboard) ConnectTo(bus *components.Bus) { k.outBus = bus }

func (k *Keyboard) Run() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-k.quit:
			return
		case <-ticker.C:
			select {
			case key := <-k.keyPressChannel:
				if key.IsDown {
					k.outBus.SetValue(uint16(key.Value))
				}
			default:
			}
		}
	}
}
