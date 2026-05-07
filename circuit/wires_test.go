package circuit

import "testing"

func TestWire_InitialValue(t *testing.T) {
	w := NewWire("a", true)
	if w.Get() != true {
		t.Fatalf("expected true, got false")
	}
}

func TestWire_UpdateAndGet(t *testing.T) {
	w := NewWire("a", false)
	w.Update(true)
	if w.Get() != true {
		t.Fatalf("expected true after Update(true)")
	}
	w.Update(false)
	if w.Get() != false {
		t.Fatalf("expected false after Update(false)")
	}
}
