package asm

import (
	"fmt"
	"testing"
)

// noLabel and noSymbol are dummy resolvers for instructions that don't use them.
var noLabel LabelResolver = func(LABEL) (uint16, error) { return 0, nil }
var noSymbol SymbolResolver = func(SYMBOL) (uint16, error) { return 0, nil }

// TestTwoRegInstructions covers LOAD, STORE, ADD, AND, OR, XOR, CMP — all 16 register pairs each.
func TestTwoRegInstructions(t *testing.T) {
	type tcase struct {
		ins      Instruction
		expected uint16
	}

	var cases []tcase
	for a := REGISTER(0); a <= REG3; a++ {
		for b := REGISTER(0); b <= REG3; b++ {
			cases = append(cases,
				tcase{LOAD{a, b}, uint16(0x0000 + (a*4) + b)},
				tcase{STORE{a, b}, uint16(0x0010 + (a*4) + b)},
				tcase{ADD{a, b}, uint16(0x0080 + (a*4) + b)},
				tcase{AND{a, b}, uint16(0x00C0 + (a*4) + b)},
				tcase{OR{a, b}, uint16(0x00D0 + (a*4) + b)},
				tcase{XOR{a, b}, uint16(0x00E0 + (a*4) + b)},
				tcase{CMP{a, b}, uint16(0x00F0 + (a*4) + b)},
			)
		}
	}

	for _, tc := range cases {
		words, err := tc.ins.Emit(noLabel, noSymbol)
		if err != nil {
			t.Errorf("%T %v: unexpected error: %v", tc.ins, tc.ins, err)
			continue
		}
		if len(words) != 1 || words[0] != tc.expected {
			t.Errorf("%T %v: got %04X, want %04X", tc.ins, tc.ins, words, tc.expected)
		}
	}
}

// TestOneRegInstructions covers SHR, SHL, NOT, JR — all 4 registers, including stride-5 for shift/not.
func TestOneRegInstructions(t *testing.T) {
	type tcase struct {
		ins      Instruction
		expected uint16
	}

	cases := []tcase{
		{SHR{REG0}, 0x0090},
		{SHR{REG1}, 0x0095},
		{SHR{REG2}, 0x009A},
		{SHR{REG3}, 0x009F},

		{SHL{REG0}, 0x00A0},
		{SHL{REG1}, 0x00A5},
		{SHL{REG2}, 0x00AA},
		{SHL{REG3}, 0x00AF},

		{NOT{REG0}, 0x00B0},
		{NOT{REG1}, 0x00B5},
		{NOT{REG2}, 0x00BA},
		{NOT{REG3}, 0x00BF},

		{JR{REG0}, 0x0030},
		{JR{REG1}, 0x0031},
		{JR{REG2}, 0x0032},
		{JR{REG3}, 0x0033},
	}

	for _, tc := range cases {
		words, err := tc.ins.Emit(noLabel, noSymbol)
		if err != nil {
			t.Errorf("%T %v: unexpected error: %v", tc.ins, tc.ins, err)
			continue
		}
		if len(words) != 1 || words[0] != tc.expected {
			t.Errorf("%T %v: got %04X, want %04X", tc.ins, tc.ins, words, tc.expected)
		}
	}
}

// TestDATAInstruction covers NUMBER operand and SYMBOL operand.
func TestDATAInstruction(t *testing.T) {
	// NUMBER operand
	for reg := REGISTER(0); reg <= REG3; reg++ {
		ins := DATA{reg, NUMBER{0x42}}
		words, err := ins.Emit(noLabel, noSymbol)
		if err != nil {
			t.Fatalf("DATA R%d NUMBER: unexpected error: %v", reg, err)
		}
		wantOpcode := uint16(0x0020 + reg)
		if len(words) != 2 || words[0] != wantOpcode || words[1] != 0x42 {
			t.Errorf("DATA R%d NUMBER: got %04X %04X, want %04X 0042", reg, words[0], words[1], wantOpcode)
		}
	}

	// SYMBOL operand
	sr := func(s SYMBOL) (uint16, error) {
		if s.Name == "ADDR" {
			return 0x0600, nil
		}
		return 0, fmt.Errorf("unknown symbol: %s", s.Name)
	}
	words, err := DATA{REG0, SYMBOL{"ADDR"}}.Emit(noLabel, sr)
	if err != nil {
		t.Fatalf("DATA R0 SYMBOL: unexpected error: %v", err)
	}
	if len(words) != 2 || words[0] != 0x0020 || words[1] != 0x0600 {
		t.Errorf("DATA R0 SYMBOL: got %04X %04X, want 0020 0600", words[0], words[1])
	}
}

// TestIOInstructions covers IN/OUT for both modes and all 4 registers.
func TestIOInstructions(t *testing.T) {
	type tcase struct {
		ins      Instruction
		expected uint16
	}

	cases := []tcase{
		{IN{DATA_MODE, REG0}, 0x0070},
		{IN{DATA_MODE, REG1}, 0x0071},
		{IN{DATA_MODE, REG2}, 0x0072},
		{IN{DATA_MODE, REG3}, 0x0073},
		{IN{ADDRESS_MODE, REG0}, 0x0074},
		{IN{ADDRESS_MODE, REG1}, 0x0075},
		{IN{ADDRESS_MODE, REG2}, 0x0076},
		{IN{ADDRESS_MODE, REG3}, 0x0077},

		{OUT{DATA_MODE, REG0}, 0x0078},
		{OUT{DATA_MODE, REG1}, 0x0079},
		{OUT{DATA_MODE, REG2}, 0x007A},
		{OUT{DATA_MODE, REG3}, 0x007B},
		{OUT{ADDRESS_MODE, REG0}, 0x007C},
		{OUT{ADDRESS_MODE, REG1}, 0x007D},
		{OUT{ADDRESS_MODE, REG2}, 0x007E},
		{OUT{ADDRESS_MODE, REG3}, 0x007F},
	}

	for _, tc := range cases {
		words, err := tc.ins.Emit(noLabel, noSymbol)
		if err != nil {
			t.Errorf("%T %v: unexpected error: %v", tc.ins, tc.ins, err)
			continue
		}
		if len(words) != 1 || words[0] != tc.expected {
			t.Errorf("%T %v: got %04X, want %04X", tc.ins, tc.ins, words[0], tc.expected)
		}
	}
}

// TestJMPInstruction verifies JMP emits [0x0040, resolvedAddress].
func TestJMPInstruction(t *testing.T) {
	lr := func(l LABEL) (uint16, error) {
		if l.Name == "target" {
			return 0x0123, nil
		}
		return 0, fmt.Errorf("unknown label: %s", l.Name)
	}
	words, err := JMP{LABEL{"target"}}.Emit(lr, noSymbol)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 2 || words[0] != 0x0040 || words[1] != 0x0123 {
		t.Errorf("got %04X %04X, want 0040 0123", words[0], words[1])
	}
}

// TestJMPFlagInstructions covers all 15 flag combinations.
func TestJMPFlagInstructions(t *testing.T) {
	lr := func(l LABEL) (uint16, error) { return 0xABCD, nil }

	cases := []struct {
		flags    []string
		expected uint16
	}{
		{[]string{"Z"}, 0x0051},
		{[]string{"E"}, 0x0052},
		{[]string{"E", "Z"}, 0x0053},
		{[]string{"A"}, 0x0054},
		{[]string{"A", "Z"}, 0x0055},
		{[]string{"A", "E"}, 0x0056},
		{[]string{"A", "E", "Z"}, 0x0057},
		{[]string{"C"}, 0x0058},
		{[]string{"C", "Z"}, 0x0059},
		{[]string{"C", "E"}, 0x005A},
		{[]string{"C", "E", "Z"}, 0x005B},
		{[]string{"C", "A"}, 0x005C},
		{[]string{"C", "A", "Z"}, 0x005D},
		{[]string{"C", "A", "E"}, 0x005E},
		{[]string{"C", "A", "E", "Z"}, 0x005F},
	}

	for _, tc := range cases {
		ins := JMPF{tc.flags, LABEL{"loop"}}
		words, err := ins.Emit(lr, noSymbol)
		if err != nil {
			t.Errorf("JMPF %v: unexpected error: %v", tc.flags, err)
			continue
		}
		if len(words) != 2 || words[0] != tc.expected || words[1] != 0xABCD {
			t.Errorf("JMPF %v: got %04X %04X, want %04X ABCD", tc.flags, words[0], words[1], tc.expected)
		}
	}
}

// TestCLFInstruction verifies CLF emits [0x0060].
func TestCLFInstruction(t *testing.T) {
	words, err := CLF{}.Emit(noLabel, noSymbol)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 1 || words[0] != 0x0060 {
		t.Errorf("got %04X, want 0060", words)
	}
}

// TestCALLInstruction verifies CALL expands to DATA R3 + JMP.
func TestCALLInstruction(t *testing.T) {
	lr := func(l LABEL) (uint16, error) {
		if l.Name == "foo" {
			return 0x0001, nil
		}
		return 0, fmt.Errorf("unknown label: %s", l.Name)
	}
	sr := func(s SYMBOL) (uint16, error) {
		if s.Name == NEXTINSTRUCTION {
			return 0x1234, nil
		}
		return 0, fmt.Errorf("unknown symbol: %s", s.Name)
	}

	words, err := CALL{LABEL{"foo"}}.Emit(lr, sr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []uint16{0x0023, 0x1234, 0x0040, 0x0001}
	if len(words) != len(want) {
		t.Fatalf("got %d words, want %d: %04X", len(words), len(want), words)
	}
	for i, w := range want {
		if words[i] != w {
			t.Errorf("word[%d]: got %04X, want %04X", i, words[i], w)
		}
	}
}
