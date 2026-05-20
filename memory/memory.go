package memory

import (
	"simple-computer/circuit"
	"simple-computer/components"
)

// Cell is a single memory cell backed by a Register.
// Three AND gates are used to gate the set and enable signals in hardware style.
type Cell struct {
	value components.Register
	gates [3]circuit.ANDGate
}

func NewCell(inputBus, outputBus *components.Bus) *Cell {
	c := &Cell{}
	r := components.NewRegister("", inputBus, outputBus)
	c.value = *r
	return c
}

func (c *Cell) Update(set, enable bool) {
	c.gates[0].Update(true, true)
	c.gates[1].Update(c.gates[0].Output(), set)
	c.gates[2].Update(c.gates[0].Output(), enable)
	if c.gates[1].Output() {
		c.value.Set()
	} else {
		c.value.Unset()
	}
	if c.gates[2].Output() {
		c.value.Enable()
	} else {
		c.value.Disable()
	}
	c.value.Update()
}

// Memory64K is a 256×256 grid of Cells (64K 16-bit words) with a Memory Address Register.
// AddressRegister is exported so the CPU can load it directly via the main bus.
type Memory64K struct {
	AddressRegister components.Register
	rowDecoder      components.Decoder8x256
	colDecoder      components.Decoder8x256
	data            [256][256]Cell
	set             circuit.Wire
	enable          circuit.Wire
	bus             *components.Bus
}

func NewMemory64K(bus *components.Bus) *Memory64K {
	m := &Memory64K{bus: bus}
	r := components.NewRegister("MAR", bus, bus)
	m.AddressRegister = *r
	for row := range 256 {
		for col := range 256 {
			c := NewCell(bus, bus)
			m.data[row][col] = *c
		}
	}
	return m
}

func (m *Memory64K) Enable()  { m.enable.Update(true) }
func (m *Memory64K) Disable() { m.enable.Update(false) }
func (m *Memory64K) Set()     { m.set.Update(true) }
func (m *Memory64K) Unset()   { m.set.Update(false) }

// Update latches the MAR from the bus, decodes the address, then reads or writes the selected cell.
// Bits 0–7 of the MAR (high byte) select the row; bits 8–15 (low byte) select the column.
func (m *Memory64K) Update() {
	m.AddressRegister.Update()
	m.rowDecoder.Update(
		m.AddressRegister.Bit(0), m.AddressRegister.Bit(1),
		m.AddressRegister.Bit(2), m.AddressRegister.Bit(3),
		m.AddressRegister.Bit(4), m.AddressRegister.Bit(5),
		m.AddressRegister.Bit(6), m.AddressRegister.Bit(7),
	)
	m.colDecoder.Update(
		m.AddressRegister.Bit(8),  m.AddressRegister.Bit(9),
		m.AddressRegister.Bit(10), m.AddressRegister.Bit(11),
		m.AddressRegister.Bit(12), m.AddressRegister.Bit(13),
		m.AddressRegister.Bit(14), m.AddressRegister.Bit(15),
	)
	row := m.rowDecoder.Index()
	col := m.colDecoder.Index()
	m.data[row][col].Update(m.set.Get(), m.enable.Get())
}

func (m *Memory64K) String() string {
	return "MAR: " + m.AddressRegister.String()
}
