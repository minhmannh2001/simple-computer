package components

import "simple-computer/circuit"

const BUS_WIDTH = 16

type Component interface {
	ConnectOutput(Component)
	SetInputWire(int, bool)
	GetOutputWire(int) bool
}

type Bit struct {
	gates [4]circuit.NANDGate
	wireO circuit.Wire
}

func NewBit() *Bit {
	b := &Bit{}
	// gates[3] must start true for the hold-false state to be stable;
	// all-zero gate outputs is not a valid latch configuration.
	b.gates[3].Update(false, false)
	return b
}

func (b *Bit) Get() bool { return b.wireO.Get() }

func (b *Bit) Update(wireI bool, wireS bool) {
	for range 2 {
		b.gates[0].Update(wireI, wireS)
		b.gates[1].Update(b.gates[0].Output(), wireS)
		b.gates[2].Update(b.gates[0].Output(), b.gates[3].Output())
		b.gates[3].Update(b.gates[2].Output(), b.gates[1].Output())
		b.wireO.Update(b.gates[2].Output())
	}
}

type Word struct {
	inputs  [BUS_WIDTH]circuit.Wire
	bits    [BUS_WIDTH]Bit
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewWord() *Word { return &Word{} }

func (w *Word) ConnectOutput(b Component) { w.next = b }

func (w *Word) SetInputWire(index int, value bool) { w.inputs[index].Update(value) }

func (w *Word) GetOutputWire(index int) bool { return w.outputs[index].Get() }

func (w *Word) Update(set bool) {
	for i := range BUS_WIDTH {
		w.bits[i].Update(w.inputs[i].Get(), set)
		w.outputs[i].Update(w.bits[i].Get())
	}
	if w.next != nil {
		for i := range BUS_WIDTH {
			w.next.SetInputWire(i, w.outputs[i].Get())
		}
	}
}
