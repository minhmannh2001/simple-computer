package components

import "simple-computer/circuit"

// --- Compare2 ---
// Single-bit comparator stage. Designed to chain from MSB to LSB.
//
// equalOut   = equalIn AND NOT(XOR(a, b))
// isLargerOut = (equalIn AND a AND XOR(a, b)) OR isLargerIn
//
// Feeding this chain MSB-first means the first bit that differs determines the
// result, and all lower-bit stages simply carry the decision forward unchanged.

type Compare2 struct {
	inputA, inputB         circuit.Wire
	xor1                   circuit.XORGate
	not1                   circuit.NOTGate
	and1                   circuit.ANDGate
	andgate3               ANDGate3
	or1                    circuit.ORGate
	out                    circuit.Wire
	equalIn, equalOut      circuit.Wire
	isLargerIn, isLargerOut circuit.Wire
}

func NewCompare2() *Compare2 { return &Compare2{} }

func (c *Compare2) Update(inputA, inputB, equalIn, isLargerIn bool) {
	c.inputA.Update(inputA)
	c.inputB.Update(inputB)
	c.equalIn.Update(equalIn)
	c.isLargerIn.Update(isLargerIn)

	c.xor1.Update(inputA, inputB)
	c.not1.Update(c.xor1.Output())
	c.and1.Update(equalIn, c.not1.Output())
	c.andgate3.Update(equalIn, inputA, c.xor1.Output())
	c.or1.Update(c.andgate3.Output(), isLargerIn)

	c.out.Update(c.xor1.Output())
	c.equalOut.Update(c.and1.Output())
	c.isLargerOut.Update(c.or1.Output())
}

func (c *Compare2) Equal() bool  { return c.equalOut.Get() }
func (c *Compare2) Larger() bool { return c.isLargerOut.Get() }
func (c *Compare2) Output() bool { return c.out.Get() }

// --- Comparator ---
// 16-bit comparator: chains 16 Compare2 stages from MSB (index 0) to LSB (index 15).
// inputs 0–15 = operand A, inputs 16–31 = operand B.
// Seed: equalIn=true, aIsLargerIn=false.

type Comparator struct {
	inputs       [BUS_WIDTH * 2]circuit.Wire
	equalIn      circuit.Wire
	aIsLargerIn  circuit.Wire
	compares     [BUS_WIDTH]Compare2
	outputs      [BUS_WIDTH]circuit.Wire
	equalOut     circuit.Wire
	aIsLargerOut circuit.Wire
	next         Component
}

func NewComparator() *Comparator { return &Comparator{} }

func (c *Comparator) ConnectOutput(b Component)          { c.next = b }
func (c *Comparator) SetInputWire(index int, value bool) { c.inputs[index].Update(value) }
func (c *Comparator) GetOutputWire(index int) bool       { return c.outputs[index].Get() }
func (c *Comparator) Equal() bool                        { return c.equalOut.Get() }
func (c *Comparator) Larger() bool                       { return c.aIsLargerOut.Get() }

func (c *Comparator) Update() {
	eqIn := true
	lgIn := false

	for i := range BUS_WIDTH {
		a := c.inputs[i].Get()
		b := c.inputs[i+BUS_WIDTH].Get()
		c.compares[i].Update(a, b, eqIn, lgIn)
		c.outputs[i].Update(c.compares[i].Output())
		eqIn = c.compares[i].Equal()
		lgIn = c.compares[i].Larger()
	}

	c.equalOut.Update(eqIn)
	c.aIsLargerOut.Update(lgIn)

	if c.next != nil {
		for i := range BUS_WIDTH {
			c.next.SetInputWire(i, c.outputs[i].Get())
		}
	}
}

// --- BusOne ---
// Reads from inputBus and writes to outputBus.
//
// When disabled (bus1=false):
//   outputs[i] = inputs[i]  (pass-through for all 16 bits)
//
// When enabled (bus1=true):
//   outputs[0..14] = inputs[i] AND NOT(bus1) = 0
//   outputs[15]    = inputs[15] OR bus1       = 1
//   Net: constant value 0x0001 regardless of input.
//
// The same AND/OR formula handles both modes; no branching needed.

type BusOne struct {
	inputBus  *Bus
	outputBus *Bus
	inputs    [BUS_WIDTH]circuit.Wire
	bus1      circuit.Wire
	andGates  [BUS_WIDTH - 1]circuit.ANDGate
	notGate   circuit.NOTGate
	orGate    circuit.ORGate
	outputs   [BUS_WIDTH]circuit.Wire
	next      Component
}

func NewBusOne(inputBus, outputBus *Bus) *BusOne {
	return &BusOne{inputBus: inputBus, outputBus: outputBus}
}

func (b *BusOne) Enable()  { b.bus1.Update(true) }
func (b *BusOne) Disable() { b.bus1.Update(false) }

func (b *BusOne) ConnectOutput(c Component)          { b.next = c }
func (b *BusOne) SetInputWire(index int, value bool) { b.inputs[index].Update(value) }
func (b *BusOne) GetOutputWire(index int) bool       { return b.outputs[index].Get() }

func (b *BusOne) Update() {
	for i := range BUS_WIDTH {
		b.inputs[i].Update(b.inputBus.GetOutputWire(i))
	}

	b.notGate.Update(b.bus1.Get())

	for i := range BUS_WIDTH - 1 {
		b.andGates[i].Update(b.inputs[i].Get(), b.notGate.Output())
		b.outputs[i].Update(b.andGates[i].Output())
	}

	b.orGate.Update(b.inputs[BUS_WIDTH-1].Get(), b.bus1.Get())
	b.outputs[BUS_WIDTH-1].Update(b.orGate.Output())

	for i := range BUS_WIDTH {
		b.outputBus.SetInputWire(i, b.outputs[i].Get())
	}

	if b.next != nil {
		for i := range BUS_WIDTH {
			b.next.SetInputWire(i, b.outputs[i].Get())
		}
	}
}

func (b *BusOne) String() string { return b.outputBus.String() }
