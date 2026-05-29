package asm

import (
	"fmt"
	"strings"
)

const (
	CURRENTINSTRUCTION = "CURRENTINSTRUCTION"
	NEXTINSTRUCTION    = "NEXTINSTRUCTION"
)

type Assembler struct {
	labels  map[string]uint16
	symbols map[string]uint16
}

func (a *Assembler) ResolveLabel(label LABEL) (uint16, error) {
	addr, ok := a.labels[label.Name]
	if !ok {
		return 0, fmt.Errorf("undefined label: %s", label.Name)
	}
	return addr, nil
}

func (a *Assembler) ResolveSymbol(symbol SYMBOL) (uint16, error) {
	val, ok := a.symbols[symbol.Name]
	if !ok {
		return 0, fmt.Errorf("undefined symbol: %s", symbol.Name)
	}
	return val, nil
}

// Process runs two passes over instructions and returns the assembled binary.
func (a *Assembler) Process(codeStartOffset uint16, instructions []Instruction) ([]uint16, error) {
	a.labels = make(map[string]uint16)
	a.symbols = make(map[string]uint16)

	// Pass 1: collect label addresses and symbol values.
	position := uint16(0)
	for _, ins := range instructions {
		switch v := ins.(type) {
		case DEFLABEL:
			if _, exists := a.labels[v.Name]; exists {
				return nil, fmt.Errorf("duplicate label: %s", v.Name)
			}
			a.labels[v.Name] = position + codeStartOffset
		case DEFSYMBOL:
			if v.Name == CURRENTINSTRUCTION || v.Name == NEXTINSTRUCTION {
				return nil, fmt.Errorf("cannot define reserved symbol: %s", v.Name)
			}
			if _, exists := a.symbols[v.Name]; exists {
				return nil, fmt.Errorf("duplicate symbol: %s", v.Name)
			}
			a.symbols[v.Name] = v.Value
		}
		position += uint16(ins.Size())
	}

	// Pass 2: emit opcodes with resolved labels and symbols.
	var output []uint16
	position = 0
	for i, ins := range instructions {
		switch ins.(type) {
		case DEFLABEL, DEFSYMBOL:
			continue
		}

		currentOffset := position + codeStartOffset
		a.symbols[CURRENTINSTRUCTION] = currentOffset
		a.symbols[NEXTINSTRUCTION] = getNextExecutableInstructionLoc(currentOffset, i, instructions)

		words, err := ins.Emit(a.ResolveLabel, a.ResolveSymbol)
		if err != nil {
			return nil, err
		}
		output = append(output, words...)
		position += uint16(ins.Size())
	}

	return output, nil
}

// ToString returns a human-readable hex dump of the assembled binary.
func (a *Assembler) ToString(codeStartOffset uint16, instructions []Instruction) (string, error) {
	words, err := a.Process(codeStartOffset, instructions)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for i, w := range words {
		fmt.Fprintf(&sb, "0x%04X: 0x%04X\n", uint16(i)+codeStartOffset, w)
	}
	return sb.String(), nil
}

// getNextExecutableInstructionLoc returns the absolute address of the instruction
// that follows instructions[currentIndex], skipping any DEFLABEL or DEFSYMBOL entries.
func getNextExecutableInstructionLoc(currentOffset uint16, currentIndex int, instructions []Instruction) uint16 {
	size := uint16(instructions[currentIndex].Size())
	if currentIndex == len(instructions)-1 {
		return currentOffset + size
	}
	for j := currentIndex + 1; j < len(instructions); j++ {
		switch instructions[j].(type) {
		case DEFLABEL, DEFSYMBOL:
			continue
		default:
			return currentOffset + size
		}
	}
	return currentOffset + size
}
