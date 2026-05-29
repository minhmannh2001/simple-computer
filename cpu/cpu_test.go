package cpu

import (
	"testing"

	"simple-computer/components"
	"simple-computer/memory"
)

const busWidth = 16

// ---- setup helpers ----

func newTestCPU() (*CPU, *memory.Memory64K, *components.Bus) {
	bus := components.NewBus(busWidth)
	mem := memory.NewMemory64K(bus)
	cpu := NewCPU(bus, mem)
	return cpu, mem, bus
}

// writeMem is kept for the tracer-bullet test; new tests use setMemLoc.
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

func setMemLoc(c *CPU, addr, val uint16) {
	c.memory.AddressRegister.Set()
	c.mainBus.SetValue(addr)
	c.memory.Update()
	c.memory.AddressRegister.Unset()
	c.memory.Update()
	c.mainBus.SetValue(val)
	c.memory.Set()
	c.memory.Update()
	c.memory.Unset()
	c.memory.Update()
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

func setRegisters(c *CPU, values [4]uint16) {
	for i, v := range values {
		setReg(c, i, v)
	}
}

func runInstr(c *CPU) {
	for i := 0; i < 6; i++ {
		c.Step()
	}
}

// ---- check helpers ----

func checkIAR(c *CPU, want uint16, t *testing.T) {
	t.Helper()
	if got := c.iar.Value(); got != want {
		t.Fatalf("IAR: got 0x%04X, want 0x%04X", got, want)
	}
}

func checkIR(c *CPU, want uint16, t *testing.T) {
	t.Helper()
	if got := c.ir.Value(); got != want {
		t.Fatalf("IR: got 0x%04X, want 0x%04X", got, want)
	}
}

func checkRegister(c *CPU, reg int, want uint16, t *testing.T) {
	t.Helper()
	var got uint16
	switch reg {
	case 0:
		got = c.gpReg0.Value()
	case 1:
		got = c.gpReg1.Value()
	case 2:
		got = c.gpReg2.Value()
	case 3:
		got = c.gpReg3.Value()
	default:
		t.Fatalf("unknown register %d", reg)
	}
	if got != want {
		t.Fatalf("R%d: got 0x%04X, want 0x%04X", reg, got, want)
	}
}

func checkRegisters(c *CPU, r0, r1, r2, r3 uint16, t *testing.T) {
	t.Helper()
	checkRegister(c, 0, r0, t)
	checkRegister(c, 1, r1, t)
	checkRegister(c, 2, r2, t)
	checkRegister(c, 3, r3, t)
}

func checkFlagsRegister(c *CPU, carry, larger, equal, zero bool, t *testing.T) {
	t.Helper()
	if got := c.flagsBus.GetOutputWire(FLAGS_BUS_CARRY); got != carry {
		t.Fatalf("flags carry: got %v, want %v", got, carry)
	}
	if got := c.flagsBus.GetOutputWire(FLAGS_BUS_A_LARGER); got != larger {
		t.Fatalf("flags larger: got %v, want %v", got, larger)
	}
	if got := c.flagsBus.GetOutputWire(FLAGS_BUS_EQUAL); got != equal {
		t.Fatalf("flags equal: got %v, want %v", got, equal)
	}
	if got := c.flagsBus.GetOutputWire(FLAGS_BUS_ZERO); got != zero {
		t.Fatalf("flags zero: got %v, want %v", got, zero)
	}
}

// ---- DumbPeripheral (used by IO tests) ----

type DumbPeripheral struct {
	ioBus   *components.IOBus
	mainBus *components.Bus

	value             components.Register
	outputDataMode    bool
	outputAddressMode bool
}

func NewDumbPeripheral() *DumbPeripheral {
	return new(DumbPeripheral)
}

func (p *DumbPeripheral) Connect(ioBus *components.IOBus, mainBus *components.Bus) {
	p.ioBus = ioBus
	p.mainBus = mainBus
	p.value = *components.NewRegister("P", p.mainBus, p.mainBus)
}

func (p *DumbPeripheral) Update() {
	p.updateEnabled()
	p.value.Update()
	p.updateSet()
	p.value.Update()
}

func (p *DumbPeripheral) refreshValue(addrVal, dataVal uint16) {
	p.value.Set()
	p.value.Update()
	if p.ioBus.GetOutputWire(components.DATA_OR_ADDRESS) {
		p.mainBus.SetValue(addrVal)
	} else {
		p.mainBus.SetValue(dataVal)
	}
	p.value.Update()
	p.value.Unset()
	p.value.Update()
}

func (p *DumbPeripheral) updateEnabled() {
	if p.ioBus.GetOutputWire(components.CLOCK_ENABLE) {
		p.refreshValue(0x00AD, 0x00DA)
		p.value.Enable()
	} else {
		p.value.Disable()
	}
}

func (p *DumbPeripheral) updateSet() {
	if p.ioBus.GetOutputWire(components.CLOCK_SET) {
		p.outputDataMode = !p.ioBus.GetOutputWire(components.DATA_OR_ADDRESS)
		p.outputAddressMode = p.ioBus.GetOutputWire(components.DATA_OR_ADDRESS)
		p.value.Set()
	} else {
		p.value.Unset()
	}
}

// ============================================================
// Tests
// ============================================================

// Tracer bullet: DATA loads an immediate 16-bit value into a register.
func TestDATALoadsImmediateIntoRegister(t *testing.T) {
	c, mem, bus := newTestCPU()
	// DATA R0, 0x0042
	writeMem(mem, bus, 0x0000, 0x0020)
	writeMem(mem, bus, 0x0001, 0x0042)
	c.SetIAR(0x0000)

	runInstr(c)

	if got := c.gpReg0.Value(); got != 0x0042 {
		t.Errorf("DATA R0, 0x0042: R0 = 0x%04X, want 0x0042", got)
	}
}

func TestIARIncrementedOnEveryCycle(t *testing.T) {
	c, _, _ := newTestCPU()
	c.SetIAR(0x0000)

	for q := uint16(0); q < 1000; q++ {
		runInstr(c)
		checkIAR(c, q+1, t)
	}
}

func TestInstructionReceivedFromMemory(t *testing.T) {
	c, _, _ := newTestCPU()

	instructions := []uint16{0x008A, 0x0082, 0x0088, 0x0094, 0x00B1}
	var addr uint16 = 0xF00F
	for _, b := range instructions {
		setMemLoc(c, addr, b)
		addr++
	}
	c.SetIAR(0xF00F)

	for _, b := range instructions {
		runInstr(c)
		checkIR(c, b, t)
	}
}

// ---- Flags tests ----

func TestFlagsRegisterAllFalse(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081) // ADD R0, R1 → result = 9+10 = 19, no flags
	setRegisters(c, [4]uint16{0x0009, 0x000A, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, false, false, false, false, t)
}

func TestFlagsRegisterCarryFlagEnabled(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081)
	setRegisters(c, [4]uint16{0x0020, 0xFFFF, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, true, false, false, false, t)
}

func TestFlagsRegisterIsLargerFlagEnabled(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081)
	setRegisters(c, [4]uint16{0x0021, 0x0020, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, false, true, false, false, t)
}

func TestFlagsRegisterIsEqualsFlagEnabled(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081)
	setRegisters(c, [4]uint16{0x0021, 0x0021, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, false, false, true, false, t)
}

func TestFlagsRegisterIsZeroFlagEnabled(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081)
	setRegisters(c, [4]uint16{0x0001, 0xFFFF, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, true, false, false, true, t)
}

func TestFlagsRegisterMultipleEnabled(t *testing.T) {
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0081)
	setRegisters(c, [4]uint16{0xFFFF, 0x0001, 0x0002, 0x0003})
	c.SetIAR(0x0000)
	runInstr(c)
	checkFlagsRegister(c, true, true, false, true, t)
}

// ---- LD ----

func TestLD(t *testing.T) {
	testLD(0x0000, 0x0080, 0x0023, []uint16{0x0080, 0x0081, 0x0082, 0x0083}, []uint16{0x0023, 0x0081, 0x0082, 0x0083}, t)
	testLD(0x0001, 0x0084, 0x00F2, []uint16{0x0084, 0x0085, 0x0086, 0x0087}, []uint16{0x0084, 0x00F2, 0x0086, 0x0087}, t)
	testLD(0x0002, 0x0088, 0x0001, []uint16{0x0088, 0x0089, 0x008A, 0x008B}, []uint16{0x0088, 0x0089, 0x0001, 0x008B}, t)
	testLD(0x0003, 0x008C, 0x005A, []uint16{0x008C, 0x008D, 0x008E, 0x008F}, []uint16{0x008C, 0x008D, 0x008E, 0x005A}, t)

	testLD(0x0004, 0x0091, 0x0023, []uint16{0x0090, 0x0091, 0x0092, 0x0093}, []uint16{0x0023, 0x0091, 0x0092, 0x0093}, t)
	testLD(0x0005, 0x0095, 0x00F2, []uint16{0x0094, 0x0095, 0x0096, 0x0097}, []uint16{0x0094, 0x00F2, 0x0096, 0x0097}, t)
	testLD(0x0006, 0x0099, 0x0001, []uint16{0x0098, 0x0099, 0x009A, 0x009B}, []uint16{0x0098, 0x0099, 0x0001, 0x009B}, t)
	testLD(0x0007, 0x009D, 0x005A, []uint16{0x009C, 0x009D, 0x009E, 0x009F}, []uint16{0x009C, 0x009D, 0x009E, 0x005A}, t)

	testLD(0x0008, 0x00A2, 0x0023, []uint16{0x00A0, 0x00A1, 0x00A2, 0x00A3}, []uint16{0x0023, 0x00A1, 0x00A2, 0x00A3}, t)
	testLD(0x0009, 0x00A6, 0x00F2, []uint16{0x00A4, 0x00A5, 0x00A6, 0x00A7}, []uint16{0x00A4, 0x00F2, 0x00A6, 0x00A7}, t)
	testLD(0x000A, 0x00AA, 0x0001, []uint16{0x00A8, 0x00A9, 0x00AA, 0x00AB}, []uint16{0x00A8, 0x00A9, 0x0001, 0x00AB}, t)
	testLD(0x000B, 0x00AE, 0x005A, []uint16{0x00AC, 0x00AD, 0x00AE, 0x00AF}, []uint16{0x00AC, 0x00AD, 0x00AE, 0x005A}, t)

	testLD(0x000C, 0x00B3, 0x0023, []uint16{0x00B0, 0x00B1, 0x00B2, 0x00B3}, []uint16{0x0023, 0x00B1, 0x00B2, 0x00B3}, t)
	testLD(0x000D, 0x00B7, 0x00F2, []uint16{0x00B4, 0x00B5, 0x00B6, 0x00B7}, []uint16{0x00B4, 0x00F2, 0x00B6, 0x00B7}, t)
	testLD(0x000E, 0x22BB, 0xAB01, []uint16{0x00B8, 0x00B9, 0x00BA, 0x22BB}, []uint16{0x00B8, 0x00B9, 0xAB01, 0x22BB}, t)
	testLD(0x000F, 0x00BF, 0x005A, []uint16{0x00BC, 0x00BD, 0x00BE, 0x00BF}, []uint16{0x00BC, 0x00BD, 0x00BE, 0x005A}, t)
}

func testLD(instruction, memAddr, memVal uint16, inputRegs, wantRegs []uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, instruction)
	c.SetIAR(0x0000)
	setMemLoc(c, memAddr, memVal)
	for i, v := range inputRegs {
		setReg(c, i, v)
	}
	runInstr(c)
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
}

func TestSTThenLD(t *testing.T) {
	c, _, _ := newTestCPU()

	for i := uint16(0); i < 512; i++ {
		setMemLoc(c, i, 0x001B) // ST R2, R3
	}
	c.SetIAR(0x0000)

	value := uint16(0x0400)
	for i := uint16(0x0200); i < 0x0400; i++ {
		setRegisters(c, [4]uint16{0x0001, 0x0001, i, value})
		runInstr(c)
		value--
	}

	for i := uint16(0); i < 512; i++ {
		setMemLoc(c, i, 0x000B) // LD R2, R3
	}
	c.SetIAR(0x0000)

	value = 0x0400
	for i := uint16(0x0200); i < 0x0400; i++ {
		setRegisters(c, [4]uint16{0x0001, 0x0001, i, 0x0001})
		runInstr(c)
		checkRegisters(c, 0x0001, 0x0001, i, value, t)
		value--
	}
}

func TestLD4Times(t *testing.T) {
	c, _, _ := newTestCPU()

	var addr uint16 = 0x00A2
	values := []uint16{0x0088, 0x0090, 0x0092, 0x00AB}

	for i := uint16(0); i < uint16(len(values)); i++ {
		setMemLoc(c, i, 0x0001)
		setMemLoc(c, addr, values[i])
		addr++
	}
	c.SetIAR(0x0000)

	addr = 0x00A2
	for _, v := range values {
		setReg(c, 0, addr)
		setReg(c, 1, 0x0001)
		setReg(c, 2, 0x0001)
		setReg(c, 3, 0x0001)
		runInstr(c)
		checkRegisters(c, addr, v, 0x0001, 0x0001, t)
		addr++
	}
}

// ---- ST ----

func TestST(t *testing.T) {
	testST(0x0010, [4]uint16{0x00A0, 0x0001, 0x0001, 0x0001}, 0x00A0, 0x00A0, t)
	testST(0x0011, [4]uint16{0x00A1, 0x0029, 0x0001, 0x0001}, 0x00A1, 0x0029, t)
	testST(0x0012, [4]uint16{0x00A2, 0x0001, 0x007F, 0x0001}, 0x00A2, 0x007F, t)
	testST(0x0013, [4]uint16{0x00A3, 0x0001, 0x0001, 0x001B}, 0x00A3, 0x001B, t)

	testST(0x0014, [4]uint16{0x00A0, 0x00B4, 0x0001, 0x0001}, 0x00B4, 0x00A0, t)
	testST(0x0015, [4]uint16{0x0001, 0x00B5, 0x0001, 0x0001}, 0x00B5, 0x00B5, t)
	testST(0x0016, [4]uint16{0x0001, 0x00B6, 0x007F, 0x0001}, 0x00B6, 0x007F, t)
	testST(0x0017, [4]uint16{0x0001, 0x00B7, 0x0001, 0x001B}, 0x00B7, 0x001B, t)

	testST(0x0018, [4]uint16{0x00A0, 0x0001, 0x00C8, 0x0001}, 0x00C8, 0x00A0, t)
	testST(0x0019, [4]uint16{0x0001, 0x0029, 0x00C9, 0x0001}, 0x00C9, 0x0029, t)
	testST(0x001A, [4]uint16{0x0001, 0x0001, 0x00CA, 0x0001}, 0x00CA, 0x00CA, t)
	testST(0x001B, [4]uint16{0x0001, 0x0001, 0x00CB, 0x001B}, 0x00CB, 0x001B, t)

	testST(0x001C, [4]uint16{0x00A0, 0x0001, 0x0001, 0x00DC}, 0x00DC, 0x00A0, t)
	testST(0x001D, [4]uint16{0x0001, 0x0029, 0x0001, 0x00DD}, 0x00DD, 0x0029, t)
	testST(0x001E, [4]uint16{0x0001, 0x0001, 0x1A7F, 0xFCDE}, 0xFCDE, 0x1A7F, t)
	testST(0x001F, [4]uint16{0x0001, 0x0001, 0x0001, 0x00DF}, 0x00DF, 0x00DF, t)
}

func testST(instruction uint16, inputRegs [4]uint16, wantAddr, wantVal uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	const insAddr uint16 = 0x0000
	setMemLoc(c, insAddr, instruction)
	c.SetIAR(insAddr)
	setRegisters(c, inputRegs)
	runInstr(c)

	// verify by loading from the stored address
	setMemLoc(c, insAddr+1, 0x0000) // LD R0, R0
	c.SetIAR(insAddr + 1)
	setRegisters(c, [4]uint16{wantAddr, inputRegs[1], inputRegs[2], inputRegs[3]})
	runInstr(c)

	checkRegister(c, 0, wantVal, t)
}

// ---- DATA ----

func TestDATA(t *testing.T) {
	c, _, _ := newTestCPU()
	const insAddr uint16 = 0x0000

	setMemLoc(c, insAddr, 0x0020)
	setMemLoc(c, insAddr+1, 0xF071)
	setMemLoc(c, insAddr+2, 0x0021)
	setMemLoc(c, insAddr+3, 0xF172)
	setMemLoc(c, insAddr+4, 0x0022)
	setMemLoc(c, insAddr+5, 0xF273)
	setMemLoc(c, insAddr+6, 0x0023)
	setMemLoc(c, insAddr+7, 0xF374)

	c.SetIAR(insAddr)
	setRegisters(c, [4]uint16{0x0001, 0x0001, 0x0001, 0x0001})

	for i := 0; i < 4; i++ {
		runInstr(c)
	}

	checkRegisters(c, 0xF071, 0xF172, 0xF273, 0xF374, t)
	checkIAR(c, 0x0008, t)
}

// ---- JMP ----

func TestJMPR(t *testing.T) {
	testJMPR(0x0030, [4]uint16{0x0083, 0x0001, 0x0001, 0x0001}, 0x0083, t)
	testJMPR(0x0031, [4]uint16{0x0001, 0x00F1, 0x0001, 0x0001}, 0x00F1, t)
	testJMPR(0x0032, [4]uint16{0x0001, 0x0001, 0x00BB, 0x0001}, 0x00BB, t)
	testJMPR(0x0033, [4]uint16{0x0001, 0x0001, 0x0001, 0xFF19}, 0xFF19, t)
}

func testJMPR(instruction uint16, inputRegs [4]uint16, wantIAR uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, instruction)
	c.SetIAR(0x0000)
	setRegisters(c, inputRegs)
	runInstr(c)
	checkRegisters(c, inputRegs[0], inputRegs[1], inputRegs[2], inputRegs[3], t)
	checkIAR(c, wantIAR, t)
}

func TestJMP(t *testing.T) {
	for i := 0; i < 0x0400; i++ {
		testJMP(uint16(i), t)
	}
}

func testJMP(wantIAR uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, 0x0040)
	setMemLoc(c, 0x0001, wantIAR)
	c.SetIAR(0x0000)
	in := [4]uint16{0x0001, 0x0001, 0x0001, 0x0001}
	setRegisters(c, in)
	runInstr(c)
	checkRegisters(c, in[0], in[1], in[2], in[3], t)
	checkIAR(c, wantIAR, t)
}

// ---- Conditional JMPs ----

func TestJMPC(t *testing.T) {
	testJMPConditional(0x0058, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x0058, 0x0091, 0x0081, [4]uint16{0x0005, 0x0006, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPA(t *testing.T) {
	testJMPConditional(0x0054, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0054, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPE(t *testing.T) {
	testJMPConditional(0x0052, 0x00AE, 0x00F1, [4]uint16{0x0000, 0x0000, 0x0001, 0x002}, 0x00AE, t)
	testJMPConditional(0x0052, 0x00AF, 0x00F1, [4]uint16{0x0010, 0x0011, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPZ(t *testing.T) {
	testJMPConditional(0x0051, 0x00AE, 0x00B0, [4]uint16{0xFFFF, 0x0001, 0x0001, 0x001}, 0x00AE, t)
	testJMPConditional(0x0051, 0x00AF, 0x00B0, [4]uint16{0x0000, 0x0011, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCA(t *testing.T) {
	testJMPConditional(0x005C, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005C, 0x0090, 0x0081, [4]uint16{0x000A, 0x0001, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005C, 0x0091, 0x0081, [4]uint16{0x0001, 0x0001, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCE(t *testing.T) {
	testJMPConditional(0x005A, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005A, 0x0090, 0x0081, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005A, 0x0091, 0x0081, [4]uint16{0x0001, 0x0002, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCZ(t *testing.T) {
	testJMPConditional(0x0059, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x0059, 0x0090, 0x00B0, [4]uint16{0xFFFF, 0x00FE, 0x00FE, 0x00FE}, 0x0090, t)
	testJMPConditional(0x0059, 0x0091, 0x0081, [4]uint16{0x0001, 0x0002, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPAE(t *testing.T) {
	testJMPConditional(0x0056, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0056, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0056, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPAZ(t *testing.T) {
	testJMPConditional(0x0055, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0055, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x0055, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPEZ(t *testing.T) {
	testJMPConditional(0x0053, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0053, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x0053, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCAE(t *testing.T) {
	testJMPConditional(0x005E, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005E, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005E, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005E, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCAZ(t *testing.T) {
	testJMPConditional(0x005D, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005D, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005D, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x005D, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCEZ(t *testing.T) {
	testJMPConditional(0x005B, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005B, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005B, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x005B, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPAEZ(t *testing.T) {
	testJMPConditional(0x0057, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0057, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x0057, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x0057, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

func TestJMPCAEZ(t *testing.T) {
	testJMPConditional(0x005F, 0x0090, 0x0081, [4]uint16{0x0004, 0xFFFF, 0x0001, 0x002}, 0x0090, t)
	testJMPConditional(0x005F, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0001, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005F, 0x0020, 0x00F1, [4]uint16{0x0002, 0x0002, 0x0001, 0x002}, 0x0020, t)
	testJMPConditional(0x005F, 0x0020, 0x00C1, [4]uint16{0x0001, 0x00FE, 0x0002, 0x0002}, 0x0020, t)
	testJMPConditional(0x005F, 0x0021, 0x00F1, [4]uint16{0x0001, 0x0003, 0x0001, 0x0001}, 0x0003, t)
}

// testJMPConditional: executes initialInstr at addr 0, then jmpInstr at addr 1 with destination at addr 2.
func testJMPConditional(jmpInstr, dest, initialInstr uint16, inputRegs [4]uint16, wantIAR uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, initialInstr)
	setMemLoc(c, 0x0001, jmpInstr)
	setMemLoc(c, 0x0002, dest)
	c.SetIAR(0x0000)
	setRegisters(c, inputRegs)
	runInstr(c) // execute initialInstr (sets flags)
	runInstr(c) // execute conditional jump
	checkIAR(c, wantIAR, t)
}

// ---- CLF ----

func TestCLF(t *testing.T) {
	testCLF(0x0081, [4]uint16{0xFFFF, 0x0001, 0x0000, 0x0000}, t) // carry + zero + greater
	testCLF(0x0081, [4]uint16{0x0001, 0x0001, 0x0000, 0x0000}, t) // equal flag
	testCLF(0x0081, [4]uint16{0x0001, 0x0002, 0x0000, 0x0000}, t) // all flags false anyway
}

func testCLF(initialInstr uint16, initialRegs [4]uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, initialInstr)
	setMemLoc(c, 0x0001, 0x0060) // CLF
	c.SetIAR(0x0000)
	setRegisters(c, initialRegs)
	runInstr(c) // set flags
	runInstr(c) // CLF
	checkFlagsRegister(c, false, false, false, false, t)
}

// ---- ALU ----

func testInstruction(instruction uint16, inputRegs, wantRegs [4]uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, instruction)
	for i, r := range inputRegs {
		setReg(c, i, r)
	}
	c.SetIAR(0x0000)
	runInstr(c)
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
}

func TestALUAdd(t *testing.T) {
	in := [4]uint16{0x0002, 0x0003, 0xFD04, 0x0005}
	testInstruction(0x0080, in, [4]uint16{in[0] + in[0], in[1], in[2], in[3]}, t)
	testInstruction(0x0081, in, [4]uint16{in[0], in[1] + in[0], in[2], in[3]}, t)
	testInstruction(0x0082, in, [4]uint16{in[0], in[1], in[2] + in[0], in[3]}, t)
	testInstruction(0x0083, in, [4]uint16{in[0], in[1], in[2], in[3] + in[0]}, t)

	testInstruction(0x0084, in, [4]uint16{in[0] + in[1], in[1], in[2], in[3]}, t)
	testInstruction(0x0085, in, [4]uint16{in[0], in[1] + in[1], in[2], in[3]}, t)
	testInstruction(0x0086, in, [4]uint16{in[0], in[1], in[2] + in[1], in[3]}, t)
	testInstruction(0x0087, in, [4]uint16{in[0], in[1], in[2], in[3] + in[1]}, t)

	testInstruction(0x0088, in, [4]uint16{in[0] + in[2], in[1], in[2], in[3]}, t)
	testInstruction(0x0089, in, [4]uint16{in[0], in[1] + in[2], in[2], in[3]}, t)
	testInstruction(0x008A, in, [4]uint16{in[0], in[1], in[2] + in[2], in[3]}, t)
	testInstruction(0x008B, in, [4]uint16{in[0], in[1], in[2], in[3] + in[2]}, t)

	testInstruction(0x008C, in, [4]uint16{in[0] + in[3], in[1], in[2], in[3]}, t)
	testInstruction(0x008D, in, [4]uint16{in[0], in[1] + in[3], in[2], in[3]}, t)
	testInstruction(0x008E, in, [4]uint16{in[0], in[1], in[2] + in[3], in[3]}, t)
	testInstruction(0x008F, in, [4]uint16{in[0], in[1], in[2], in[3] + in[3]}, t)
}

func TestALUAddWithCarry(t *testing.T) {
	testALUAddWithCarry(0x0080,
		[4]uint16{0xFFFE, 0x0000, 0x0000, 0x0000},
		[4]uint16{0x0001, 0x0000, 0x0000, 0x0000}, t)
	testALUAddWithCarry(0x0081,
		[4]uint16{0xFFFE, 0x0005, 0x0000, 0x0000},
		[4]uint16{0x0000, 0x0001, 0x0000, 0x0000}, t)
	testALUAddWithCarry(0x0082,
		[4]uint16{0xFFFE, 0x0000, 0x0005, 0x0000},
		[4]uint16{0x0000, 0x0000, 0x0001, 0x0000}, t)
	testALUAddWithCarry(0x0083,
		[4]uint16{0xFFFE, 0x0000, 0x0000, 0x0005},
		[4]uint16{0x0000, 0x0000, 0x0000, 0x0001}, t)
}

func testALUAddWithCarry(instruction uint16, inputRegs, wantRegs [4]uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, instruction)
	setMemLoc(c, 0x0001, instruction)
	c.SetIAR(0x0000)
	setRegisters(c, inputRegs)
	runInstr(c)
	setRegisters(c, [4]uint16{0x0000, 0x0000, 0x0000, 0x0000})
	runInstr(c)
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
}

func TestALUNOT(t *testing.T) {
	in := [4]uint16{0xFFFF, 0xFEFE, 0xFDFD, 0xFCFC}
	testInstruction(0x00B0, in, [4]uint16{0x0000, 0xFEFE, 0xFDFD, 0xFCFC}, t)
	testInstruction(0x00B5, in, [4]uint16{0xFFFF, 0x0101, 0xFDFD, 0xFCFC}, t)
	testInstruction(0x00BA, in, [4]uint16{0xFFFF, 0xFEFE, 0x0202, 0xFCFC}, t)
	testInstruction(0x00BF, in, [4]uint16{0xFFFF, 0xFEFE, 0xFDFD, 0x0303}, t)
}

func TestALUAND(t *testing.T) {
	in := [4]uint16{0x0002, 0xABC3, 0x0004, 0x0005}
	testInstruction(0x00C0, in, [4]uint16{in[0] & in[0], in[1], in[2], in[3]}, t)
	testInstruction(0x00C1, in, [4]uint16{in[0], in[1] & in[0], in[2], in[3]}, t)
	testInstruction(0x00C2, in, [4]uint16{in[0], in[1], in[2] & in[0], in[3]}, t)
	testInstruction(0x00C3, in, [4]uint16{in[0], in[1], in[2], in[3] & in[0]}, t)

	testInstruction(0x00C4, in, [4]uint16{in[0] & in[1], in[1], in[2], in[3]}, t)
	testInstruction(0x00C5, in, [4]uint16{in[0], in[1] & in[1], in[2], in[3]}, t)
	testInstruction(0x00C6, in, [4]uint16{in[0], in[1], in[2] & in[1], in[3]}, t)
	testInstruction(0x00C7, in, [4]uint16{in[0], in[1], in[2], in[3] & in[1]}, t)

	testInstruction(0x00C8, in, [4]uint16{in[0] & in[2], in[1], in[2], in[3]}, t)
	testInstruction(0x00C9, in, [4]uint16{in[0], in[1] & in[2], in[2], in[3]}, t)
	testInstruction(0x00CA, in, [4]uint16{in[0], in[1], in[2] & in[2], in[3]}, t)
	testInstruction(0x00CB, in, [4]uint16{in[0], in[1], in[2], in[3] & in[2]}, t)

	testInstruction(0x00CC, in, [4]uint16{in[0] & in[3], in[1], in[2], in[3]}, t)
	testInstruction(0x00CD, in, [4]uint16{in[0], in[1] & in[3], in[2], in[3]}, t)
	testInstruction(0x00CE, in, [4]uint16{in[0], in[1], in[2] & in[3], in[3]}, t)
	testInstruction(0x00CF, in, [4]uint16{in[0], in[1], in[2], in[3] & in[3]}, t)
}

func TestALUOR(t *testing.T) {
	in := [4]uint16{0x2092, 0x0091, 0xCF45, 0x00AF}
	testInstruction(0x00D0, in, [4]uint16{in[0] | in[0], in[1], in[2], in[3]}, t)
	testInstruction(0x00D1, in, [4]uint16{in[0], in[1] | in[0], in[2], in[3]}, t)
	testInstruction(0x00D2, in, [4]uint16{in[0], in[1], in[2] | in[0], in[3]}, t)
	testInstruction(0x00D3, in, [4]uint16{in[0], in[1], in[2], in[3] | in[0]}, t)

	testInstruction(0x00D4, in, [4]uint16{in[0] | in[1], in[1], in[2], in[3]}, t)
	testInstruction(0x00D5, in, [4]uint16{in[0], in[1] | in[1], in[2], in[3]}, t)
	testInstruction(0x00D6, in, [4]uint16{in[0], in[1], in[2] | in[1], in[3]}, t)
	testInstruction(0x00D7, in, [4]uint16{in[0], in[1], in[2], in[3] | in[1]}, t)

	testInstruction(0x00D8, in, [4]uint16{in[0] | in[2], in[1], in[2], in[3]}, t)
	testInstruction(0x00D9, in, [4]uint16{in[0], in[1] | in[2], in[2], in[3]}, t)
	testInstruction(0x00DA, in, [4]uint16{in[0], in[1], in[2] | in[2], in[3]}, t)
	testInstruction(0x00DB, in, [4]uint16{in[0], in[1], in[2], in[3] | in[2]}, t)

	testInstruction(0x00DC, in, [4]uint16{in[0] | in[3], in[1], in[2], in[3]}, t)
	testInstruction(0x00DD, in, [4]uint16{in[0], in[1] | in[3], in[2], in[3]}, t)
	testInstruction(0x00DE, in, [4]uint16{in[0], in[1], in[2] | in[3], in[3]}, t)
	testInstruction(0x00DF, in, [4]uint16{in[0], in[1], in[2], in[3] | in[3]}, t)
}

func TestALUXOR(t *testing.T) {
	in := [4]uint16{0x0092, 0x8791, 0x0045, 0xD1AF}
	testInstruction(0x00E0, in, [4]uint16{in[0] ^ in[0], in[1], in[2], in[3]}, t)
	testInstruction(0x00E1, in, [4]uint16{in[0], in[1] ^ in[0], in[2], in[3]}, t)
	testInstruction(0x00E2, in, [4]uint16{in[0], in[1], in[2] ^ in[0], in[3]}, t)
	testInstruction(0x00E3, in, [4]uint16{in[0], in[1], in[2], in[3] ^ in[0]}, t)

	testInstruction(0x00E4, in, [4]uint16{in[0] ^ in[1], in[1], in[2], in[3]}, t)
	testInstruction(0x00E5, in, [4]uint16{in[0], in[1] ^ in[1], in[2], in[3]}, t)
	testInstruction(0x00E6, in, [4]uint16{in[0], in[1], in[2] ^ in[1], in[3]}, t)
	testInstruction(0x00E7, in, [4]uint16{in[0], in[1], in[2], in[3] ^ in[1]}, t)

	testInstruction(0x00E8, in, [4]uint16{in[0] ^ in[2], in[1], in[2], in[3]}, t)
	testInstruction(0x00E9, in, [4]uint16{in[0], in[1] ^ in[2], in[2], in[3]}, t)
	testInstruction(0x00EA, in, [4]uint16{in[0], in[1], in[2] ^ in[2], in[3]}, t)
	testInstruction(0x00EB, in, [4]uint16{in[0], in[1], in[2], in[3] ^ in[2]}, t)

	testInstruction(0x00EC, in, [4]uint16{in[0] ^ in[3], in[1], in[2], in[3]}, t)
	testInstruction(0x00ED, in, [4]uint16{in[0], in[1] ^ in[3], in[2], in[3]}, t)
	testInstruction(0x00EE, in, [4]uint16{in[0], in[1], in[2] ^ in[3], in[3]}, t)
	testInstruction(0x00EF, in, [4]uint16{in[0], in[1], in[2], in[3] ^ in[3]}, t)
}

func TestCMP(t *testing.T) {
	in := [4]uint16{0xAB92, 0x0091, 0x0045, 0x00AF}
	instr := uint16(0x00F0)
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			testCMP(instr, in, in, a, b, t)
			instr++
		}
	}

	zeroes := [4]uint16{0x0000, 0x0000, 0x0000, 0x0000}
	instr = 0x00F0
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			testCMP(instr, zeroes, zeroes, a, b, t)
			instr++
		}
	}
}

func testCMP(instruction uint16, inputRegs, wantRegs [4]uint16, compareA, compareB int, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setMemLoc(c, 0x0000, instruction)
	for i, r := range inputRegs {
		setReg(c, i, r)
	}
	c.SetIAR(0x0000)
	runInstr(c)
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
	checkFlagsRegister(c, false,
		inputRegs[compareA] > inputRegs[compareB],
		inputRegs[compareA] == inputRegs[compareB],
		false, t)
}

func TestALUShiftLeft(t *testing.T) {
	ones := [4]uint16{0x0001, 0x0001, 0x0001, 0x0001}
	for shifts := uint16(0); shifts < 16; shifts++ {
		testShift(0x00A0, ones, [4]uint16{1 << shifts, 0x0001, 0x0001, 0x0001}, shifts, t)
		testShift(0x00A5, ones, [4]uint16{0x0001, 1 << shifts, 0x0001, 0x0001}, shifts, t)
		testShift(0x00AA, ones, [4]uint16{0x0001, 0x0001, 1 << shifts, 0x0001}, shifts, t)
		testShift(0x00AF, ones, [4]uint16{0x0001, 0x0001, 0x0001, 1 << shifts}, shifts, t)
	}
}

func TestALUShiftRight(t *testing.T) {
	in := [4]uint16{0x8000, 0x8000, 0x8000, 0x8000}
	for shifts := uint16(0); shifts < 16; shifts++ {
		testShift(0x0090, in, [4]uint16{0x8000 >> shifts, 0x8000, 0x8000, 0x8000}, shifts, t)
		testShift(0x0095, in, [4]uint16{0x8000, 0x8000 >> shifts, 0x8000, 0x8000}, shifts, t)
		testShift(0x009A, in, [4]uint16{0x8000, 0x8000, 0x8000 >> shifts, 0x8000}, shifts, t)
		testShift(0x009F, in, [4]uint16{0x8000, 0x8000, 0x8000, 0x8000 >> shifts}, shifts, t)
	}
}

func testShift(instruction uint16, inputRegs, wantRegs [4]uint16, shifts uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	for i := uint16(0); i < shifts; i++ {
		setMemLoc(c, i, instruction)
	}
	for i, r := range inputRegs {
		setReg(c, i, r)
	}
	c.SetIAR(0x0000)
	for i := uint16(0); i < shifts; i++ {
		runInstr(c)
	}
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
}

// ---- Multi-instruction programs ----

func TestSubtract(t *testing.T) {
	testSubtract(0, 0, t)
	testSubtract(1, 0, t)
	testSubtract(37, 21, t)
	testSubtract(0x00FF, 0x00FF, t)
	testSubtract(10, 3, t)
	testSubtract(100, 99, t)
}

func testSubtract(a, b uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	setRegisters(c, [4]uint16{a, b, 1, 0})
	setMemLoc(c, 0x0000, 0x00B5) // NOT R1
	setMemLoc(c, 0x0001, 0x0089) // ADD R2, R1
	setMemLoc(c, 0x0002, 0x0060) // CLF
	setMemLoc(c, 0x0003, 0x0081) // ADD R0, R1
	c.SetIAR(0x0000)
	runInstr(c)
	runInstr(c)
	runInstr(c)
	runInstr(c)
	checkRegister(c, 1, a-b, t)
}

func TestMultiply(t *testing.T) {
	testMultiply(0, 0, t)
	testMultiply(1, 1, t)
	testMultiply(1, 2, t)
	testMultiply(2, 1, t)
	testMultiply(5, 5, t)
	testMultiply(8, 12, t)
	testMultiply(19, 13, t)
}

func testMultiply(a, b uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()

	setMemLoc(c, 50, 0x0023) // DATA R3
	setMemLoc(c, 51, 0x0001) // .. 1
	setMemLoc(c, 52, 0x00EA) // XOR R2, R2
	setMemLoc(c, 53, 0x0060) // CLF
	setMemLoc(c, 54, 0x0090) // SHR R0
	setMemLoc(c, 55, 0x0058) // JC
	setMemLoc(c, 56, 59)
	setMemLoc(c, 57, 0x0040) // JMP
	setMemLoc(c, 58, 61)
	setMemLoc(c, 59, 0x0060) // CLF
	setMemLoc(c, 60, 0x0086) // ADD R1, R2
	setMemLoc(c, 61, 0x0060) // CLF
	setMemLoc(c, 62, 0x00A5) // SHL R1
	setMemLoc(c, 63, 0x00AF) // SHL R3
	setMemLoc(c, 64, 0x0058) // JC
	setMemLoc(c, 65, 68)
	setMemLoc(c, 66, 0x0040) // JMP
	setMemLoc(c, 67, 53)

	setRegisters(c, [4]uint16{a, b, 0, 0})
	c.SetIAR(50)

	for {
		runInstr(c)
		if c.iar.Value() >= 68 {
			break
		}
	}

	checkRegister(c, 2, a*b, t)
}

// ---- IO ----

func TestIOInputInstruction(t *testing.T) {
	zeros := [4]uint16{0, 0, 0, 0}
	testIOInput(0x0070, zeros, [4]uint16{0x00DA, 0x0000, 0x0000, 0x0000}, t)
	testIOInput(0x0071, zeros, [4]uint16{0x0000, 0x00DA, 0x0000, 0x0000}, t)
	testIOInput(0x0072, zeros, [4]uint16{0x0000, 0x0000, 0x00DA, 0x0000}, t)
	testIOInput(0x0073, zeros, [4]uint16{0x0000, 0x0000, 0x0000, 0x00DA}, t)
	testIOInput(0x0074, zeros, [4]uint16{0x00AD, 0x0000, 0x0000, 0x0000}, t)
	testIOInput(0x0075, zeros, [4]uint16{0x0000, 0x00AD, 0x0000, 0x0000}, t)
	testIOInput(0x0076, zeros, [4]uint16{0x0000, 0x0000, 0x00AD, 0x0000}, t)
	testIOInput(0x0077, zeros, [4]uint16{0x0000, 0x0000, 0x0000, 0x00AD}, t)
}

func testIOInput(instruction uint16, inputRegs, wantRegs [4]uint16, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	c.ConnectPeripheral(NewDumbPeripheral())
	setMemLoc(c, 0x0000, instruction)
	setRegisters(c, inputRegs)
	c.SetIAR(0x0000)
	runInstr(c)
	checkRegisters(c, wantRegs[0], wantRegs[1], wantRegs[2], wantRegs[3], t)
}

func TestIOOutputInstruction(t *testing.T) {
	testIOOutput(0x0078, [4]uint16{0x00DD, 0x0008, 0x0007, 0x0006}, 0x00DD, true, false, t)
	testIOOutput(0x0079, [4]uint16{0x0009, 0x00DD, 0x0007, 0x0006}, 0x00DD, true, false, t)
	testIOOutput(0x007A, [4]uint16{0x0009, 0x0008, 0x00DD, 0x0006}, 0x00DD, true, false, t)
	testIOOutput(0x007B, [4]uint16{0x0009, 0x0008, 0x0007, 0x00DD}, 0x00DD, true, false, t)

	testIOOutput(0x007C, [4]uint16{0x00AA, 0x0008, 0x0007, 0x0006}, 0x00AA, false, true, t)
	testIOOutput(0x007D, [4]uint16{0x0009, 0x00AA, 0x0007, 0x0006}, 0x00AA, false, true, t)
	testIOOutput(0x007E, [4]uint16{0x0009, 0x0008, 0x00AA, 0x0006}, 0x00AA, false, true, t)
	testIOOutput(0x007F, [4]uint16{0x0009, 0x0008, 0x0007, 0x00AA}, 0x00AA, false, true, t)
}

func testIOOutput(instruction uint16, inputRegs [4]uint16, wantVal uint16, wantData, wantAddr bool, t *testing.T) {
	t.Helper()
	c, _, _ := newTestCPU()
	p := NewDumbPeripheral()
	c.ConnectPeripheral(p)
	setMemLoc(c, 0x0000, instruction)
	setRegisters(c, inputRegs)
	c.SetIAR(0x0000)
	runInstr(c)
	if p.value.Value() != wantVal {
		t.Fatalf("peripheral value: got 0x%04X, want 0x%04X", p.value.Value(), wantVal)
	}
	if p.outputDataMode != wantData {
		t.Fatalf("peripheral data mode: got %v, want %v", p.outputDataMode, wantData)
	}
	if p.outputAddressMode != wantAddr {
		t.Fatalf("peripheral addr mode: got %v, want %v", p.outputAddressMode, wantAddr)
	}
}
