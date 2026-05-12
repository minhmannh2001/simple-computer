package components

import "testing"

func TestBusInitialState(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	for i := range BUS_WIDTH {
		if b.GetOutputWire(i) != false {
			t.Errorf("wire %d should be false initially", i)
		}
	}
}

func TestBusSetInputWire(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetInputWire(0, true)
	if !b.GetOutputWire(0) {
		t.Error("wire 0 should be true after SetInputWire(0, true)")
	}
	for i := 1; i < BUS_WIDTH; i++ {
		if b.GetOutputWire(i) {
			t.Errorf("wire %d should still be false", i)
		}
	}
}

func TestBusSetValueLSB(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0x0001)
	if !b.GetOutputWire(15) {
		t.Error("wire 15 (LSB) should be true for 0x0001")
	}
	for i := range 15 {
		if b.GetOutputWire(i) {
			t.Errorf("wire %d should be false for 0x0001", i)
		}
	}
}

func TestBusSetValueMSB(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0x8000)
	if !b.GetOutputWire(0) {
		t.Error("wire 0 (MSB) should be true for 0x8000")
	}
	for i := 1; i < BUS_WIDTH; i++ {
		if b.GetOutputWire(i) {
			t.Errorf("wire %d should be false for 0x8000", i)
		}
	}
}

func TestBusSetValueAllOnes(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0xFFFF)
	for i := range BUS_WIDTH {
		if !b.GetOutputWire(i) {
			t.Errorf("wire %d should be true for 0xFFFF", i)
		}
	}
}

func TestBusSetValueAllZeros(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0xFFFF)
	b.SetValue(0x0000)
	for i := range BUS_WIDTH {
		if b.GetOutputWire(i) {
			t.Errorf("wire %d should be false for 0x0000", i)
		}
	}
}

func TestBusSetValueLowByte(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0x00FF)
	for i := range 8 {
		if b.GetOutputWire(i) {
			t.Errorf("wire %d should be false for 0x00FF", i)
		}
	}
	for i := 8; i < BUS_WIDTH; i++ {
		if !b.GetOutputWire(i) {
			t.Errorf("wire %d should be true for 0x00FF", i)
		}
	}
}

func TestBusStringAllOnes(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	b.SetValue(0xFFFF)
	want := "1111111111111111"
	if got := b.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBusStringAllZeros(t *testing.T) {
	b := NewBus(BUS_WIDTH)
	want := "0000000000000000"
	if got := b.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
