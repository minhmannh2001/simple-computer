package asm

import (
	"strings"
	"testing"
)

// TestBasicAssembly verifies DATA with a number literal emits two words.
func TestBasicAssembly(t *testing.T) {
	a := &Assembler{}
	out, err := a.Process(0x0500, []Instruction{
		DATA{REG0, NUMBER{0x0042}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []uint16{0x0020, 0x0042}
	if !eq16(out, want) {
		t.Errorf("got %04X, want %04X", out, want)
	}
}

// TestLabelResolution verifies JMP target resolves to the address of the instruction after the label.
func TestLabelResolution(t *testing.T) {
	a := &Assembler{}
	instructions := []Instruction{
		JMP{LABEL{"end"}},         // size=2, at 0x0500–0x0501
		DATA{REG0, NUMBER{0xDEAD}}, // size=2, at 0x0502–0x0503
		DEFLABEL{"end"},            // size=0, marks 0x0504
		DATA{REG0, NUMBER{0x0001}}, // size=2, at 0x0504–0x0505
	}
	out, err := a.Process(0x0500, instructions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "end" is at position 4 = absolute 0x0504
	want := []uint16{0x0040, 0x0504, 0x0020, 0xDEAD, 0x0020, 0x0001}
	if !eq16(out, want) {
		t.Errorf("got %04X, want %04X", out, want)
	}
}

// TestSymbolResolution verifies DEFSYMBOL is substituted in DATA.
func TestSymbolResolution(t *testing.T) {
	a := &Assembler{}
	instructions := []Instruction{
		DEFSYMBOL{"ADDR", 0x0600},
		DATA{REG0, SYMBOL{"ADDR"}},
	}
	out, err := a.Process(0x0500, instructions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []uint16{0x0020, 0x0600}
	if !eq16(out, want) {
		t.Errorf("got %04X, want %04X", out, want)
	}
}

// TestReservedSymbolRejection verifies Process rejects DEFSYMBOL with reserved names.
func TestReservedSymbolRejection(t *testing.T) {
	for _, name := range []string{NEXTINSTRUCTION, CURRENTINSTRUCTION} {
		a := &Assembler{}
		_, err := a.Process(0x0500, []Instruction{
			DEFSYMBOL{name, 0x1234},
		})
		if err == nil {
			t.Errorf("expected error for reserved symbol %q, got nil", name)
			continue
		}
		if !strings.Contains(err.Error(), "reserved") {
			t.Errorf("error for %q should mention 'reserved', got: %v", name, err)
		}
	}
}

// TestDuplicateLabelRejection verifies Process rejects duplicate label definitions.
func TestDuplicateLabelRejection(t *testing.T) {
	a := &Assembler{}
	_, err := a.Process(0x0500, []Instruction{
		DEFLABEL{"foo"},
		DEFLABEL{"foo"},
	})
	if err == nil {
		t.Fatal("expected error for duplicate label, got nil")
	}
}

// TestCurrentAndNextInstructionInjection verifies the auto-injected symbols resolve correctly.
func TestCurrentAndNextInstructionInjection(t *testing.T) {
	a := &Assembler{}
	instructions := []Instruction{
		DATA{REG0, SYMBOL{CURRENTINSTRUCTION}}, // at 0x0500, size=2
		DATA{REG1, SYMBOL{NEXTINSTRUCTION}},     // at 0x0502, size=2
	}
	out, err := a.Process(0x0500, instructions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DATA R0: CURRENTINSTRUCTION=0x0500 → [0x0020, 0x0500]
	// DATA R1: NEXTINSTRUCTION = address after DATA R1 = 0x0504 → [0x0021, 0x0504]
	want := []uint16{0x0020, 0x0500, 0x0021, 0x0504}
	if !eq16(out, want) {
		t.Errorf("got %04X, want %04X", out, want)
	}
}

// TestCALLExpansion verifies CALL emits DATA R3 + JMP and sets the correct return address.
func TestCALLExpansion(t *testing.T) {
	a := &Assembler{}
	instructions := []Instruction{
		CALL{LABEL{"myfunc"}}, // size=4, at 0x0500–0x0503; NEXTINSTRUCTION=0x0504
		DEFLABEL{"myfunc"},    // size=0, marks 0x0504
		CLF{},                 // size=1, at 0x0504
	}
	out, err := a.Process(0x0500, instructions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// myfunc = 0x0504, NEXTINSTRUCTION = 0x0504
	// CALL emits: DATA R3, 0x0504 = [0x0023, 0x0504]; JMP myfunc = [0x0040, 0x0504]
	// CLF emits: [0x0060]
	want := []uint16{0x0023, 0x0504, 0x0040, 0x0504, 0x0060}
	if !eq16(out, want) {
		t.Errorf("got %04X, want %04X", out, want)
	}
}

// eq16 compares two uint16 slices.
func eq16(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
