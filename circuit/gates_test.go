package circuit

import "testing"

func testTwoInputGate(t *testing.T, name string, update func(a, b bool), output func() bool, cases [][3]bool) {
	t.Helper()
	for _, c := range cases {
		update(c[0], c[1])
		if output() != c[2] {
			t.Errorf("%s(%v,%v): got %v, want %v", name, c[0], c[1], output(), c[2])
		}
	}
}

func TestNANDGate(t *testing.T) {
	g := NewNANDGate()
	testTwoInputGate(t, "NAND", g.Update, g.Output, [][3]bool{
		{false, false, true},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	})
}

func TestANDGate(t *testing.T) {
	g := NewANDGate()
	testTwoInputGate(t, "AND", g.Update, g.Output, [][3]bool{
		{false, false, false},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	})
}

func TestORGate(t *testing.T) {
	g := NewORGate()
	testTwoInputGate(t, "OR", g.Update, g.Output, [][3]bool{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, true},
	})
}

func TestXORGate(t *testing.T) {
	g := NewXORGate()
	testTwoInputGate(t, "XOR", g.Update, g.Output, [][3]bool{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	})
}

func TestNORGate(t *testing.T) {
	g := NewNORGate()
	testTwoInputGate(t, "NOR", g.Update, g.Output, [][3]bool{
		{false, false, true},
		{true, false, false},
		{false, true, false},
		{true, true, false},
	})
}

func TestNOTGate(t *testing.T) {
	g := NewNOTGate()
	g.Update(false)
	if g.Output() != true {
		t.Errorf("NOT(false): got false, want true")
	}
	g.Update(true)
	if g.Output() != false {
		t.Errorf("NOT(true): got true, want false")
	}
}
