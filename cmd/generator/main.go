package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"simple-computer/asm"
)

// RAM layout:
// 0x0000 - 0x03FF  ASCII/font table
// 0x0400           pen position
// 0x0401           keycode register
// 0x0500 - 0xFEFD  user code + memory
// 0xFEFE - 0xFEFF  sentinel JMP back to user code
// 0xFF00 - 0xFFFF  temporary variables
func main() {
	instructions := asm.Instructions{}
	instructions.AddBlocks(initialiseCommonCode())

	if len(os.Args) < 2 {
		log.Fatalf("Please provide a program name to generate")
	}

	switch os.Args[1] {
	case "ascii":
		asciiTable(instructions)
	case "brush":
		brush(instructions)
	case "text-writer":
		textWriter(instructions)
	case "me":
		me(instructions)
	default:
		log.Fatalf("unknown program: %s", os.Args[1])
	}
}

func me(instructions asm.Instructions) {
	instructions.Add(asm.DEFLABEL{"main"})
	instructions.AddBlocks(
		updatePenPosition(0x00F7),
		renderString("Daniel Harper"),
		resetLinex(),
		updatePenPosition(0x01E0),
		renderString(strings.Repeat("-", 30)),
		resetLinex(),
		updatePenPosition(0x03C1),
		renderString("djhworld.github.io"),
		resetLinex(),
		updatePenPosition(0x05A1),
		renderString("@djhworld"),
		resetLinex(),
		updatePenPosition(0x087C),
		renderString(":^)"),
		resetLinex(),
	)
	instructions.Add(
		asm.DEFLABEL{"noop"},
		asm.CLF{},
		asm.JMP{asm.LABEL{"noop"}},
	)
	fmt.Println(instructions.String())
}

func asciiTable(instructions asm.Instructions) {
	instructions.Add(
		asm.DEFLABEL{"main"},
		asm.DATA{asm.REG0, asm.NUMBER{0x0020}},
		asm.DATA{asm.REG2, asm.NUMBER{0xFF23}},
		asm.STORE{asm.REG2, asm.REG0},
		asm.DATA{asm.REG2, asm.SYMBOL{"LINEX"}},
		asm.DATA{asm.REG1, asm.NUMBER{0x0000}},
		asm.STORE{asm.REG2, asm.REG1},
	)
	instructions.AddBlocks(updatePenPosition(0x00F0))

	instructions.Add(asm.DEFLABEL{"main-loop"})
	instructions.Add(
		asm.DATA{asm.REG2, asm.NUMBER{0xFF23}},
		asm.LOAD{asm.REG2, asm.REG0},
		asm.DATA{asm.REG1, asm.SYMBOL{"ONE"}},
		asm.ADD{asm.REG1, asm.REG0},
		asm.DATA{asm.REG2, asm.SYMBOL{"KEYCODE-REGISTER"}},
		asm.STORE{asm.REG2, asm.REG0},
		asm.DATA{asm.REG2, asm.NUMBER{0xFF23}},
		asm.STORE{asm.REG2, asm.REG0},
	)
	instructions.AddBlocks(callRoutine("ROUTINE-io-drawFontCharacter"))
	instructions.Add(
		asm.DATA{asm.REG2, asm.NUMBER{0xFF23}},
		asm.LOAD{asm.REG2, asm.REG0},
		asm.DATA{asm.REG2, asm.NUMBER{0x7E}},
		asm.CMP{asm.REG0, asm.REG2},
		asm.JMPF{[]string{"E"}, asm.LABEL{"main"}},
	)
	instructions.Add(asm.JMP{asm.LABEL{"main-loop"}})

	fmt.Println(instructions.String())
}

func brush(instructions asm.Instructions) {
	instructions.Add(asm.DEFLABEL{"main"})
	instructions.AddBlocks(updatePenPosition(0x00F0))

	instructions.Add(asm.DEFLABEL{"main-getInput"})
	instructions.AddBlocks(
		callRoutine("drawBrush"),
		callRoutine("ROUTINE-io-pollKeyboard"),
		callRoutine("drawBrush"),
	)
	instructions.Add(asm.JMP{asm.LABEL{"main-getInput"}})
	instructions.AddBlocks(routine_drawBrush("drawBrush"))

	fmt.Println(instructions.String())
}

func textWriter(instructions asm.Instructions) {
	instructions.Add(asm.DEFLABEL{"main"})
	instructions.AddBlocks(updatePenPosition(0x00F0))

	instructions.Add(asm.DEFLABEL{"main-getInput"})
	instructions.AddBlocks(
		callRoutine("ROUTINE-io-pollKeyboard"),
		callRoutine("ROUTINE-io-drawFontCharacter"),
	)
	instructions.Add(asm.JMP{asm.LABEL{"main-getInput"}})

	fmt.Println(instructions.String())
}

func routine_drawBrush(labelPrefix string) []asm.Instruction {
	fontYAddr := uint16(0xFF00)

	instructions := asm.Instructions{}
	instructions.Add(asm.DEFLABEL{labelPrefix})

	instructions.Add(
		asm.DATA{asm.REG2, asm.SYMBOL{"CALL-RETURN-ADDRESS"}},
		asm.STORE{asm.REG2, asm.REG3},
		asm.DATA{asm.REG2, asm.SYMBOL{"PEN-POSITION-ADDR"}},
		asm.LOAD{asm.REG2, asm.REG2},
	)

	penPositionRegister := asm.REG2

	instructions.Add(
		asm.DATA{asm.REG0, asm.NUMBER{fontYAddr}},
		asm.DATA{asm.REG1, asm.NUMBER{0x0000}},
		asm.STORE{asm.REG0, asm.REG1},
	)

	instructions.Add(
		asm.DATA{asm.REG3, asm.SYMBOL{"KEYCODE-REGISTER"}},
		asm.LOAD{asm.REG3, asm.REG3},

		asm.DATA{asm.REG1, asm.NUMBER{0x0107}},
		asm.CMP{asm.REG3, asm.REG1},
		asm.JMPF{[]string{"E"}, asm.LABEL{labelPrefix+"-left"}},
		asm.DATA{asm.REG1, asm.NUMBER{0x0106}},
		asm.CMP{asm.REG3, asm.REG1},
		asm.JMPF{[]string{"E"}, asm.LABEL{labelPrefix+"-right"}},
		asm.DATA{asm.REG1, asm.NUMBER{0x0108}},
		asm.CMP{asm.REG3, asm.REG1},
		asm.JMPF{[]string{"E"}, asm.LABEL{labelPrefix+"-down"}},
		asm.DATA{asm.REG1, asm.NUMBER{0x0109}},
		asm.CMP{asm.REG3, asm.REG1},
		asm.JMPF{[]string{"E"}, asm.LABEL{labelPrefix+"-up"}},
		asm.JMP{asm.LABEL{labelPrefix+"-start"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-right"},
		asm.DATA{asm.REG1, asm.SYMBOL{"ONE"}},
		asm.ADD{asm.REG1, asm.REG2},
		asm.DATA{asm.REG3, asm.SYMBOL{"PEN-POSITION-ADDR"}},
		asm.STORE{asm.REG3, asm.REG2},
		asm.JMP{asm.LABEL{labelPrefix+"-start"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-down"},
		asm.DATA{asm.REG1, asm.NUMBER{0x00F0}},
		asm.ADD{asm.REG1, asm.REG2},
		asm.DATA{asm.REG3, asm.SYMBOL{"PEN-POSITION-ADDR"}},
		asm.STORE{asm.REG3, asm.REG2},
		asm.JMP{asm.LABEL{labelPrefix+"-start"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-up"},
		asm.DATA{asm.REG0, asm.SYMBOL{"ONE"}},
		asm.DATA{asm.REG1, asm.NUMBER{0x00F0}},
		asm.NOT{asm.REG1},
		asm.ADD{asm.REG0, asm.REG1},
		asm.CLF{},
		asm.ADD{asm.REG1, asm.REG2},
		asm.DATA{asm.REG3, asm.SYMBOL{"PEN-POSITION-ADDR"}},
		asm.STORE{asm.REG3, asm.REG2},
		asm.JMP{asm.LABEL{labelPrefix+"-start"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-left"},
		asm.DATA{asm.REG0, asm.SYMBOL{"ONE"}},
		asm.DATA{asm.REG1, asm.SYMBOL{"ONE"}},
		asm.NOT{asm.REG1},
		asm.ADD{asm.REG0, asm.REG1},
		asm.CLF{},
		asm.ADD{asm.REG1, asm.REG2},
		asm.DATA{asm.REG3, asm.SYMBOL{"PEN-POSITION-ADDR"}},
		asm.STORE{asm.REG3, asm.REG2},
		asm.JMP{asm.LABEL{labelPrefix+"-start"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-start"})
	instructions.AddBlocks(selectDisplayAdapter(asm.REG3))

	instructions.Add(
		asm.DEFLABEL{labelPrefix+"-STARTLOOP"},
		asm.DATA{asm.REG3, asm.NUMBER{0x0000}},
		asm.SHL{asm.REG3},
		asm.SHL{asm.REG3},
		asm.SHL{asm.REG3},
		asm.DATA{asm.REG0, asm.NUMBER{fontYAddr}},
		asm.LOAD{asm.REG0, asm.REG0},
		asm.ADD{asm.REG0, asm.REG3},

		asm.DATA{asm.REG1, asm.SYMBOL{"ONE"}},
		asm.ADD{asm.REG1, asm.REG0},
		asm.DATA{asm.REG1, asm.NUMBER{fontYAddr}},
		asm.STORE{asm.REG1, asm.REG0},

		asm.LOAD{asm.REG3, asm.REG0},

		asm.OUT{asm.DATA_MODE, penPositionRegister},
		asm.OUT{asm.DATA_MODE, asm.REG0},
		asm.DATA{asm.REG1, asm.SYMBOL{"LINE-WIDTH"}},
		asm.ADD{asm.REG1, penPositionRegister},

		asm.DATA{asm.REG0, asm.NUMBER{fontYAddr}},
		asm.LOAD{asm.REG0, asm.REG0},
		asm.DATA{asm.REG1, asm.NUMBER{0x0008}},
		asm.CMP{asm.REG0, asm.REG1},

		asm.JMPF{[]string{"E"}, asm.LABEL{labelPrefix+"-ENDLOOP"}},
		asm.JMP{asm.LABEL{labelPrefix+"-STARTLOOP"}},
	)

	instructions.Add(asm.DEFLABEL{labelPrefix+"-ENDLOOP"})
	instructions.Add(asm.DEFLABEL{labelPrefix+"-deselectIO"})
	instructions.AddBlocks(deselectIO(asm.REG3))

	instructions.Add(
		asm.CLF{},
		asm.DATA{asm.REG3, asm.SYMBOL{"CALL-RETURN-ADDRESS"}},
		asm.LOAD{asm.REG3, asm.REG3},
		asm.JR{asm.REG3},
	)

	return instructions.Get()
}
