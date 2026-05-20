package alu

import (
	"testing"

	"simple-computer/components"
)

func newTestALU() (*ALU, *components.Bus, *components.Bus, *components.Bus, *components.Bus) {
	inputA := components.NewBus(components.BUS_WIDTH)
	inputB := components.NewBus(components.BUS_WIDTH)
	output := components.NewBus(components.BUS_WIDTH)
	flags := components.NewBus(4)
	return NewALU(inputA, inputB, output, flags), inputA, inputB, output, flags
}

func setOp(a *ALU, op int) {
	a.Op[0].Update(op&1 != 0)
	a.Op[1].Update(op&2 != 0)
	a.Op[2].Update(op&4 != 0)
}

func busValue(b *components.Bus) uint16 {
	var v uint16
	for i := range components.BUS_WIDTH {
		if b.GetOutputWire(i) {
			v |= 1 << uint(components.BUS_WIDTH-1-i)
		}
	}
	return v
}

// --- ADD ---

func TestALU_ADD_Simple(t *testing.T) {
	a, inputA, inputB, output, flags := newTestALU()
	inputA.SetValue(3)
	inputB.SetValue(5)
	setOp(a, ADD)
	a.Update()

	if got := busValue(output); got != 8 {
		t.Errorf("3+5: got %#x, want 0x0008", got)
	}
	if flags.GetOutputWire(0) {
		t.Error("3+5: carry should be false")
	}
	if flags.GetOutputWire(3) {
		t.Error("3+5: zero should be false")
	}
}

func TestALU_ADD_Overflow(t *testing.T) {
	a, inputA, inputB, output, flags := newTestALU()
	inputA.SetValue(0xFFFF)
	inputB.SetValue(1)
	setOp(a, ADD)
	a.Update()

	if got := busValue(output); got != 0 {
		t.Errorf("0xFFFF+1 overflow: got %#x, want 0x0000", got)
	}
	if !flags.GetOutputWire(0) {
		t.Error("0xFFFF+1: carry should be true")
	}
}

// --- SHR ---

func TestALU_SHR_Simple(t *testing.T) {
	a, inputA, _, output, flags := newTestALU()
	inputA.SetValue(0x0002)
	setOp(a, SHR)
	a.Update()

	if got := busValue(output); got != 0x0001 {
		t.Errorf("SHR 0x0002: got %#x, want 0x0001", got)
	}
	if flags.GetOutputWire(0) {
		t.Error("SHR 0x0002: carry should be false")
	}
}

func TestALU_SHR_CarryOut(t *testing.T) {
	a, inputA, _, output, flags := newTestALU()
	inputA.SetValue(0x0001)
	setOp(a, SHR)
	a.Update()

	if got := busValue(output); got != 0x0000 {
		t.Errorf("SHR 0x0001: got %#x, want 0x0000", got)
	}
	if !flags.GetOutputWire(0) {
		t.Error("SHR 0x0001: carry should be true")
	}
}

// --- SHL ---

func TestALU_SHL_Simple(t *testing.T) {
	a, inputA, _, output, flags := newTestALU()
	inputA.SetValue(0x0001)
	setOp(a, SHL)
	a.Update()

	if got := busValue(output); got != 0x0002 {
		t.Errorf("SHL 0x0001: got %#x, want 0x0002", got)
	}
	if flags.GetOutputWire(0) {
		t.Error("SHL 0x0001: carry should be false")
	}
}

func TestALU_SHL_CarryOut(t *testing.T) {
	a, inputA, _, output, flags := newTestALU()
	inputA.SetValue(0x8000)
	setOp(a, SHL)
	a.Update()

	if got := busValue(output); got != 0x0000 {
		t.Errorf("SHL 0x8000: got %#x, want 0x0000", got)
	}
	if !flags.GetOutputWire(0) {
		t.Error("SHL 0x8000: carry should be true")
	}
}

// --- NOT ---

func TestALU_NOT_Zero(t *testing.T) {
	a, inputA, _, output, _ := newTestALU()
	inputA.SetValue(0x0000)
	setOp(a, NOT)
	a.Update()

	if got := busValue(output); got != 0xFFFF {
		t.Errorf("NOT 0x0000: got %#x, want 0xFFFF", got)
	}
}

func TestALU_NOT_AllOnes(t *testing.T) {
	a, inputA, _, output, _ := newTestALU()
	inputA.SetValue(0xFFFF)
	setOp(a, NOT)
	a.Update()

	if got := busValue(output); got != 0x0000 {
		t.Errorf("NOT 0xFFFF: got %#x, want 0x0000", got)
	}
}

// --- AND ---

func TestALU_AND(t *testing.T) {
	a, inputA, inputB, output, _ := newTestALU()
	inputA.SetValue(0xFF00)
	inputB.SetValue(0x0FF0)
	setOp(a, AND)
	a.Update()

	if got := busValue(output); got != 0x0F00 {
		t.Errorf("AND: got %#x, want 0x0F00", got)
	}
}

// --- OR ---

func TestALU_OR(t *testing.T) {
	a, inputA, inputB, output, _ := newTestALU()
	inputA.SetValue(0xFF00)
	inputB.SetValue(0x00FF)
	setOp(a, OR)
	a.Update()

	if got := busValue(output); got != 0xFFFF {
		t.Errorf("OR: got %#x, want 0xFFFF", got)
	}
}

// --- XOR ---

func TestALU_XOR_ZeroResult(t *testing.T) {
	a, inputA, inputB, output, flags := newTestALU()
	inputA.SetValue(0xFFFF)
	inputB.SetValue(0xFFFF)
	setOp(a, XOR)
	a.Update()

	if got := busValue(output); got != 0x0000 {
		t.Errorf("XOR same: got %#x, want 0x0000", got)
	}
	if !flags.GetOutputWire(3) {
		t.Error("XOR same values: zero flag should be true")
	}
}

// --- CMP ---

func TestALU_CMP_Equal(t *testing.T) {
	a, inputA, inputB, output, flags := newTestALU()
	inputA.SetValue(0x1234)
	inputB.SetValue(0x1234)
	setOp(a, CMP)
	a.Update()

	if !flags.GetOutputWire(2) {
		t.Error("CMP equal: isEqual flag should be true")
	}
	if got := busValue(output); got != 0 {
		t.Errorf("CMP output should be 0, got %#x", got)
	}
}

func TestALU_CMP_ALarger(t *testing.T) {
	a, inputA, inputB, _, flags := newTestALU()
	inputA.SetValue(0x0010)
	inputB.SetValue(0x0001)
	setOp(a, CMP)
	a.Update()

	if !flags.GetOutputWire(1) {
		t.Error("CMP A>B: aIsLarger flag should be true")
	}
	if flags.GetOutputWire(2) {
		t.Error("CMP A>B: isEqual flag should be false")
	}
}

func TestALU_CMP_BLarger(t *testing.T) {
	a, inputA, inputB, _, flags := newTestALU()
	inputA.SetValue(0x0001)
	inputB.SetValue(0x0010)
	setOp(a, CMP)
	a.Update()

	if flags.GetOutputWire(1) {
		t.Error("CMP A<B: aIsLarger flag should be false")
	}
	if flags.GetOutputWire(2) {
		t.Error("CMP A<B: isEqual flag should be false")
	}
}

// --- Zero flag ---

func TestALU_ZeroFlagSet(t *testing.T) {
	a, inputA, inputB, _, flags := newTestALU()
	inputA.SetValue(0)
	inputB.SetValue(0)
	setOp(a, ADD)
	a.Update()

	if !flags.GetOutputWire(3) {
		t.Error("0+0=0: zero flag should be true")
	}
}

func TestALU_ZeroFlagClear(t *testing.T) {
	a, inputA, inputB, _, flags := newTestALU()
	inputA.SetValue(1)
	inputB.SetValue(1)
	setOp(a, ADD)
	a.Update()

	if flags.GetOutputWire(3) {
		t.Error("1+1=2: zero flag should be false")
	}
}
