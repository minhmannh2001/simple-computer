package components

import "simple-computer/circuit"

// --- Add2: 1-bit full adder ---

type Add2 struct {
	inputA, inputB, carryIn circuit.Wire
	xor1, xor2              circuit.XORGate
	and1, and2              circuit.ANDGate
	or1                     circuit.ORGate
	carryOut, sumOut        circuit.Wire
}

func NewAdd2() *Add2 { return &Add2{} }

func (a *Add2) Update(inputA, inputB, carryIn bool) {
	a.inputA.Update(inputA)
	a.inputB.Update(inputB)
	a.carryIn.Update(carryIn)

	a.xor1.Update(a.inputA.Get(), a.inputB.Get())
	a.xor2.Update(a.xor1.Output(), a.carryIn.Get())
	a.sumOut.Update(a.xor2.Output())

	a.and1.Update(a.carryIn.Get(), a.xor1.Output())
	a.and2.Update(a.inputA.Get(), a.inputB.Get())
	a.or1.Update(a.and1.Output(), a.and2.Output())
	a.carryOut.Update(a.or1.Output())
}

func (a *Add2) Sum() bool   { return a.sumOut.Get() }
func (a *Add2) Carry() bool { return a.carryOut.Get() }

// --- Adder: 16-bit ripple-carry adder ---

type Adder struct {
	inputs   [32]circuit.Wire
	carryIn  circuit.Wire
	adds     [16]Add2
	carryOut circuit.Wire
	outputs  [16]circuit.Wire
	next     Component
}

func NewAdder() *Adder { return &Adder{} }

func (a *Adder) ConnectOutput(b Component) { a.next = b }

func (a *Adder) SetInputWire(index int, value bool) { a.inputs[index].Update(value) }

func (a *Adder) GetOutputWire(index int) bool { return a.outputs[index].Get() }

func (a *Adder) Carry() bool { return a.carryOut.Get() }

func (a *Adder) Update(carryIn bool) {
	a.carryIn.Update(carryIn)
	carry := carryIn

	// A occupies inputs[0..15] (MSB at 0, LSB at 15)
	// B occupies inputs[16..31] (MSB at 16, LSB at 31)
	// Addition ripples from LSB (index 15) to MSB (index 0)
	awire := 15
	bwire := 31
	for i := 15; i >= 0; i-- {
		a.adds[i].Update(a.inputs[awire].Get(), a.inputs[bwire].Get(), carry)
		a.outputs[i].Update(a.adds[i].Sum())
		carry = a.adds[i].Carry()
		awire--
		bwire--
	}
	a.carryOut.Update(carry)

	if a.next != nil {
		for i := range BUS_WIDTH {
			a.next.SetInputWire(i, a.outputs[i].Get())
		}
	}
}
