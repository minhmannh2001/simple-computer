package components

import "simple-computer/circuit"

// Decoder2x4: 2 inputs → exactly one of 4 outputs active.
// inputA=MSB, inputB=LSB. out[i] is true when inputs encode i in binary.

type Decoder2x4 struct {
	inputA, inputB circuit.Wire
	notGates       [2]circuit.NOTGate
	andGates       [4]circuit.ANDGate
	outputs        [4]circuit.Wire
}

func NewDecoder2x4() *Decoder2x4 { return &Decoder2x4{} }

func (d *Decoder2x4) GetOutputWire(index int) bool { return d.outputs[index].Get() }

func (d *Decoder2x4) Update(inputA, inputB bool) {
	d.inputA.Update(inputA)
	d.inputB.Update(inputB)
	d.notGates[0].Update(inputA)
	d.notGates[1].Update(inputB)
	na, nb := d.notGates[0].Output(), d.notGates[1].Output()

	d.andGates[0].Update(na, nb) // 00
	d.andGates[1].Update(na, inputB) // 01
	d.andGates[2].Update(inputA, nb) // 10
	d.andGates[3].Update(inputA, inputB) // 11

	for i := range 4 {
		d.outputs[i].Update(d.andGates[i].Output())
	}
}

// Decoder3x8: 3 inputs → exactly one of 8 outputs active.
// inputA=MSB, inputC=LSB.

type Decoder3x8 struct {
	inputA, inputB, inputC circuit.Wire
	notGates               [3]circuit.NOTGate
	andGates               [8]ANDGate3
	outputs                [8]circuit.Wire
}

func NewDecoder3x8() *Decoder3x8 { return &Decoder3x8{} }

func (d *Decoder3x8) GetOutputWire(index int) bool { return d.outputs[index].Get() }

func (d *Decoder3x8) Index() int {
	for i := range 8 {
		if d.outputs[i].Get() {
			return i
		}
	}
	return 0
}

func (d *Decoder3x8) Update(inputA, inputB, inputC bool) {
	d.inputA.Update(inputA)
	d.inputB.Update(inputB)
	d.inputC.Update(inputC)
	d.notGates[0].Update(inputA)
	d.notGates[1].Update(inputB)
	d.notGates[2].Update(inputC)
	na, nb, nc := d.notGates[0].Output(), d.notGates[1].Output(), d.notGates[2].Output()

	d.andGates[0].Update(na, nb, nc)       // 000
	d.andGates[1].Update(na, nb, inputC)   // 001
	d.andGates[2].Update(na, inputB, nc)   // 010
	d.andGates[3].Update(na, inputB, inputC) // 011
	d.andGates[4].Update(inputA, nb, nc)   // 100
	d.andGates[5].Update(inputA, nb, inputC) // 101
	d.andGates[6].Update(inputA, inputB, nc) // 110
	d.andGates[7].Update(inputA, inputB, inputC) // 111

	for i := range 8 {
		d.outputs[i].Update(d.andGates[i].Output())
	}
}

// Decoder4x16: 4 inputs → exactly one of 16 outputs active.
// inputA=MSB (8), inputB=4, inputC=2, inputD=LSB (1).

type Decoder4x16 struct {
	notGates [4]circuit.NOTGate
	andGates [16]ANDGate4
	outputs  [16]circuit.Wire
	index    int
}

func NewDecoder4x16() *Decoder4x16 { return &Decoder4x16{} }

func (d *Decoder4x16) GetOutputWire(index int) bool { return d.outputs[index].Get() }
func (d *Decoder4x16) Index() int                   { return d.index }

func (d *Decoder4x16) Update(inputA, inputB, inputC, inputD bool) {
	d.notGates[0].Update(inputA)
	d.notGates[1].Update(inputB)
	d.notGates[2].Update(inputC)
	d.notGates[3].Update(inputD)
	na, nb, nc, nd := d.notGates[0].Output(), d.notGates[1].Output(), d.notGates[2].Output(), d.notGates[3].Output()

	d.andGates[0].Update(na, nb, nc, nd)
	d.andGates[1].Update(na, nb, nc, inputD)
	d.andGates[2].Update(na, nb, inputC, nd)
	d.andGates[3].Update(na, nb, inputC, inputD)
	d.andGates[4].Update(na, inputB, nc, nd)
	d.andGates[5].Update(na, inputB, nc, inputD)
	d.andGates[6].Update(na, inputB, inputC, nd)
	d.andGates[7].Update(na, inputB, inputC, inputD)
	d.andGates[8].Update(inputA, nb, nc, nd)
	d.andGates[9].Update(inputA, nb, nc, inputD)
	d.andGates[10].Update(inputA, nb, inputC, nd)
	d.andGates[11].Update(inputA, nb, inputC, inputD)
	d.andGates[12].Update(inputA, inputB, nc, nd)
	d.andGates[13].Update(inputA, inputB, nc, inputD)
	d.andGates[14].Update(inputA, inputB, inputC, nd)
	d.andGates[15].Update(inputA, inputB, inputC, inputD)

	for i := range 16 {
		d.outputs[i].Update(d.andGates[i].Output())
		if d.outputs[i].Get() {
			d.index = i
		}
	}
}

// Decoder8x256: 8 inputs → selects one of 256 outputs.
// Built from 1 selector Decoder4x16 (high nibble: a,b,c,d)
// and 16 sub Decoder4x16s (low nibble: e,f,g,h).
// Only the sub-decoder selected by the high nibble runs its Update.

type Decoder8x256 struct {
	decoderSelector Decoder4x16
	decoders4x16    [16]Decoder4x16
	index           int
}

func NewDecoder8x256() *Decoder8x256 { return &Decoder8x256{} }

func (d *Decoder8x256) Index() int { return d.index }

func (d *Decoder8x256) Update(a, b, c, dd, e, f, g, h bool) {
	d.decoderSelector.Update(a, b, c, dd)
	sel := d.decoderSelector.Index()
	d.decoders4x16[sel].Update(e, f, g, h)
	d.index = sel*16 + d.decoders4x16[sel].Index()
}
