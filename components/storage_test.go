package components

import "testing"

// --- Bit tests ---

func TestBitInitialState(t *testing.T) {
	b := NewBit()
	if b.Get() != false {
		t.Error("new Bit should return false initially")
	}
}

func TestBitSetFalseDoesNotChange(t *testing.T) {
	b := NewBit()
	b.Update(true, false) // set=false → no change
	if b.Get() != false {
		t.Error("Update with set=false should not change value")
	}
}

func TestBitLatchTrue(t *testing.T) {
	b := NewBit()
	b.Update(true, true)
	if b.Get() != true {
		t.Error("Update(true, true) should latch true")
	}
}

func TestBitLatchFalse(t *testing.T) {
	b := NewBit()
	b.Update(true, true) // latch true first
	b.Update(false, true)
	if b.Get() != false {
		t.Error("Update(false, true) should latch false")
	}
}

func TestBitHoldsAfterSetFalse(t *testing.T) {
	b := NewBit()
	b.Update(true, true) // latch true
	b.Update(false, false)
	if b.Get() != true {
		t.Error("Update(false, false) after latching true should still return true")
	}
}

func TestBitHoldsWhenInputChangesButSetFalse(t *testing.T) {
	b := NewBit()
	b.Update(true, true) // latch true
	b.Update(true, false) // set=false: no change despite input=true
	if b.Get() != true {
		t.Error("Update(true, false) after latching true should still return true (S=0 no change)")
	}
}

// --- Word tests ---

func TestWordInitialState(t *testing.T) {
	w := NewWord()
	for i := 0; i < BUS_WIDTH; i++ {
		if w.GetOutputWire(i) != false {
			t.Errorf("wire %d should be false initially", i)
		}
	}
}

func TestWordLatchAllTrue(t *testing.T) {
	w := NewWord()
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, true)
	}
	w.Update(true)
	for i := 0; i < BUS_WIDTH; i++ {
		if w.GetOutputWire(i) != true {
			t.Errorf("wire %d should be true after Update(true) with all inputs true", i)
		}
	}
}

func TestWordHoldsWithSetFalse(t *testing.T) {
	w := NewWord()
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, true)
	}
	w.Update(true)
	w.Update(false)
	for i := 0; i < BUS_WIDTH; i++ {
		if w.GetOutputWire(i) != true {
			t.Errorf("wire %d should still be true after Update(false)", i)
		}
	}
}

func TestWordDoesNotLatchWhenSetFalse(t *testing.T) {
	w := NewWord()
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, true)
	}
	w.Update(true) // latch all true
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, false)
	}
	w.Update(false) // set=false: inputs don't get latched
	for i := 0; i < BUS_WIDTH; i++ {
		if w.GetOutputWire(i) != true {
			t.Errorf("wire %d should still be true (not latched new false)", i)
		}
	}
}

func TestWordLatchAllFalse(t *testing.T) {
	w := NewWord()
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, true)
	}
	w.Update(true) // latch true
	for i := 0; i < BUS_WIDTH; i++ {
		w.SetInputWire(i, false)
	}
	w.Update(true) // latch false
	for i := 0; i < BUS_WIDTH; i++ {
		if w.GetOutputWire(i) != false {
			t.Errorf("wire %d should be false after latching false", i)
		}
	}
}

func TestWordPartialUpdate(t *testing.T) {
	w := NewWord()
	w.SetInputWire(7, true)
	w.Update(true)
	for i := 0; i < BUS_WIDTH; i++ {
		want := i == 7
		if w.GetOutputWire(i) != want {
			t.Errorf("wire %d = %v, want %v", i, w.GetOutputWire(i), want)
		}
	}
}
