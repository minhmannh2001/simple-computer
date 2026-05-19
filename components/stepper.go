package components

import (
	"simple-computer/circuit"
	"strings"
)

type Stepper struct {
	bits           [12]Bit
	reset          circuit.Wire
	resetNotGate   circuit.NOTGate
	clockIn        circuit.Wire
	clockInNotGate circuit.NOTGate
	inputOrGates   [2]circuit.ORGate
	outputs        [7]circuit.Wire
	outputAndGates [5]circuit.ANDGate
	outputOrGate   circuit.ORGate
	outputNotGates [6]circuit.NOTGate
}

func NewStepper() *Stepper {
	s := &Stepper{}
	// Bit value-types need gates[3] primed (same as NewBit) to be stable in hold mode
	for i := range 12 {
		s.bits[i].gates[3].Update(false, false)
	}
	// clockIn=false, reset=false → bits[0] latches true via resetNotGate, establishing step 0
	s.step()
	return s
}

func (s *Stepper) step() {
	s.clockInNotGate.Update(s.clockIn.Get())
	s.resetNotGate.Update(s.reset.Get())

	// OR(reset, NOT(clock)): high when clock LOW or resetting
	s.inputOrGates[0].Update(s.reset.Get(), s.clockInNotGate.Output())
	// OR(reset, clock): high when clock HIGH or resetting
	s.inputOrGates[1].Update(s.reset.Get(), s.clockIn.Get())

	// Pair 0: master input is NOT(reset) — true normally, false during reset (clears the ring)
	s.bits[0].Update(s.resetNotGate.Output(), s.inputOrGates[0].Output())
	s.bits[1].Update(s.bits[0].Get(), s.inputOrGates[1].Output())
	// Pairs 1–5: master input is previous pair's slave
	for i := 1; i < 6; i++ {
		s.bits[2*i].Update(s.bits[2*i-1].Get(), s.inputOrGates[0].Output())
		s.bits[2*i+1].Update(s.bits[2*i].Get(), s.inputOrGates[1].Output())
	}

	// NOT(slave[i]) — used to detect "next step not yet active"
	for i := range 6 {
		s.outputNotGates[i].Update(s.bits[2*i+1].Get())
	}

	// Step 0: OR(reset, NOT(slave[0])) — default active when nothing else is, or on reset
	s.outputOrGate.Update(s.reset.Get(), s.outputNotGates[0].Output())
	s.outputs[0].Update(s.outputOrGate.Output())

	// Steps 1–5: AND(slave[i], NOT(slave[i+1]))
	for i := range 5 {
		s.outputAndGates[i].Update(s.bits[2*i+1].Get(), s.outputNotGates[i+1].Output())
		s.outputs[i+1].Update(s.outputAndGates[i].Output())
	}

	// Sentinel: fires when token reaches last slave, triggering immediate reset
	s.outputs[6].Update(s.bits[11].Get())
}

func (s *Stepper) Update(clockIn bool) {
	s.clockIn.Update(clockIn)
	s.reset.Update(s.outputs[6].Get())
	s.step()
	if s.outputs[6].Get() {
		s.reset.Update(true)
		s.step()
	}
}

func (s *Stepper) GetOutputWire(index int) bool {
	return s.outputs[index].Get()
}

func (s *Stepper) String() string {
	parts := make([]string, 6)
	for i := range 6 {
		if s.outputs[i].Get() {
			parts[i] = "*"
		} else {
			parts[i] = "-"
		}
	}
	return strings.Join(parts, " ")
}
