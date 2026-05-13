package components

import "testing"

// --- Compare2 ---

func TestCompare2EqualBitsZero(t *testing.T) {
	c := NewCompare2()
	c.Update(false, false, true, false)
	if !c.Equal() {
		t.Error("(0,0,eq=true,lg=false): Equal should be true")
	}
	if c.Larger() {
		t.Error("(0,0,eq=true,lg=false): Larger should be false")
	}
}

func TestCompare2ALargerThanB(t *testing.T) {
	c := NewCompare2()
	c.Update(true, false, true, false)
	if c.Equal() {
		t.Error("(1,0,eq=true,lg=false): Equal should be false")
	}
	if !c.Larger() {
		t.Error("(1,0,eq=true,lg=false): Larger should be true")
	}
}

func TestCompare2BLargerThanA(t *testing.T) {
	c := NewCompare2()
	c.Update(false, true, true, false)
	if c.Equal() {
		t.Error("(0,1,eq=true,lg=false): Equal should be false")
	}
	if c.Larger() {
		t.Error("(0,1,eq=true,lg=false): Larger should be false")
	}
}

func TestCompare2EqualBitsOne(t *testing.T) {
	c := NewCompare2()
	c.Update(true, true, true, false)
	if !c.Equal() {
		t.Error("(1,1,eq=true,lg=false): Equal should be true")
	}
	if c.Larger() {
		t.Error("(1,1,eq=true,lg=false): Larger should be false")
	}
}

func TestCompare2IsLargerInCarriesThrough(t *testing.T) {
	// equalIn=false means higher bits already differed; isLargerIn=true carries forward
	c := NewCompare2()
	c.Update(true, false, false, true)
	if c.Equal() {
		t.Error("(1,0,eq=false,lg=true): Equal should be false")
	}
	if !c.Larger() {
		t.Error("(1,0,eq=false,lg=true): Larger should be true (carried from isLargerIn)")
	}
}

// --- Comparator ---

func TestComparatorEqualValues(t *testing.T) {
	c := NewComparator()
	setWires(c.SetInputWire, 0, 0x0001)
	setWires(c.SetInputWire, BUS_WIDTH, 0x0001)
	c.Update()
	if !c.Equal() {
		t.Error("0x0001 == 0x0001: Equal should be true")
	}
	if c.Larger() {
		t.Error("0x0001 == 0x0001: Larger should be false")
	}
}

func TestComparatorALarger(t *testing.T) {
	c := NewComparator()
	setWires(c.SetInputWire, 0, 0x0002)
	setWires(c.SetInputWire, BUS_WIDTH, 0x0001)
	c.Update()
	if c.Equal() {
		t.Error("0x0002 > 0x0001: Equal should be false")
	}
	if !c.Larger() {
		t.Error("0x0002 > 0x0001: Larger should be true")
	}
}

func TestComparatorBLarger(t *testing.T) {
	c := NewComparator()
	setWires(c.SetInputWire, 0, 0x0001)
	setWires(c.SetInputWire, BUS_WIDTH, 0x0002)
	c.Update()
	if c.Equal() {
		t.Error("0x0001 < 0x0002: Equal should be false")
	}
	if c.Larger() {
		t.Error("0x0001 < 0x0002: Larger should be false")
	}
}

func TestComparatorAllFFFF(t *testing.T) {
	c := NewComparator()
	setWires(c.SetInputWire, 0, 0xFFFF)
	setWires(c.SetInputWire, BUS_WIDTH, 0xFFFF)
	c.Update()
	if !c.Equal() {
		t.Error("0xFFFF == 0xFFFF: Equal should be true")
	}
}

func TestComparatorMSBDecides(t *testing.T) {
	c := NewComparator()
	setWires(c.SetInputWire, 0, 0x8000)
	setWires(c.SetInputWire, BUS_WIDTH, 0x7FFF)
	c.Update()
	if !c.Larger() {
		t.Error("0x8000 > 0x7FFF: Larger should be true")
	}
}

// --- BusOne ---

func TestBusOneDisabledPassThrough(t *testing.T) {
	in := NewBus(BUS_WIDTH)
	out := NewBus(BUS_WIDTH)
	b := NewBusOne(in, out)
	in.SetValue(0x1234)
	b.Disable()
	b.Update()
	if got := readUint16(out.GetOutputWire); got != 0x1234 {
		t.Errorf("disabled passthrough: got 0x%04X, want 0x1234", got)
	}
}

func TestBusOneEnabledZeroInput(t *testing.T) {
	in := NewBus(BUS_WIDTH)
	out := NewBus(BUS_WIDTH)
	b := NewBusOne(in, out)
	in.SetValue(0x0000)
	b.Enable()
	b.Update()
	if got := readUint16(out.GetOutputWire); got != 0x0001 {
		t.Errorf("enabled zero input: got 0x%04X, want 0x0001", got)
	}
}

func TestBusOneEnabledAllOnes(t *testing.T) {
	in := NewBus(BUS_WIDTH)
	out := NewBus(BUS_WIDTH)
	b := NewBusOne(in, out)
	in.SetValue(0xFFFF)
	b.Enable()
	b.Update()
	if got := readUint16(out.GetOutputWire); got != 0x0001 {
		t.Errorf("enabled 0xFFFF: got 0x%04X, want 0x0001", got)
	}
}

func TestBusOneEnabledNonZeroInput(t *testing.T) {
	in := NewBus(BUS_WIDTH)
	out := NewBus(BUS_WIDTH)
	b := NewBusOne(in, out)
	in.SetValue(0x0002)
	b.Enable()
	b.Update()
	if got := readUint16(out.GetOutputWire); got != 0x0001 {
		t.Errorf("enabled 0x0002: got 0x%04X, want 0x0001", got)
	}
}
