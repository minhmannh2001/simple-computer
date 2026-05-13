package components

import "testing"

func busValue(bus *Bus) uint16 {
	var result uint16
	for i := 0; i < BUS_WIDTH; i++ {
		if bus.GetOutputWire(i) {
			result |= 1 << (BUS_WIDTH - 1 - uint(i))
		}
	}
	return result
}

func latch(r *Register, bus *Bus, value uint16) {
	bus.SetValue(value)
	r.Set()
	r.Update()
	r.Unset()
	r.Update()
}

func TestRegisterStoreValue(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0x1234)

	if got := r.Value(); got != 0x1234 {
		t.Errorf("Value() = 0x%04X, want 0x1234", got)
	}
}

func TestRegisterEnablePushesToBus(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0x5678)

	bus.SetValue(0x0000)
	r.Enable()
	r.Update()

	if got := busValue(bus); got != 0x5678 {
		t.Errorf("bus after Enable = 0x%04X, want 0x5678", got)
	}
}

func TestRegisterDisabledDoesNotDriveBus(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0xFFFF)

	bus.SetValue(0x0000)
	r.Disable()
	r.Update()

	if got := busValue(bus); got != 0x0000 {
		t.Errorf("bus after Disable = 0x%04X, want 0x0000", got)
	}
}

func TestRegisterUnsetDoesNotLatch(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0xAAAA)

	bus.SetValue(0xBBBB)
	r.Unset()
	r.Update()

	if got := r.Value(); got != 0xAAAA {
		t.Errorf("Value() after unset update = 0x%04X, want 0xAAAA", got)
	}
}

func TestRegisterBitLSB(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0x0001)

	if !r.Bit(15) {
		t.Error("Bit(15) should be true for 0x0001")
	}
	for i := 0; i < 15; i++ {
		if r.Bit(i) {
			t.Errorf("Bit(%d) should be false for 0x0001", i)
		}
	}
}

func TestRegisterBitMSB(t *testing.T) {
	bus := NewBus(BUS_WIDTH)
	r := NewRegister("R0", bus, bus)

	latch(r, bus, 0x8000)

	if !r.Bit(0) {
		t.Error("Bit(0) should be true for 0x8000")
	}
	for i := 1; i < BUS_WIDTH; i++ {
		if r.Bit(i) {
			t.Errorf("Bit(%d) should be false for 0x8000", i)
		}
	}
}

func TestRegisterSeparateBuses(t *testing.T) {
	inputBus := NewBus(BUS_WIDTH)
	outputBus := NewBus(BUS_WIDTH)
	r := NewRegister("R1", inputBus, outputBus)

	inputBus.SetValue(0xABCD)
	r.Set()
	r.Update()
	r.Unset()

	r.Enable()
	r.Update()

	if got := busValue(outputBus); got != 0xABCD {
		t.Errorf("outputBus = 0x%04X, want 0xABCD", got)
	}
	if got := busValue(inputBus); got != 0xABCD {
		t.Errorf("inputBus should be unchanged = 0x%04X, want 0xABCD", got)
	}
}
