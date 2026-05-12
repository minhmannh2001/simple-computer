package components

import "simple-computer/circuit"

// --- Enabler ---

type Enabler struct {
	inputs  [BUS_WIDTH]circuit.Wire
	gates   [BUS_WIDTH]circuit.ANDGate
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewEnabler() *Enabler { return &Enabler{} }

func (e *Enabler) ConnectOutput(b Component) { e.next = b }
func (e *Enabler) SetInputWire(index int, value bool) { e.inputs[index].Update(value) }
func (e *Enabler) GetOutputWire(index int) bool       { return e.outputs[index].Get() }

func (e *Enabler) Update(enable bool) {
	for i := range BUS_WIDTH {
		e.gates[i].Update(e.inputs[i].Get(), enable)
		e.outputs[i].Update(e.gates[i].Output())
	}
	if e.next != nil {
		for i := range BUS_WIDTH {
			e.next.SetInputWire(i, e.outputs[i].Get())
		}
	}
}

// --- NOTer ---

type NOTer struct {
	inputs  [BUS_WIDTH]circuit.Wire
	gates   [BUS_WIDTH]circuit.NOTGate
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewNOTer() *NOTer { return &NOTer{} }

func (n *NOTer) ConnectOutput(b Component) { n.next = b }
func (n *NOTer) SetInputWire(index int, value bool) { n.inputs[index].Update(value) }
func (n *NOTer) GetOutputWire(index int) bool       { return n.outputs[index].Get() }

func (n *NOTer) Update() {
	for i := range BUS_WIDTH {
		n.gates[i].Update(n.inputs[i].Get())
		n.outputs[i].Update(n.gates[i].Output())
	}
	if n.next != nil {
		for i := range BUS_WIDTH {
			n.next.SetInputWire(i, n.outputs[i].Get())
		}
	}
}

// --- ANDer ---
// inputs 0–15 = operand A, inputs 16–31 = operand B

type ANDer struct {
	inputs  [BUS_WIDTH * 2]circuit.Wire
	gates   [BUS_WIDTH]circuit.ANDGate
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewANDer() *ANDer { return &ANDer{} }

func (a *ANDer) ConnectOutput(b Component) { a.next = b }
func (a *ANDer) SetInputWire(index int, value bool) { a.inputs[index].Update(value) }
func (a *ANDer) GetOutputWire(index int) bool       { return a.outputs[index].Get() }

func (a *ANDer) Update() {
	for i := range BUS_WIDTH {
		a.gates[i].Update(a.inputs[i].Get(), a.inputs[i+BUS_WIDTH].Get())
		a.outputs[i].Update(a.gates[i].Output())
	}
	if a.next != nil {
		for i := range BUS_WIDTH {
			a.next.SetInputWire(i, a.outputs[i].Get())
		}
	}
}

// --- ORer ---
// inputs 0–15 = operand A, inputs 16–31 = operand B

type ORer struct {
	inputs  [BUS_WIDTH * 2]circuit.Wire
	gates   [BUS_WIDTH]circuit.ORGate
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewORer() *ORer { return &ORer{} }

func (o *ORer) ConnectOutput(b Component) { o.next = b }
func (o *ORer) SetInputWire(index int, value bool) { o.inputs[index].Update(value) }
func (o *ORer) GetOutputWire(index int) bool       { return o.outputs[index].Get() }

func (o *ORer) Update() {
	for i := range BUS_WIDTH {
		o.gates[i].Update(o.inputs[i].Get(), o.inputs[i+BUS_WIDTH].Get())
		o.outputs[i].Update(o.gates[i].Output())
	}
	if o.next != nil {
		for i := range BUS_WIDTH {
			o.next.SetInputWire(i, o.outputs[i].Get())
		}
	}
}

// --- XORer ---
// inputs 0–15 = operand A, inputs 16–31 = operand B

type XORer struct {
	inputs  [BUS_WIDTH * 2]circuit.Wire
	gates   [BUS_WIDTH]circuit.XORGate
	outputs [BUS_WIDTH]circuit.Wire
	next    Component
}

func NewXORer() *XORer { return &XORer{} }

func (x *XORer) ConnectOutput(b Component) { x.next = b }
func (x *XORer) SetInputWire(index int, value bool) { x.inputs[index].Update(value) }
func (x *XORer) GetOutputWire(index int) bool       { return x.outputs[index].Get() }

func (x *XORer) Update() {
	for i := range BUS_WIDTH {
		x.gates[i].Update(x.inputs[i].Get(), x.inputs[i+BUS_WIDTH].Get())
		x.outputs[i].Update(x.gates[i].Output())
	}
	if x.next != nil {
		for i := range BUS_WIDTH {
			x.next.SetInputWire(i, x.outputs[i].Get())
		}
	}
}

// --- LeftShifter ---
// outputs[i] = inputs[i+1] for i=0..14; outputs[15] = shiftIn; shiftOut = inputs[0]

type LeftShifter struct {
	inputs   [BUS_WIDTH]circuit.Wire
	outputs  [BUS_WIDTH]circuit.Wire
	shiftIn  circuit.Wire
	shiftOut circuit.Wire
	next     Component
}

func NewLeftShifter() *LeftShifter { return &LeftShifter{} }

func (l *LeftShifter) ConnectOutput(b Component) { l.next = b }
func (l *LeftShifter) SetInputWire(index int, value bool) { l.inputs[index].Update(value) }
func (l *LeftShifter) GetOutputWire(index int) bool       { return l.outputs[index].Get() }
func (l *LeftShifter) ShiftOut() bool                     { return l.shiftOut.Get() }

func (l *LeftShifter) Update(shiftIn bool) {
	l.shiftOut.Update(l.inputs[0].Get())
	for i := range BUS_WIDTH - 1 {
		l.outputs[i].Update(l.inputs[i+1].Get())
	}
	l.outputs[BUS_WIDTH-1].Update(shiftIn)
	if l.next != nil {
		for i := range BUS_WIDTH {
			l.next.SetInputWire(i, l.outputs[i].Get())
		}
	}
}

// --- RightShifter ---
// outputs[0] = shiftIn; outputs[i] = inputs[i-1] for i=1..15; shiftOut = inputs[15]

type RightShifter struct {
	inputs   [BUS_WIDTH]circuit.Wire
	outputs  [BUS_WIDTH]circuit.Wire
	shiftIn  circuit.Wire
	shiftOut circuit.Wire
	next     Component
}

func NewRightShifter() *RightShifter { return &RightShifter{} }

func (r *RightShifter) ConnectOutput(b Component) { r.next = b }
func (r *RightShifter) SetInputWire(index int, value bool) { r.inputs[index].Update(value) }
func (r *RightShifter) GetOutputWire(index int) bool       { return r.outputs[index].Get() }
func (r *RightShifter) ShiftOut() bool                     { return r.shiftOut.Get() }

func (r *RightShifter) Update(shiftIn bool) {
	r.shiftOut.Update(r.inputs[BUS_WIDTH-1].Get())
	r.outputs[0].Update(shiftIn)
	for i := 1; i < BUS_WIDTH; i++ {
		r.outputs[i].Update(r.inputs[i-1].Get())
	}
	if r.next != nil {
		for i := range BUS_WIDTH {
			r.next.SetInputWire(i, r.outputs[i].Get())
		}
	}
}

// --- IsZero ---
// OR all 16 inputs together (using ORer with same value on both A and B sides),
// then NOT the result. True only when all inputs are false.

type IsZero struct {
	inputs  [BUS_WIDTH]circuit.Wire
	orer    ORer
	notGate circuit.NOTGate
	output  circuit.Wire
}

func NewIsZero() *IsZero { return &IsZero{} }

func (z *IsZero) SetInputWire(index int, value bool) { z.inputs[index].Update(value) }
func (z *IsZero) GetOutputWire(_ int) bool           { return z.output.Get() }

func (z *IsZero) Reset() {
	for i := range BUS_WIDTH {
		z.inputs[i].Update(false)
	}
}

func (z *IsZero) Update() {
	for i := range BUS_WIDTH {
		v := z.inputs[i].Get()
		z.orer.SetInputWire(i, v)
		z.orer.SetInputWire(i+BUS_WIDTH, v)
	}
	z.orer.Update()
	anyTrue := false
	for i := range BUS_WIDTH {
		if z.orer.GetOutputWire(i) {
			anyTrue = true
			break
		}
	}
	z.notGate.Update(anyTrue)
	z.output.Update(z.notGate.Output())
}
