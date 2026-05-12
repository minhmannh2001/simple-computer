package components

import (
	"simple-computer/circuit"
	"strings"
)

type Bus struct {
	wires []circuit.Wire
	width int
}

func NewBus(width int) *Bus {
	return &Bus{
		wires: make([]circuit.Wire, width),
		width: width,
	}
}

func (b *Bus) SetInputWire(index int, value bool) { b.wires[index].Update(value) }

func (b *Bus) GetOutputWire(index int) bool { return b.wires[index].Get() }

// SetValue decomposes a uint16 MSB-first into the wire array.
// Wire 0 = MSB (bit 15), wire 15 = LSB (bit 0).
func (b *Bus) SetValue(value uint16) {
	for i := b.width - 1; i >= 0; i-- {
		b.wires[i].Update(value&1 == 1)
		value >>= 1
	}
}

func (b *Bus) String() string {
	var sb strings.Builder
	for i := range b.width {
		if b.wires[i].Get() {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}

func (b *Bus) ConnectOutput(_ Component) {}
