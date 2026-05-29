package asm

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDataHexLiteral(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("DATA R0, 0x0042"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{DATA{REG0, NUMBER{0x0042}}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseLabel(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("myloop:"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{DEFLABEL{"myloop"}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseSymbol(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("%SCREEN = 0x0400"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{DEFSYMBOL{"SCREEN", 0x0400}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseMultiLineProgram(t *testing.T) {
	src := `
DATA R0, 3
DATA R1, 5
ADD R0, R1
`
	p := &Parser{}
	result, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{
		DATA{REG0, NUMBER{3}},
		DATA{REG1, NUMBER{5}},
		ADD{REG0, REG1},
	}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseLabelAndJumpRoundTrip(t *testing.T) {
	src := `
JMP end
DATA R0, 0xDEAD
end:
DATA R0, 0x0001
`
	p := &Parser{}
	result, err := p.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 instructions, got %d: %v", len(result), result)
	}
	// DEFLABEL "end" is at index 2
	dl, ok := result[2].(DEFLABEL)
	if !ok || dl.Name != "end" {
		t.Errorf("index 2: want DEFLABEL{end}, got %v", result[2])
	}
}

func TestParseConditionalJump(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("JMPCAEZ myloop"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{JMPF{[]string{"C", "A", "E", "Z"}, LABEL{"myloop"}}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseDataWithSymbol(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("DATA R2, %SCREEN"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{DATA{REG2, SYMBOL{"SCREEN"}}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseDecimalLiteral(t *testing.T) {
	p := &Parser{}
	result, err := p.Parse(strings.NewReader("DATA R0, 3"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Instruction{DATA{REG0, NUMBER{3}}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParseAllOneRegInstructions(t *testing.T) {
	cases := []struct {
		src  string
		want Instruction
	}{
		{"SHR R0", SHR{REG0}},
		{"SHL R1", SHL{REG1}},
		{"NOT R2", NOT{REG2}},
		{"JR R3", JR{REG3}},
	}
	p := &Parser{}
	for _, tc := range cases {
		result, err := p.Parse(strings.NewReader(tc.src))
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tc.src, err)
			continue
		}
		if len(result) != 1 || !reflect.DeepEqual(result[0], tc.want) {
			t.Errorf("%q: got %v, want %v", tc.src, result, tc.want)
		}
	}
}

func TestParseIOInstructions(t *testing.T) {
	cases := []struct {
		src  string
		want Instruction
	}{
		{"IN Data, R0", IN{DATA_MODE, REG0}},
		{"IN Addr, R1", IN{ADDRESS_MODE, REG1}},
		{"OUT Data, R2", OUT{DATA_MODE, REG2}},
		{"OUT Addr, R3", OUT{ADDRESS_MODE, REG3}},
	}
	p := &Parser{}
	for _, tc := range cases {
		result, err := p.Parse(strings.NewReader(tc.src))
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tc.src, err)
			continue
		}
		if len(result) != 1 || !reflect.DeepEqual(result[0], tc.want) {
			t.Errorf("%q: got %v, want %v", tc.src, result, tc.want)
		}
	}
}

// --- Error cases ---

func TestParseUnknownMnemonic(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(strings.NewReader("FOO R0, R1"))
	if err == nil {
		t.Fatal("expected error for unknown mnemonic, got nil")
	}
}

func TestParseInvalidRegister(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(strings.NewReader("SHR R5"))
	if err == nil {
		t.Fatal("expected error for invalid register R5, got nil")
	}
}

func TestParseUnsupportedIOMode(t *testing.T) {
	p := &Parser{}
	_, err := p.Parse(strings.NewReader("IN Foo, R0"))
	if err == nil {
		t.Fatal("expected error for unsupported IO mode, got nil")
	}
}
