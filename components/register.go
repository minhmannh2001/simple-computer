package components

import (
	"fmt"
	"simple-computer/circuit"
)

type Register struct {
	name      string
	set       circuit.Wire
	enable    circuit.Wire
	word      *Word
	enabler   *Enabler
	outputs   [BUS_WIDTH]circuit.Wire
	inputBus  *Bus
	outputBus *Bus
}

func NewRegister(name string, inputBus *Bus, outputBus *Bus) *Register {
	r := &Register{
		name:      name,
		inputBus:  inputBus,
		outputBus: outputBus,
		word:      NewWord(),
		enabler:   NewEnabler(),
	}
	r.word.ConnectOutput(r.enabler)
	return r
}

func (r *Register) Enable()  { r.enable.Update(true) }
func (r *Register) Disable() { r.enable.Update(false) }
func (r *Register) Set()     { r.set.Update(true) }
func (r *Register) Unset()   { r.set.Update(false) }

func (r *Register) Update() {
	for i := range BUS_WIDTH {
		r.word.SetInputWire(i, r.inputBus.GetOutputWire(i))
	}
	r.word.Update(r.set.Get())
	r.enabler.Update(r.enable.Get())
	for i := range BUS_WIDTH {
		r.outputs[i].Update(r.enabler.GetOutputWire(i))
	}
	if r.enable.Get() {
		for i := range BUS_WIDTH {
			r.outputBus.SetInputWire(i, r.outputs[i].Get())
		}
	}
}

func (r *Register) Bit(index int) bool { return r.word.GetOutputWire(index) }

func (r *Register) Value() uint16 {
	var result uint16
	for i := range BUS_WIDTH {
		if r.word.GetOutputWire(i) {
			result |= 1 << (BUS_WIDTH - 1 - uint(i))
		}
	}
	return result
}

func (r *Register) String() string {
	return fmt.Sprintf("%s: 0x%04X | enable=%v | set=%v", r.name, r.Value(), r.enable.Get(), r.set.Get())
}
