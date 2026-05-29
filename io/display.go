package io

import (
	"simple-computer/circuit"
	"simple-computer/components"
	"time"
)

// displayRAM is the frame buffer for the display, separate from main memory.
// It has two address registers so the CPU (InputAddressRegister) and ScreenControl
// (OutputAddressRegister) can operate concurrently without corrupting each other's pointer.
type displayRAM struct {
	InputAddressRegister  components.Register
	OutputAddressRegister components.Register
	cells                 [4800]uint16 // 30 bytes × 160 rows = 4800 16-bit words
	mainBus               *components.Bus
	screenBus             *components.Bus
	set                   circuit.Wire
	enable                circuit.Wire
}

func newDisplayRAM(mainBus, screenBus *components.Bus) *displayRAM {
	scanBus := components.NewBus(components.BUS_WIDTH)
	d := &displayRAM{
		mainBus:   mainBus,
		screenBus: screenBus,
	}
	d.InputAddressRegister = *components.NewRegister("inputMAR", mainBus, mainBus)
	d.OutputAddressRegister = *components.NewRegister("outputMAR", scanBus, screenBus)
	return d
}

func (d *displayRAM) Set()     { d.set.Update(true) }
func (d *displayRAM) Unset()   { d.set.Update(false) }
func (d *displayRAM) Enable()  { d.enable.Update(true) }
func (d *displayRAM) Disable() { d.enable.Update(false) }

// UpdateIncoming writes the current mainBus value to cells[InputAddressRegister.Value()]
// when set. The InputAddressRegister must already hold the target address.
func (d *displayRAM) UpdateIncoming() {
	if !d.set.Get() {
		return
	}
	addr := d.InputAddressRegister.Value()
	var val uint16
	for i := range components.BUS_WIDTH {
		if d.mainBus.GetOutputWire(i) {
			val |= 1 << (components.BUS_WIDTH - 1 - uint(i))
		}
	}
	if int(addr) < len(d.cells) {
		d.cells[addr] = val
	}
}

// UpdateOutgoing puts cells[OutputAddressRegister.Value()] onto screenBus when enabled.
func (d *displayRAM) UpdateOutgoing() {
	if !d.enable.Get() {
		return
	}
	addr := d.OutputAddressRegister.Value()
	if int(addr) < len(d.cells) {
		d.screenBus.SetValue(d.cells[addr])
	}
}

// DisplayAdapter is the I/O peripheral for the display, at device address 0x0007.
// The CPU writes to it via two successive OUT instructions: first the target display RAM
// address, then the pixel data. A writeToRAM latch alternates between the two modes.
type DisplayAdapter struct {
	ioBus      *components.IOBus
	mainBus    *components.Bus
	screenBus  *components.Bus
	displayRAM *displayRAM

	displayAdapterActiveBit *components.Bit

	addressSelectAndGate  components.ANDGate8
	addressSelectNOTGates [5]circuit.NOTGate

	// isAddressOutputModeGate fires when IOBus = SET + ADDRESS + OUTPUT,
	// used as the set-enable for displayAdapterActiveBit.
	isAddressOutputModeGate components.ANDGate3

	inputMARSetGate     components.ANDGate5
	inputMarSetNOTGates [2]circuit.NOTGate
	writeToRAM          *components.Bit
	writeToRAMToggleGate circuit.NOTGate

	displayRAMSetGate components.ANDGate5
}

func NewDisplayAdapter() *DisplayAdapter {
	return &DisplayAdapter{
		displayAdapterActiveBit: components.NewBit(),
		writeToRAM:              components.NewBit(),
	}
}

func (d *DisplayAdapter) Connect(ioBus *components.IOBus, mainBus *components.Bus) {
	d.ioBus = ioBus
	d.mainBus = mainBus
	d.screenBus = components.NewBus(components.BUS_WIDTH)
	d.displayRAM = newDisplayRAM(mainBus, d.screenBus)
}

func (d *DisplayAdapter) Update() {
	// Detect address 0x0007: wires 8-12 (bits 7-3) must be LOW, wires 13-15 (bits 2-0) HIGH.
	for i, wi := range [5]int{8, 9, 10, 11, 12} {
		d.addressSelectNOTGates[i].Update(d.mainBus.GetOutputWire(wi))
	}
	d.addressSelectAndGate.Update(
		d.addressSelectNOTGates[0].Output(),
		d.addressSelectNOTGates[1].Output(),
		d.addressSelectNOTGates[2].Output(),
		d.addressSelectNOTGates[3].Output(),
		d.addressSelectNOTGates[4].Output(),
		d.mainBus.GetOutputWire(13),
		d.mainBus.GetOutputWire(14),
		d.mainBus.GetOutputWire(15),
	)

	// isAddressOutputModeGate: SET AND ADDRESS AND OUTPUT — the activation condition.
	d.isAddressOutputModeGate.Update(
		d.ioBus.GetOutputWire(components.CLOCK_SET),
		d.ioBus.GetOutputWire(components.DATA_OR_ADDRESS),
		d.ioBus.GetOutputWire(components.MODE),
	)

	// Latch or reset device selection on each activation-phase cycle.
	d.displayAdapterActiveBit.Update(
		d.addressSelectAndGate.Output(),
		d.isAddressOutputModeGate.Output(),
	)

	// NOT gates shared by both write gates.
	d.inputMarSetNOTGates[0].Update(d.ioBus.GetOutputWire(components.DATA_OR_ADDRESS)) // NOT(ADDRESS)
	d.inputMarSetNOTGates[1].Update(d.writeToRAM.Get())                                // NOT(writeToRAM)

	// inputMARSetGate fires when: selected AND SET AND OUTPUT AND DATA AND NOT(writeToRAM).
	d.inputMARSetGate.Update(
		d.displayAdapterActiveBit.Get(),
		d.ioBus.GetOutputWire(components.CLOCK_SET),
		d.ioBus.GetOutputWire(components.MODE),
		d.inputMarSetNOTGates[0].Output(),
		d.inputMarSetNOTGates[1].Output(),
	)

	// displayRAMSetGate fires when: selected AND SET AND OUTPUT AND DATA AND writeToRAM.
	d.displayRAMSetGate.Update(
		d.displayAdapterActiveBit.Get(),
		d.ioBus.GetOutputWire(components.CLOCK_SET),
		d.ioBus.GetOutputWire(components.MODE),
		d.inputMarSetNOTGates[0].Output(),
		d.writeToRAM.Get(),
	)

	if d.inputMARSetGate.Output() {
		d.displayRAM.InputAddressRegister.Set()
		d.displayRAM.InputAddressRegister.Update()
		d.displayRAM.InputAddressRegister.Unset()
		d.toggleWriteToRAM()
	}

	if d.displayRAMSetGate.Output() {
		d.displayRAM.Set()
		d.displayRAM.UpdateIncoming()
		d.displayRAM.Unset()
		d.toggleWriteToRAM()
	}
}

func (d *DisplayAdapter) toggleWriteToRAM() {
	d.writeToRAMToggleGate.Update(d.writeToRAM.Get())
	d.writeToRAM.Update(d.writeToRAMToggleGate.Output(), true)
}

// ScreenControl scans the display RAM at ~30fps and pushes frame snapshots to outputChan.
type ScreenControl struct {
	adapter    *DisplayAdapter
	outputChan chan *[160][240]byte
	clock      <-chan time.Time
	quit       chan bool
	output     [160][240]byte
}

func NewScreenControl(adapter *DisplayAdapter, outputChan chan *[160][240]byte, quit chan bool) *ScreenControl {
	return &ScreenControl{
		adapter:    adapter,
		outputChan: outputChan,
		quit:       quit,
	}
}

func (s *ScreenControl) Run() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()
	s.clock = ticker.C
	for {
		select {
		case <-s.quit:
			return
		case <-s.clock:
			s.Update()
			frame := s.output
			select {
			case s.outputChan <- &frame:
			default:
			}
		}
	}
}

// Update scans all 4800 display RAM cells (30 bytes × 160 rows) and populates the output
// array. Each cell holds 8 pixels in bits 15-8 (the high byte); bit 15 = leftmost pixel.
func (s *ScreenControl) Update() {
	const widthInBytes = 30
	for y := range 160 {
		for xByte := range widthInBytes {
			cell := s.adapter.displayRAM.cells[y*widthInBytes+xByte]
			for bit := range 8 {
				s.output[y][xByte*8+bit] = byte((cell >> (15 - uint(bit))) & 1)
			}
		}
	}
}
