package asm

import (
	"fmt"
	"strings"
)

type LabelResolver func(LABEL) (uint16, error)
type SymbolResolver func(SYMBOL) (uint16, error)

type REGISTER int
type IO_MODE string

const (
	REG0 = REGISTER(iota)
	REG1
	REG2
	REG3
)

const (
	ADDRESS_MODE = IO_MODE("Addr")
	DATA_MODE    = IO_MODE("Data")
)

type Instruction interface {
	String() string
	Emit(LabelResolver, SymbolResolver) ([]uint16, error)
	Size() int
}

// --- Two-register instructions ---

type LOAD struct {
	MemoryAddressReg REGISTER
	ToRegister       REGISTER
}

func (i LOAD) Size() int { return 1 }
func (i LOAD) String() string {
	return fmt.Sprintf("LD R%d, R%d", i.MemoryAddressReg, i.ToRegister)
}
func (i LOAD) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x0000 + (i.MemoryAddressReg * 4) + i.ToRegister)}, nil
}

type STORE struct {
	FromRegister REGISTER
	ToRegister   REGISTER
}

func (i STORE) Size() int { return 1 }
func (i STORE) String() string {
	return fmt.Sprintf("ST R%d, R%d", i.FromRegister, i.ToRegister)
}
func (i STORE) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x0010 + (i.FromRegister * 4) + i.ToRegister)}, nil
}

type ADD struct {
	ARegister REGISTER
	BRegister REGISTER
}

func (i ADD) Size() int { return 1 }
func (i ADD) String() string {
	return fmt.Sprintf("ADD R%d, R%d", i.ARegister, i.BRegister)
}
func (i ADD) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x0080 + (i.ARegister * 4) + i.BRegister)}, nil
}

type AND struct {
	ARegister REGISTER
	BRegister REGISTER
}

func (i AND) Size() int { return 1 }
func (i AND) String() string {
	return fmt.Sprintf("AND R%d, R%d", i.ARegister, i.BRegister)
}
func (i AND) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00C0 + (i.ARegister * 4) + i.BRegister)}, nil
}

type OR struct {
	ARegister REGISTER
	BRegister REGISTER
}

func (i OR) Size() int { return 1 }
func (i OR) String() string {
	return fmt.Sprintf("OR R%d, R%d", i.ARegister, i.BRegister)
}
func (i OR) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00D0 + (i.ARegister * 4) + i.BRegister)}, nil
}

type XOR struct {
	ARegister REGISTER
	BRegister REGISTER
}

func (i XOR) Size() int { return 1 }
func (i XOR) String() string {
	return fmt.Sprintf("XOR R%d, R%d", i.ARegister, i.BRegister)
}
func (i XOR) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00E0 + (i.ARegister * 4) + i.BRegister)}, nil
}

type CMP struct {
	ARegister REGISTER
	BRegister REGISTER
}

func (i CMP) Size() int { return 1 }
func (i CMP) String() string {
	return fmt.Sprintf("CMP R%d, R%d", i.ARegister, i.BRegister)
}
func (i CMP) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00F0 + (i.ARegister * 4) + i.BRegister)}, nil
}

// --- Single-register instructions ---

type JR struct{ Register REGISTER }

func (i JR) Size() int                                           { return 1 }
func (i JR) String() string                                      { return fmt.Sprintf("JR R%d", i.Register) }
func (i JR) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x0030 + i.Register)}, nil
}

type SHR struct{ Register REGISTER }

func (i SHR) Size() int { return 1 }
func (i SHR) String() string { return fmt.Sprintf("SHR R%d", i.Register) }
func (i SHR) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x0090 + i.Register*5)}, nil
}

type SHL struct{ Register REGISTER }

func (i SHL) Size() int { return 1 }
func (i SHL) String() string { return fmt.Sprintf("SHL R%d", i.Register) }
func (i SHL) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00A0 + i.Register*5)}, nil
}

type NOT struct{ Register REGISTER }

func (i NOT) Size() int { return 1 }
func (i NOT) String() string { return fmt.Sprintf("NOT R%d", i.Register) }
func (i NOT) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	return []uint16{uint16(0x00B0 + i.Register*5)}, nil
}

// --- DATA instruction ---

type DATA struct {
	ToRegister REGISTER
	Value      marker
}

func (i DATA) Size() int { return 2 }
func (i DATA) String() string {
	return fmt.Sprintf("DATA R%d, %s", i.ToRegister, i.Value)
}
func (i DATA) Emit(_ LabelResolver, sr SymbolResolver) ([]uint16, error) {
	opcode := uint16(0x0020 + i.ToRegister)
	switch m := i.Value.(type) {
	case NUMBER:
		return []uint16{opcode, m.Value}, nil
	case SYMBOL:
		val, err := sr(m)
		if err != nil {
			return nil, err
		}
		return []uint16{opcode, val}, nil
	default:
		return nil, fmt.Errorf("DATA: unsupported marker type %T", i.Value)
	}
}

// --- Jump instructions ---

type JMP struct{ JumpLoc LABEL }

func (i JMP) Size() int { return 2 }
func (i JMP) String() string { return fmt.Sprintf("JMP %s", i.JumpLoc.Name) }
func (i JMP) Emit(lr LabelResolver, _ SymbolResolver) ([]uint16, error) {
	addr, err := lr(i.JumpLoc)
	if err != nil {
		return nil, err
	}
	return []uint16{0x0040, addr}, nil
}

// jmpfOpcodes maps flag combinations (in CAEZ order) to opcodes.
var jmpfOpcodes = map[string]uint16{
	"Z":    0x0051,
	"E":    0x0052,
	"EZ":   0x0053,
	"A":    0x0054,
	"AZ":   0x0055,
	"AE":   0x0056,
	"AEZ":  0x0057,
	"C":    0x0058,
	"CZ":   0x0059,
	"CE":   0x005A,
	"CEZ":  0x005B,
	"CA":   0x005C,
	"CAZ":  0x005D,
	"CAE":  0x005E,
	"CAEZ": 0x005F,
}

type JMPF struct {
	Flags   []string
	JumpLoc LABEL
}

func (i JMPF) Size() int { return 2 }
func (i JMPF) String() string {
	return fmt.Sprintf("JMP%s %s", strings.Join(i.Flags, ""), i.JumpLoc.Name)
}
func (i JMPF) Emit(lr LabelResolver, _ SymbolResolver) ([]uint16, error) {
	key := strings.Join(i.Flags, "")
	opcode, ok := jmpfOpcodes[key]
	if !ok {
		return nil, fmt.Errorf("unknown flag combination: %q", key)
	}
	addr, err := lr(i.JumpLoc)
	if err != nil {
		return nil, err
	}
	return []uint16{opcode, addr}, nil
}

// --- CLF ---

type CLF struct{}

func (CLF) Size() int                                                { return 1 }
func (CLF) String() string                                           { return "CLF" }
func (CLF) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) { return []uint16{0x0060}, nil }

// --- I/O instructions ---

type IN struct {
	IoMode     IO_MODE
	ToRegister REGISTER
}

func (i IN) Size() int { return 1 }
func (i IN) String() string { return fmt.Sprintf("IN %s, R%d", i.IoMode, i.ToRegister) }
func (i IN) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	base := uint16(0x0070)
	if i.IoMode == ADDRESS_MODE {
		base = 0x0074
	}
	return []uint16{base + uint16(i.ToRegister)}, nil
}

type OUT struct {
	IoMode       IO_MODE
	FromRegister REGISTER
}

func (i OUT) Size() int { return 1 }
func (i OUT) String() string { return fmt.Sprintf("OUT %s, R%d", i.IoMode, i.FromRegister) }
func (i OUT) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) {
	base := uint16(0x0078)
	if i.IoMode == ADDRESS_MODE {
		base = 0x007C
	}
	return []uint16{base + uint16(i.FromRegister)}, nil
}

// --- CALL pseudo-instruction ---

// CALL expands to DATA R3, <NEXTINSTRUCTION> + JMP <Routine>.
// NEXTINSTRUCTION is injected into the symbol map by Assembler.Process before each Emit call.
type CALL struct{ Routine LABEL }

func (c CALL) Size() int { return 4 }
func (c CALL) String() string { return fmt.Sprintf("CALL %s", c.Routine.Name) }
func (c CALL) Emit(lr LabelResolver, sr SymbolResolver) ([]uint16, error) {
	next, err := sr(SYMBOL{NEXTINSTRUCTION})
	if err != nil {
		return nil, fmt.Errorf("CALL: cannot resolve return address: %w", err)
	}
	dataWords, err := DATA{REG3, NUMBER{next}}.Emit(lr, sr)
	if err != nil {
		return nil, err
	}
	jmpWords, err := JMP{c.Routine}.Emit(lr, sr)
	if err != nil {
		return nil, err
	}
	return append(dataWords, jmpWords...), nil
}

// --- Placeholder instructions (Size=0) ---

type DEFLABEL struct{ Name string }

func (DEFLABEL) Size() int { return 0 }
func (d DEFLABEL) String() string { return d.Name }
func (DEFLABEL) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) { return nil, nil }

type DEFSYMBOL struct {
	Name  string
	Value uint16
}

func (DEFSYMBOL) Size() int { return 0 }
func (d DEFSYMBOL) String() string { return fmt.Sprintf("%%%s = 0x%X", d.Name, d.Value) }
func (DEFSYMBOL) Emit(_ LabelResolver, _ SymbolResolver) ([]uint16, error) { return nil, nil }

// --- Instructions helper ---

type Instructions struct {
	instructions []Instruction
}

func (s *Instructions) Add(ins ...Instruction) {
	s.instructions = append(s.instructions, ins...)
}

func (s *Instructions) AddBlocks(blocks ...[]Instruction) {
	for _, block := range blocks {
		s.instructions = append(s.instructions, block...)
	}
}

func (s *Instructions) Get() []Instruction {
	return s.instructions
}

func (s *Instructions) String() string {
	parts := make([]string, len(s.instructions))
	for i, ins := range s.instructions {
		parts[i] = ins.String()
	}
	return strings.Join(parts, "\n")
}
