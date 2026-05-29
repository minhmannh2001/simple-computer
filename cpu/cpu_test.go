package cpu

import (
	"testing"

	"simple-computer/components"
	"simple-computer/memory"
)

const busWidth = 16

func newTestCPU() (*CPU, *memory.Memory64K, *components.Bus) {
	bus := components.NewBus(busWidth)
	mem := memory.NewMemory64K(bus)
	cpu := NewCPU(bus, mem)
	return cpu, mem, bus
}

func writeMem(mem *memory.Memory64K, bus *components.Bus, addr, val uint16) {
	mem.AddressRegister.Set()
	bus.SetValue(addr)
	mem.Update()
	mem.AddressRegister.Unset()
	mem.Update()
	bus.SetValue(val)
	mem.Set()
	mem.Update()
	mem.Unset()
	mem.Update()
}

func setReg(c *CPU, reg int, val uint16) {
	var r *components.Register
	switch reg {
	case 0:
		r = &c.gpReg0
	case 1:
		r = &c.gpReg1
	case 2:
		r = &c.gpReg2
	case 3:
		r = &c.gpReg3
	}
	r.Set()
	r.Update()
	c.mainBus.SetValue(val)
	r.Update()
	r.Unset()
	r.Update()
}

func runInstr(c *CPU) {
	for i := 0; i < 6; i++ {
		c.Step()
	}
}

// Tracer bullet: DATA loads an immediate 16-bit value into a register.
func TestDATALoadsImmediateIntoRegister(t *testing.T) {
	c, mem, bus := newTestCPU()
	// DATA R0, 0x0042
	// opcode 0x0020 at address 0, immediate 0x0042 at address 1
	writeMem(mem, bus, 0x0000, 0x0020)
	writeMem(mem, bus, 0x0001, 0x0042)
	c.SetIAR(0x0000)

	runInstr(c)

	if got := c.gpReg0.Value(); got != 0x0042 {
		t.Errorf("DATA R0, 0x0042: R0 = 0x%04X, want 0x0042", got)
	}
}
