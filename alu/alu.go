package alu

import (
	"fmt"

	"simple-computer/circuit"
	"simple-computer/components"
)

const (
	ADD = iota // 0
	SHR        // 1
	SHL        // 2
	NOT        // 3
	AND        // 4
	OR         // 5
	XOR        // 6
	CMP        // 7
)

type ALU struct {
	inputABus      *components.Bus
	inputBBus      *components.Bus
	outputBus      *components.Bus
	flagsOutputBus *components.Bus

	Op      [3]circuit.Wire // Op[0]=LSB, Op[2]=MSB
	CarryIn circuit.Wire

	carryOut  circuit.Wire
	aIsLarger circuit.Wire
	isEqual   circuit.Wire

	opDecoder   components.Decoder3x8
	comparator  components.Comparator
	xorer       components.XORer
	orer        components.ORer
	ander       components.ANDer
	notter      components.NOTer
	leftShifer  components.LeftShifter
	rightShifer components.RightShifter
	adder       components.Adder
	isZero      components.IsZero
	enablers    [7]components.Enabler
	andGates    [3]circuit.ANDGate
}

func NewALU(inputABus, inputBBus, outputBus, flagsOutputBus *components.Bus) *ALU {
	return &ALU{
		inputABus:      inputABus,
		inputBBus:      inputBBus,
		outputBus:      outputBus,
		flagsOutputBus: flagsOutputBus,
	}
}

// setWireOnComponent copies A bus to indices 0–15 and B bus to indices 16–31.
func (a *ALU) setWireOnComponent(b components.Component) {
	for i := range components.BUS_WIDTH {
		b.SetInputWire(i, a.inputABus.GetOutputWire(i))
		b.SetInputWire(i+components.BUS_WIDTH, a.inputBBus.GetOutputWire(i))
	}
}

// wireToEnabler copies 16 outputs from b into enablers[index] inputs.
func (a *ALU) wireToEnabler(b components.Component, index int) {
	for i := range components.BUS_WIDTH {
		a.enablers[index].SetInputWire(i, b.GetOutputWire(i))
	}
}

func (a *ALU) updateOpDecoder() {
	a.opDecoder.Update(a.Op[2].Get(), a.Op[1].Get(), a.Op[0].Get())
}

func (a *ALU) updateComparator() {
	a.setWireOnComponent(&a.comparator)
	a.comparator.Update()
	a.aIsLarger.Update(a.comparator.Larger())
	a.isEqual.Update(a.comparator.Equal())
}

func (a *ALU) Update() {
	a.updateOpDecoder()
	a.updateComparator()

	op := a.opDecoder.Index()

	switch op {
	case ADD:
		a.setWireOnComponent(&a.adder)
		a.adder.Update(a.CarryIn.Get())
		a.wireToEnabler(&a.adder, ADD)
	case SHR:
		for i := range components.BUS_WIDTH {
			a.rightShifer.SetInputWire(i, a.inputABus.GetOutputWire(i))
		}
		a.rightShifer.Update(false)
		a.wireToEnabler(&a.rightShifer, SHR)
	case SHL:
		for i := range components.BUS_WIDTH {
			a.leftShifer.SetInputWire(i, a.inputABus.GetOutputWire(i))
		}
		a.leftShifer.Update(false)
		a.wireToEnabler(&a.leftShifer, SHL)
	case NOT:
		for i := range components.BUS_WIDTH {
			a.notter.SetInputWire(i, a.inputABus.GetOutputWire(i))
		}
		a.notter.Update()
		a.wireToEnabler(&a.notter, NOT)
	case AND:
		a.setWireOnComponent(&a.ander)
		a.ander.Update()
		a.wireToEnabler(&a.ander, AND)
	case OR:
		a.setWireOnComponent(&a.orer)
		a.orer.Update()
		a.wireToEnabler(&a.orer, OR)
	case XOR:
		a.setWireOnComponent(&a.xorer)
		a.xorer.Update()
		a.wireToEnabler(&a.xorer, XOR)
	}

	// Carry gates: only pass carry for the active shift/add op.
	a.andGates[0].Update(a.adder.Carry(), a.opDecoder.GetOutputWire(ADD))
	a.andGates[1].Update(a.rightShifer.ShiftOut(), a.opDecoder.GetOutputWire(SHR))
	a.andGates[2].Update(a.leftShifer.ShiftOut(), a.opDecoder.GetOutputWire(SHL))
	a.carryOut.Update(a.andGates[0].Output() || a.andGates[1].Output() || a.andGates[2].Output())

	if op != CMP {
		a.enablers[op].Update(true)
		for i := range components.BUS_WIDTH {
			a.outputBus.SetInputWire(i, a.enablers[op].GetOutputWire(i))
			a.isZero.SetInputWire(i, a.enablers[op].GetOutputWire(i))
		}
	} else {
		// CMP has no result output; force zero flag to false.
		for i := range components.BUS_WIDTH {
			a.outputBus.SetInputWire(i, false)
			a.isZero.SetInputWire(i, true)
		}
	}

	a.isZero.Update()

	a.flagsOutputBus.SetInputWire(0, a.carryOut.Get())
	a.flagsOutputBus.SetInputWire(1, a.aIsLarger.Get())
	a.flagsOutputBus.SetInputWire(2, a.isEqual.Get())
	a.flagsOutputBus.SetInputWire(3, a.isZero.GetOutputWire(0))
}

func (a *ALU) String() string {
	return fmt.Sprintf("out=%s flags=%s", a.outputBus, a.flagsOutputBus)
}
