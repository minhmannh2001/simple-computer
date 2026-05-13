package components

import "testing"

func TestAdd2AllCombinations(t *testing.T) {
	cases := []struct {
		a, b, carryIn bool
		wantSum       bool
		wantCarry     bool
	}{
		{false, false, false, false, false},
		{false, false, true, true, false},
		{false, true, false, true, false},
		{false, true, true, false, true},
		{true, false, false, true, false},
		{true, false, true, false, true},
		{true, true, false, false, true},
		{true, true, true, true, true},
	}

	for _, c := range cases {
		add := NewAdd2()
		add.Update(c.a, c.b, c.carryIn)
		if got := add.Sum(); got != c.wantSum {
			t.Errorf("Add2(%v,%v,%v).Sum() = %v, want %v", c.a, c.b, c.carryIn, got, c.wantSum)
		}
		if got := add.Carry(); got != c.wantCarry {
			t.Errorf("Add2(%v,%v,%v).Carry() = %v, want %v", c.a, c.b, c.carryIn, got, c.wantCarry)
		}
	}
}

func setAdderInputs(a *Adder, aVal, bVal uint16) {
	// A bits: wire indices 0–15 (index 0 = MSB of A)
	// B bits: wire indices 16–31 (index 16 = MSB of B)
	for i := 0; i < 16; i++ {
		// A: bit (15-i) of aVal → wire i
		a.SetInputWire(i, (aVal>>(15-uint(i)))&1 == 1)
		// B: bit (15-i) of bVal → wire 16+i
		a.SetInputWire(16+i, (bVal>>(15-uint(i)))&1 == 1)
	}
}

func getAdderResult(a *Adder) uint16 {
	var result uint16
	for i := 0; i < 16; i++ {
		if a.GetOutputWire(i) {
			result |= 1 << (15 - uint(i))
		}
	}
	return result
}

func TestAdder(t *testing.T) {
	cases := []struct {
		aVal, bVal uint16
		carryIn    bool
		wantResult uint16
		wantCarry  bool
	}{
		{0, 0, false, 0, false},
		{1, 1, false, 2, false},
		{0xFFFF, 1, false, 0x0000, true},
		{0xFFFF, 0xFFFF, false, 0xFFFE, true},
		{1, 1, true, 3, false},
		{100, 200, false, 300, false},
		{0x8000, 0x8000, false, 0x0000, true},
	}

	for _, c := range cases {
		adder := NewAdder()
		setAdderInputs(adder, c.aVal, c.bVal)
		adder.Update(c.carryIn)
		got := getAdderResult(adder)
		if got != c.wantResult {
			t.Errorf("Adder(%#x + %#x, carryIn=%v) = %#x, want %#x", c.aVal, c.bVal, c.carryIn, got, c.wantResult)
		}
		if adder.Carry() != c.wantCarry {
			t.Errorf("Adder(%#x + %#x, carryIn=%v).Carry() = %v, want %v", c.aVal, c.bVal, c.carryIn, adder.Carry(), c.wantCarry)
		}
	}
}
