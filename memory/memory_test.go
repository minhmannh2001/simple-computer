package memory_test

import (
	"testing"

	"simple-computer/components"
	"simple-computer/memory"
)

func newTestMemory() (*memory.Memory64K, *components.Bus) {
	bus := components.NewBus(components.BUS_WIDTH)
	return memory.NewMemory64K(bus), bus
}

// writeToRAM loads addr into the MAR, then writes value from the bus into the selected cell.
func writeToRAM(mem *memory.Memory64K, bus *components.Bus, addr, value uint16) {
	bus.SetValue(addr)
	mem.AddressRegister.Set()
	mem.Update()
	mem.AddressRegister.Unset()

	bus.SetValue(value)
	mem.Set()
	mem.Update()
	mem.Unset()
	bus.SetValue(0)
}

// readFromRAM loads addr into the MAR, clears the bus, then enables the selected cell to drive the bus.
func readFromRAM(mem *memory.Memory64K, bus *components.Bus, addr uint16) uint16 {
	bus.SetValue(addr)
	mem.AddressRegister.Set()
	mem.Update()
	mem.AddressRegister.Unset()

	bus.SetValue(0)
	mem.Enable()
	mem.Update()
	mem.Disable()

	var result uint16
	for i := range components.BUS_WIDTH {
		if bus.GetOutputWire(i) {
			result |= 1 << uint(components.BUS_WIDTH-1-i)
		}
	}
	return result
}

// --- Write then read ---

func TestMemory_BasicWriteRead(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0000, 0x1234)
	if got := readFromRAM(mem, bus, 0x0000); got != 0x1234 {
		t.Errorf("addr 0x0000: got %#x, want 0x1234", got)
	}
}

func TestMemory_WriteReadAddr0001(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0001, 0xABCD)
	if got := readFromRAM(mem, bus, 0x0001); got != 0xABCD {
		t.Errorf("addr 0x0001: got %#x, want 0xABCD", got)
	}
}

func TestMemory_TwoAddrIndependent(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0000, 0x1111)
	writeToRAM(mem, bus, 0x0001, 0x2222)

	if got := readFromRAM(mem, bus, 0x0000); got != 0x1111 {
		t.Errorf("addr 0x0000: got %#x, want 0x1111", got)
	}
	if got := readFromRAM(mem, bus, 0x0001); got != 0x2222 {
		t.Errorf("addr 0x0001: got %#x, want 0x2222", got)
	}
}

// --- Address independence (different rows) ---

func TestMemory_RowIndependence(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0000, 0x1111)
	writeToRAM(mem, bus, 0x0100, 0x2222)

	if got := readFromRAM(mem, bus, 0x0000); got != 0x1111 {
		t.Errorf("addr 0x0000: got %#x, want 0x1111", got)
	}
	if got := readFromRAM(mem, bus, 0x0100); got != 0x2222 {
		t.Errorf("addr 0x0100: got %#x, want 0x2222", got)
	}
}

// --- Boundary addresses ---

func TestMemory_BoundaryFF00(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0xFF00, 0xBEEF)
	if got := readFromRAM(mem, bus, 0xFF00); got != 0xBEEF {
		t.Errorf("addr 0xFF00: got %#x, want 0xBEEF", got)
	}
}

func TestMemory_Boundary00FF(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x00FF, 0xCAFE)
	if got := readFromRAM(mem, bus, 0x00FF); got != 0xCAFE {
		t.Errorf("addr 0x00FF: got %#x, want 0xCAFE", got)
	}
}

func TestMemory_Boundary0500(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0500, 0xF00D)
	if got := readFromRAM(mem, bus, 0x0500); got != 0xF00D {
		t.Errorf("addr 0x0500: got %#x, want 0xF00D", got)
	}
}

// --- Overwrite ---

func TestMemory_Overwrite(t *testing.T) {
	mem, bus := newTestMemory()
	writeToRAM(mem, bus, 0x0000, 0xAAAA)
	writeToRAM(mem, bus, 0x0000, 0xBBBB)
	if got := readFromRAM(mem, bus, 0x0000); got != 0xBBBB {
		t.Errorf("overwrite addr 0x0000: got %#x, want 0xBBBB", got)
	}
}
