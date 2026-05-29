package asm

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	isDeflabelRe  = regexp.MustCompile(`^[A-Za-z0-9-]+:$`)
	isDefsymbolRe = regexp.MustCompile(`^%([A-Za-z0-9-]+)\s*=\s*((0x)?[0-9a-fA-F]+)$`)
	twoRegRe      = regexp.MustCompile(`^R(\d),\s*R(\d)$`)
	oneRegRe      = regexp.MustCompile(`^R(\d)$`)
	dataRe        = regexp.MustCompile(`^R(\d),\s*((0x)?[0-9a-fA-F]+|(%)([A-Za-z0-9-]+))$`)
	ioRe          = regexp.MustCompile(`^(Addr|Data),\s*R(\d)$`)
	labelRe       = regexp.MustCompile(`^([A-Za-z0-9-]+)$`)
)

type Parser struct{}

func (p *Parser) Parse(input io.Reader) ([]Instruction, error) {
	scanner := bufio.NewScanner(input)
	var result []Instruction
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		ins, err := p.parseLine(line)
		if err != nil {
			return nil, err
		}
		result = append(result, ins)
	}
	return result, scanner.Err()
}

func (p *Parser) parseLine(line string) (Instruction, error) {
	if isDeflabelRe.MatchString(line) {
		return DEFLABEL{Name: strings.TrimSuffix(line, ":")}, nil
	}
	if m := isDefsymbolRe.FindStringSubmatch(line); m != nil {
		val, err := parseNumber(m[2])
		if err != nil {
			return nil, err
		}
		return DEFSYMBOL{Name: m[1], Value: val}, nil
	}
	return p.parseInstruction(line)
}

func (p *Parser) parseInstruction(line string) (Instruction, error) {
	parts := strings.SplitN(line, " ", 2)
	mnemonic := parts[0]
	operands := ""
	if len(parts) > 1 {
		operands = strings.TrimSpace(parts[1])
	}

	switch mnemonic {
	case "CLF":
		return CLF{}, nil
	case "DATA":
		return parseDataInstruction(operands)
	case "JR", "SHR", "SHL", "NOT":
		return parseOneRegInstruction(mnemonic, operands)
	case "ADD", "AND", "OR", "XOR", "CMP", "LD", "ST":
		return parseTwoRegInstruction(mnemonic, operands)
	case "IN", "OUT":
		return parseIOInstruction(mnemonic, operands)
	case "JMP", "CALL":
		return parseLabelledJump(mnemonic, operands)
	default:
		if strings.HasPrefix(mnemonic, "JMP") {
			return parseLabelledJump(mnemonic, operands)
		}
		return nil, fmt.Errorf("unknown mnemonic: %s", mnemonic)
	}
}

func parseTwoRegInstruction(mnemonic, operands string) (Instruction, error) {
	m := twoRegRe.FindStringSubmatch(operands)
	if m == nil {
		return nil, fmt.Errorf("%s: invalid register operands: %q", mnemonic, operands)
	}
	a, err := parseRegister(m[1])
	if err != nil {
		return nil, err
	}
	b, err := parseRegister(m[2])
	if err != nil {
		return nil, err
	}
	switch mnemonic {
	case "LD":
		return LOAD{a, b}, nil
	case "ST":
		return STORE{a, b}, nil
	case "ADD":
		return ADD{a, b}, nil
	case "AND":
		return AND{a, b}, nil
	case "OR":
		return OR{a, b}, nil
	case "XOR":
		return XOR{a, b}, nil
	case "CMP":
		return CMP{a, b}, nil
	}
	return nil, fmt.Errorf("unknown two-register mnemonic: %s", mnemonic)
}

func parseOneRegInstruction(mnemonic, operands string) (Instruction, error) {
	m := oneRegRe.FindStringSubmatch(operands)
	if m == nil {
		return nil, fmt.Errorf("%s: invalid register operand: %q", mnemonic, operands)
	}
	r, err := parseRegister(m[1])
	if err != nil {
		return nil, err
	}
	switch mnemonic {
	case "JR":
		return JR{r}, nil
	case "SHR":
		return SHR{r}, nil
	case "SHL":
		return SHL{r}, nil
	case "NOT":
		return NOT{r}, nil
	}
	return nil, fmt.Errorf("unknown one-register mnemonic: %s", mnemonic)
}

func parseDataInstruction(operands string) (Instruction, error) {
	m := dataRe.FindStringSubmatch(operands)
	if m == nil {
		return nil, fmt.Errorf("DATA: invalid operands: %q", operands)
	}
	reg, err := parseRegister(m[1])
	if err != nil {
		return nil, err
	}
	if m[4] == "%" {
		return DATA{reg, SYMBOL{m[5]}}, nil
	}
	val, err := parseNumber(m[2])
	if err != nil {
		return nil, err
	}
	return DATA{reg, NUMBER{val}}, nil
}

func parseIOInstruction(mnemonic, operands string) (Instruction, error) {
	m := ioRe.FindStringSubmatch(operands)
	if m == nil {
		return nil, fmt.Errorf("%s: invalid IO operands: %q", mnemonic, operands)
	}
	mode := IO_MODE(m[1])
	reg, err := parseRegister(m[2])
	if err != nil {
		return nil, err
	}
	switch mnemonic {
	case "IN":
		return IN{mode, reg}, nil
	case "OUT":
		return OUT{mode, reg}, nil
	}
	return nil, fmt.Errorf("unknown IO mnemonic: %s", mnemonic)
}

func parseLabelledJump(mnemonic, operands string) (Instruction, error) {
	m := labelRe.FindStringSubmatch(operands)
	if m == nil {
		return nil, fmt.Errorf("%s: invalid label operand: %q", mnemonic, operands)
	}
	label := LABEL{m[1]}
	switch mnemonic {
	case "JMP":
		return JMP{label}, nil
	case "CALL":
		return CALL{label}, nil
	default:
		// JMP[CAEZ]+ — extract individual flag characters after "JMP"
		flagStr := mnemonic[3:]
		var flags []string
		for _, ch := range flagStr {
			c := string(ch)
			if c != "C" && c != "A" && c != "E" && c != "Z" {
				return nil, fmt.Errorf("invalid flag character %q in mnemonic %s", c, mnemonic)
			}
			flags = append(flags, c)
		}
		return JMPF{flags, label}, nil
	}
}

func parseRegister(s string) (REGISTER, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 3 {
		return 0, fmt.Errorf("invalid register: R%s (must be R0–R3)", s)
	}
	return REGISTER(n), nil
}

func parseNumber(s string) (uint16, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		n, err := strconv.ParseUint(s[2:], 16, 16)
		if err != nil {
			return 0, fmt.Errorf("invalid hex number: %s", s)
		}
		return uint16(n), nil
	}
	n, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal number: %s", s)
	}
	return uint16(n), nil
}
