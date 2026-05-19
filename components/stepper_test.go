package components

import "testing"

func TestStepperInitialState(t *testing.T) {
	s := NewStepper()
	if !s.GetOutputWire(0) {
		t.Error("new Stepper should have step 0 active")
	}
	for i := 1; i <= 5; i++ {
		if s.GetOutputWire(i) {
			t.Errorf("step %d should be inactive initially", i)
		}
	}
}

func TestStepperFirstCycle(t *testing.T) {
	s := NewStepper()
	s.Update(true)
	s.Update(false)
	if s.GetOutputWire(0) {
		t.Error("step 0 should be inactive after one cycle")
	}
	if !s.GetOutputWire(1) {
		t.Error("step 1 should be active after one cycle")
	}
}

func TestStepperIntermediateSteps(t *testing.T) {
	s := NewStepper()
	for expected := 1; expected <= 5; expected++ {
		s.Update(true)
		s.Update(false)
		if !s.GetOutputWire(expected) {
			t.Errorf("after %d cycle(s), step %d should be active", expected, expected)
		}
	}
	s.Update(true)
	s.Update(false)
	if !s.GetOutputWire(0) {
		t.Error("after 6 cycles, step 0 should be active again")
	}
}

func TestStepperFullCycle(t *testing.T) {
	s := NewStepper()
	for range 6 {
		s.Update(true)
		s.Update(false)
	}
	if !s.GetOutputWire(0) {
		t.Error("stepper should return to step 0 after 6 full cycles")
	}
}

func TestStepperExactlyOneActive(t *testing.T) {
	s := NewStepper()
	checkOne := func(cycle int) {
		active := 0
		for i := range 6 {
			if s.GetOutputWire(i) {
				active++
			}
		}
		if active != 1 {
			t.Errorf("cycle %d: expected exactly 1 active step, got %d", cycle, active)
		}
	}
	checkOne(0)
	for cycle := 1; cycle <= 6; cycle++ {
		s.Update(true)
		checkOne(cycle)
		s.Update(false)
		checkOne(cycle)
	}
}

func TestStepperString(t *testing.T) {
	s := NewStepper()
	if got := s.String(); got != "* - - - - -" {
		t.Errorf("String() at step 0 = %q, want %q", got, "* - - - - -")
	}
	s.Update(true)
	s.Update(false)
	if got := s.String(); got != "- * - - - -" {
		t.Errorf("String() at step 1 = %q, want %q", got, "- * - - - -")
	}
}
