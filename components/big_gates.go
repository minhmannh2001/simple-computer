package components

import "simple-computer/circuit"

// ANDGate3 chains two 2-input AND gates: output is true only when all 3 inputs are true.
type ANDGate3 struct {
	and1, and2 circuit.ANDGate
}

func NewANDGate3() *ANDGate3 { return &ANDGate3{} }

func (g *ANDGate3) Update(a, b, c bool) {
	g.and1.Update(a, b)
	g.and2.Update(g.and1.Output(), c)
}

func (g *ANDGate3) Output() bool { return g.and2.Output() }

// ANDGate4 chains three 2-input AND gates.
type ANDGate4 struct {
	and1, and2, and3 circuit.ANDGate
}

func NewANDGate4() *ANDGate4 { return &ANDGate4{} }

func (g *ANDGate4) Update(a, b, c, d bool) {
	g.and1.Update(a, b)
	g.and2.Update(c, d)
	g.and3.Update(g.and1.Output(), g.and2.Output())
}

func (g *ANDGate4) Output() bool { return g.and3.Output() }

// ANDGate5 chains four 2-input AND gates.
type ANDGate5 struct {
	and1, and2, and3, and4 circuit.ANDGate
}

func NewANDGate5() *ANDGate5 { return &ANDGate5{} }

func (g *ANDGate5) Update(a, b, c, d, e bool) {
	g.and1.Update(a, b)
	g.and2.Update(c, d)
	g.and3.Update(g.and1.Output(), g.and2.Output())
	g.and4.Update(g.and3.Output(), e)
}

func (g *ANDGate5) Output() bool { return g.and4.Output() }

// ANDGate8 chains seven 2-input AND gates.
type ANDGate8 struct {
	and1, and2, and3, and4, and5, and6, and7 circuit.ANDGate
}

func NewANDGate8() *ANDGate8 { return &ANDGate8{} }

func (g *ANDGate8) Update(a, b, c, d, e, f, h, i bool) {
	g.and1.Update(a, b)
	g.and2.Update(c, d)
	g.and3.Update(e, f)
	g.and4.Update(h, i)
	g.and5.Update(g.and1.Output(), g.and2.Output())
	g.and6.Update(g.and3.Output(), g.and4.Output())
	g.and7.Update(g.and5.Output(), g.and6.Output())
}

func (g *ANDGate8) Output() bool { return g.and7.Output() }

// ORGate3 chains two 2-input OR gates.
type ORGate3 struct {
	or1, or2 circuit.ORGate
}

func NewORGate3() *ORGate3 { return &ORGate3{} }

func (g *ORGate3) Update(a, b, c bool) {
	g.or1.Update(a, b)
	g.or2.Update(g.or1.Output(), c)
}

func (g *ORGate3) Output() bool { return g.or2.Output() }

// ORGate4 chains three 2-input OR gates.
type ORGate4 struct {
	or1, or2, or3 circuit.ORGate
}

func NewORGate4() *ORGate4 { return &ORGate4{} }

func (g *ORGate4) Update(a, b, c, d bool) {
	g.or1.Update(a, b)
	g.or2.Update(c, d)
	g.or3.Update(g.or1.Output(), g.or2.Output())
}

func (g *ORGate4) Output() bool { return g.or3.Output() }

// ORGate5 chains four 2-input OR gates.
type ORGate5 struct {
	or1, or2, or3, or4 circuit.ORGate
}

func NewORGate5() *ORGate5 { return &ORGate5{} }

func (g *ORGate5) Update(a, b, c, d, e bool) {
	g.or1.Update(a, b)
	g.or2.Update(c, d)
	g.or3.Update(g.or1.Output(), g.or2.Output())
	g.or4.Update(g.or3.Output(), e)
}

func (g *ORGate5) Output() bool { return g.or4.Output() }

// ORGate6 chains five 2-input OR gates.
type ORGate6 struct {
	or1, or2, or3, or4, or5 circuit.ORGate
}

func NewORGate6() *ORGate6 { return &ORGate6{} }

func (g *ORGate6) Update(a, b, c, d, e, f bool) {
	g.or1.Update(a, b)
	g.or2.Update(c, d)
	g.or3.Update(e, f)
	g.or4.Update(g.or1.Output(), g.or2.Output())
	g.or5.Update(g.or4.Output(), g.or3.Output())
}

func (g *ORGate6) Output() bool { return g.or5.Output() }
