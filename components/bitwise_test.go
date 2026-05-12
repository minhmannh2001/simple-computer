package components

import "testing"

// --- test helpers ---

// setWires sets BUS_WIDTH wires starting at offset from a uint16 (MSB at offset, LSB at offset+15).
func setWires(setter func(int, bool), offset int, value uint16) {
	for i := BUS_WIDTH - 1; i >= 0; i-- {
		setter(offset+i, value&1 == 1)
		value >>= 1
	}
}

// readUint16 reads BUS_WIDTH output wires from index 0 as a uint16 (wire 0 = MSB).
func readUint16(getter func(int) bool) uint16 {
	var result uint16
	for i := range BUS_WIDTH {
		if getter(i) {
			result |= 1 << uint(BUS_WIDTH-1-i)
		}
	}
	return result
}

// --- Enabler ---

func TestEnablerBlocksWhenDisabled(t *testing.T) {
	e := NewEnabler()
	for i := range BUS_WIDTH {
		e.SetInputWire(i, true)
	}
	e.Update(false)
	for i := range BUS_WIDTH {
		if e.GetOutputWire(i) {
			t.Errorf("wire %d should be false when disabled", i)
		}
	}
}

func TestEnablerPassesWhenEnabled(t *testing.T) {
	e := NewEnabler()
	for i := range BUS_WIDTH {
		e.SetInputWire(i, true)
	}
	e.Update(true)
	for i := range BUS_WIDTH {
		if !e.GetOutputWire(i) {
			t.Errorf("wire %d should be true when enabled with all-true inputs", i)
		}
	}
}

func TestEnablerMirrorsMixedInputs(t *testing.T) {
	e := NewEnabler()
	for i := range BUS_WIDTH {
		e.SetInputWire(i, i%2 == 0) // alternating: even=true, odd=false
	}
	e.Update(true)
	for i := range BUS_WIDTH {
		want := i%2 == 0
		if e.GetOutputWire(i) != want {
			t.Errorf("wire %d: got %v, want %v", i, e.GetOutputWire(i), want)
		}
	}
}

// --- NOTer ---

func TestNOTerAllTrue(t *testing.T) {
	n := NewNOTer()
	for i := range BUS_WIDTH {
		n.SetInputWire(i, true)
	}
	n.Update()
	for i := range BUS_WIDTH {
		if n.GetOutputWire(i) {
			t.Errorf("wire %d should be false after NOT(true)", i)
		}
	}
}

func TestNOTerAlternating(t *testing.T) {
	n := NewNOTer()
	for i := range BUS_WIDTH {
		n.SetInputWire(i, i%2 == 0)
	}
	n.Update()
	for i := range BUS_WIDTH {
		want := i%2 != 0
		if n.GetOutputWire(i) != want {
			t.Errorf("wire %d: got %v, want %v", i, n.GetOutputWire(i), want)
		}
	}
}

// --- ANDer ---

func TestANDerNoOverlap(t *testing.T) {
	a := NewANDer()
	setWires(a.SetInputWire, 0, 0xFF00)        // A
	setWires(a.SetInputWire, BUS_WIDTH, 0x00FF) // B
	a.Update()
	if got := readUint16(a.GetOutputWire); got != 0x0000 {
		t.Errorf("AND(0xFF00, 0x00FF) = 0x%04X, want 0x0000", got)
	}
}

func TestANDerAllOnes(t *testing.T) {
	a := NewANDer()
	setWires(a.SetInputWire, 0, 0xFFFF)
	setWires(a.SetInputWire, BUS_WIDTH, 0xFFFF)
	a.Update()
	if got := readUint16(a.GetOutputWire); got != 0xFFFF {
		t.Errorf("AND(0xFFFF, 0xFFFF) = 0x%04X, want 0xFFFF", got)
	}
}

// --- ORer ---

func TestORerComplementary(t *testing.T) {
	o := NewORer()
	setWires(o.SetInputWire, 0, 0xFF00)
	setWires(o.SetInputWire, BUS_WIDTH, 0x00FF)
	o.Update()
	if got := readUint16(o.GetOutputWire); got != 0xFFFF {
		t.Errorf("OR(0xFF00, 0x00FF) = 0x%04X, want 0xFFFF", got)
	}
}

func TestORerAllZero(t *testing.T) {
	o := NewORer()
	setWires(o.SetInputWire, 0, 0x0000)
	setWires(o.SetInputWire, BUS_WIDTH, 0x0000)
	o.Update()
	if got := readUint16(o.GetOutputWire); got != 0x0000 {
		t.Errorf("OR(0x0000, 0x0000) = 0x%04X, want 0x0000", got)
	}
}

// --- XORer ---

func TestXORerSameInputs(t *testing.T) {
	x := NewXORer()
	setWires(x.SetInputWire, 0, 0xA5A5)
	setWires(x.SetInputWire, BUS_WIDTH, 0xA5A5)
	x.Update()
	if got := readUint16(x.GetOutputWire); got != 0x0000 {
		t.Errorf("XOR(A, A) = 0x%04X, want 0x0000", got)
	}
}

func TestXORerOnesVsZero(t *testing.T) {
	x := NewXORer()
	setWires(x.SetInputWire, 0, 0xFFFF)
	setWires(x.SetInputWire, BUS_WIDTH, 0x0000)
	x.Update()
	if got := readUint16(x.GetOutputWire); got != 0xFFFF {
		t.Errorf("XOR(0xFFFF, 0x0000) = 0x%04X, want 0xFFFF", got)
	}
}

// --- LeftShifter ---

func TestLeftShifterLSB(t *testing.T) {
	l := NewLeftShifter()
	setWires(l.SetInputWire, 0, 0x0001)
	l.Update(false)
	if got := readUint16(l.GetOutputWire); got != 0x0002 {
		t.Errorf("LeftShift(0x0001, 0) = 0x%04X, want 0x0002", got)
	}
	if l.ShiftOut() {
		t.Error("shiftOut should be false (bit 0 was 0)")
	}
}

func TestLeftShifterMSB(t *testing.T) {
	l := NewLeftShifter()
	setWires(l.SetInputWire, 0, 0x8000)
	l.Update(false)
	if got := readUint16(l.GetOutputWire); got != 0x0000 {
		t.Errorf("LeftShift(0x8000, 0) = 0x%04X, want 0x0000", got)
	}
	if !l.ShiftOut() {
		t.Error("shiftOut should be true (bit 0 was 1)")
	}
}

func TestLeftShifterWithShiftIn(t *testing.T) {
	l := NewLeftShifter()
	setWires(l.SetInputWire, 0, 0x8001)
	l.Update(true)
	if got := readUint16(l.GetOutputWire); got != 0x0003 {
		t.Errorf("LeftShift(0x8001, 1) = 0x%04X, want 0x0003", got)
	}
	if !l.ShiftOut() {
		t.Error("shiftOut should be true (bit 0 was 1)")
	}
}

// --- RightShifter ---

func TestRightShifterMSB(t *testing.T) {
	r := NewRightShifter()
	setWires(r.SetInputWire, 0, 0x8000)
	r.Update(false)
	if got := readUint16(r.GetOutputWire); got != 0x4000 {
		t.Errorf("RightShift(0x8000, 0) = 0x%04X, want 0x4000", got)
	}
	if r.ShiftOut() {
		t.Error("shiftOut should be false (bit 15 was 0)")
	}
}

func TestRightShifterLSB(t *testing.T) {
	r := NewRightShifter()
	setWires(r.SetInputWire, 0, 0x0001)
	r.Update(false)
	if got := readUint16(r.GetOutputWire); got != 0x0000 {
		t.Errorf("RightShift(0x0001, 0) = 0x%04X, want 0x0000", got)
	}
	if !r.ShiftOut() {
		t.Error("shiftOut should be true (bit 15 was 1)")
	}
}

func TestRightShifterWithShiftIn(t *testing.T) {
	r := NewRightShifter()
	setWires(r.SetInputWire, 0, 0x0001)
	r.Update(true)
	if got := readUint16(r.GetOutputWire); got != 0x8000 {
		t.Errorf("RightShift(0x0001, 1) = 0x%04X, want 0x8000", got)
	}
	if !r.ShiftOut() {
		t.Error("shiftOut should be true (bit 15 was 1)")
	}
}

// --- IsZero ---

func TestIsZeroAllFalse(t *testing.T) {
	z := NewIsZero()
	z.Update()
	if !z.GetOutputWire(0) {
		t.Error("IsZero should return true when all inputs are false")
	}
}

func TestIsZeroAnyTrue(t *testing.T) {
	z := NewIsZero()
	z.SetInputWire(7, true)
	z.Update()
	if z.GetOutputWire(0) {
		t.Error("IsZero should return false when any input is true")
	}
}
