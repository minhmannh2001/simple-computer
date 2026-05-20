package io

import (
	"simple-computer/components"
	"testing"
)

// TestKeyboardAdapter_DeliverKeycode is an integration test that simulates the two-phase
// CPU-keyboard handshake without involving the Keyboard goroutine.
//
// Phase 1: CPU outputs address 0x000F on mainBus with SET+ADDRESS+OUTPUT → memoryBit latches.
// Phase 2: CPU enables with ENABLE+DATA+INPUT → mainBus receives the keycode.
func TestKeyboardAdapter_DeliverKeycode(t *testing.T) {
	mainBus := components.NewBus(components.BUS_WIDTH)
	ioBus := components.NewIOBus()
	adapter := NewKeyboardAdapter()
	adapter.Connect(ioBus, mainBus)

	// Phase 1: CPU addresses the keyboard (0x000F) with SET + ADDRESS + OUTPUT.
	mainBus.SetValue(0x000F)
	ioBus.Set()
	ioBus.Update(true, true) // MODE=output, DATA_OR_ADDRESS=address
	adapter.Update()

	// Load a keycode onto the keyboard's input bus.
	adapter.KeyboardInBus.SetValue(0x0041) // 'A'

	// Phase 2: CPU reads from keyboard with ENABLE + DATA + INPUT.
	ioBus.Enable()
	ioBus.Update(false, false) // MODE=input, DATA_OR_ADDRESS=data
	adapter.Update()

	got := uint16(0)
	for i := 0; i < components.BUS_WIDTH; i++ {
		if mainBus.GetOutputWire(i) {
			got |= 1 << (components.BUS_WIDTH - 1 - uint(i))
		}
	}
	if got != 0x0041 {
		t.Errorf("mainBus = 0x%04X, want 0x0041", got)
	}
}
