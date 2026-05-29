package computer

import (
	"testing"
	"time"
)

func newTestComputer() *SimpleComputer {
	sc := make(chan *[160][240]byte, 1)
	qc := make(chan bool, 2) // buffered 2: one for Run(), one for screenControl.Run()
	return NewComputer(sc, qc)
}

func runNSteps(comp *SimpleComputer, n int) {
	for i := 0; i < n; i++ {
		comp.cpu.Step()
	}
}

// ---- LoadToRAM ----

func TestLoadToRAM_PanicsOnLowAddress(t *testing.T) {
	comp := newTestComputer()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for reserved address 0x0000")
		}
	}()
	comp.LoadToRAM(0x0000, []uint16{1})
}

func TestLoadToRAM_PanicsOnHighAddress(t *testing.T) {
	comp := newTestComputer()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for reserved address 0xFF00")
		}
	}()
	comp.LoadToRAM(0xFF00, []uint16{1})
}

func TestLoadToRAM_SucceedsAtCodeRegionStart(t *testing.T) {
	comp := newTestComputer()
	comp.LoadToRAM(0x0500, []uint16{0x0020, 0x0042})
}

// ---- DATA instruction end-to-end ----

func TestDataInstruction_EndToEnd(t *testing.T) {
	comp := newTestComputer()
	comp.LoadToRAM(0x0500, []uint16{0x0020, 0x0042}) // DATA R0, 0x0042
	comp.cpu.SetIAR(0x0500)

	runNSteps(comp, 6) // 1 instruction = 6 Step() calls

	if got := comp.cpu.GPReg(0); got != 0x0042 {
		t.Fatalf("R0: got 0x%04X, want 0x0042", got)
	}
}

// ---- ADD then CMP sets Equal flag ----

func TestAddThenCMP_EqualFlagSet(t *testing.T) {
	comp := newTestComputer()

	// DATA R0, 3  → R0 = 3
	// DATA R1, 5  → R1 = 5
	// ADD R0, R1  → R1 = R0+R1 = 8
	// DATA R2, 8  → R2 = 8
	// CMP R1, R2  → equal flag (R1==R2==8)
	program := []uint16{
		0x0020, 0x0003, // DATA R0, 3
		0x0021, 0x0005, // DATA R1, 5
		0x0081,         // ADD R0, R1
		0x0022, 0x0008, // DATA R2, 8
		0x00F6,         // CMP R1, R2
	}
	comp.LoadToRAM(0x0500, program)
	comp.cpu.SetIAR(0x0500)

	runNSteps(comp, 5*6) // 5 instructions × 6 steps each

	if !comp.cpu.EqualFlag() {
		t.Fatal("equal flag should be set after CMP R1, R2 where R1 == R2 == 8")
	}
}

// ---- Run() smoke test ----

func TestRun_StartsAndCancels(t *testing.T) {
	comp := newTestComputer()
	comp.LoadToRAM(0x0500, []uint16{0x0020, 0x0042}) // DATA R0, 0x0042

	done := make(chan struct{})
	go func() {
		defer close(done)
		comp.Run(time.Tick(1*time.Nanosecond), PrintStateConfig{})
	}()

	time.Sleep(5 * time.Millisecond)
	comp.quitChannel <- true
	comp.quitChannel <- true // stop screenControl goroutine too

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run() did not stop after quit signal")
	}
}
