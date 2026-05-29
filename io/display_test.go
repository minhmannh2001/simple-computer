package io

import (
	"simple-computer/components"
	"testing"
)

func TestDisplayAdapter_AddressDetection(t *testing.T) {
	mainBus := components.NewBus(components.BUS_WIDTH)
	ioBus := components.NewIOBus()
	adapter := NewDisplayAdapter()
	adapter.Connect(ioBus, mainBus)

	mainBus.SetValue(0x0007)
	ioBus.Set()
	ioBus.Update(true, true) // MODE=output, DATA_OR_ADDRESS=address
	adapter.Update()
	if !adapter.displayAdapterActiveBit.Get() {
		t.Error("address 0x0007 should activate displayAdapterActiveBit")
	}

	// 0x0008 differs at wire 12 (bit 3=1), so NOT gate kills the AND; bit should reset.
	mainBus.SetValue(0x0008)
	adapter.Update()
	if adapter.displayAdapterActiveBit.Get() {
		t.Error("address 0x0008 should NOT activate displayAdapterActiveBit")
	}
}

func TestDisplayAdapter_TwoPhaseWrite(t *testing.T) {
	mainBus := components.NewBus(components.BUS_WIDTH)
	ioBus := components.NewIOBus()
	adapter := NewDisplayAdapter()
	adapter.Connect(ioBus, mainBus)

	// Step 1: activate adapter — send device address 0x0007.
	mainBus.SetValue(0x0007)
	ioBus.Set()
	ioBus.Update(true, true) // MODE=output, DATA_OR_ADDRESS=address
	adapter.Update()

	// Step 2: write display RAM address 0x0000 to InputMAR.
	mainBus.SetValue(0x0000)
	ioBus.Update(true, false) // MODE=output, DATA_OR_ADDRESS=data
	adapter.Update()
	if !adapter.writeToRAM.Get() {
		t.Error("after writing address, writeToRAM should be true")
	}

	// Step 3: write pixel data 0xFF00 to display RAM cell 0.
	mainBus.SetValue(0xFF00)
	adapter.Update() // IOBus still SET+OUTPUT+DATA
	if adapter.writeToRAM.Get() {
		t.Error("after writing data, writeToRAM should flip back to false")
	}
	if adapter.displayRAM.cells[0] != 0xFF00 {
		t.Errorf("display RAM[0] = 0x%04X, want 0xFF00", adapter.displayRAM.cells[0])
	}
}

func TestScreenControl_Update(t *testing.T) {
	mainBus := components.NewBus(components.BUS_WIDTH)
	ioBus := components.NewIOBus()
	adapter := NewDisplayAdapter()
	adapter.Connect(ioBus, mainBus)
	adapter.displayRAM.cells[0] = 0xFF00 // first 8 pixels all on, rest off

	outputChan := make(chan *[160][240]byte, 1)
	quit := make(chan bool)
	sc := NewScreenControl(adapter, outputChan, quit)
	sc.Update()

	for i := 0; i < 8; i++ {
		if sc.output[0][i] != 0x01 {
			t.Errorf("output[0][%d] = 0x%02X, want 0x01", i, sc.output[0][i])
		}
	}
	for i := 8; i < 16; i++ {
		if sc.output[0][i] != 0x00 {
			t.Errorf("output[0][%d] = 0x%02X, want 0x00", i, sc.output[0][i])
		}
	}
}
