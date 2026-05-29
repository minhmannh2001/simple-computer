package cpu

import (
	"simple-computer/alu"
	"simple-computer/circuit"
	"simple-computer/components"
	sio "simple-computer/io"
	"simple-computer/memory"
)

const BUS_WIDTH = 16

const (
	FLAGS_BUS_CARRY    = 0
	FLAGS_BUS_A_LARGER = 1
	FLAGS_BUS_EQUAL    = 2
	FLAGS_BUS_ZERO     = 3
)

type Enableable interface {
	Enable()
	Disable()
}

type Settable interface {
	Set()
	Unset()
}

type Updatable interface {
	Update()
}

type InstructionDecoder3x8 struct {
	decoder       components.Decoder3x8
	selectorGates [8]circuit.ANDGate
	bit0NOTGate   circuit.NOTGate
}

func NewInstructionDecoder3x8() *InstructionDecoder3x8 {
	i := new(InstructionDecoder3x8)
	i.decoder = *components.NewDecoder3x8()
	for n := range i.selectorGates {
		i.selectorGates[n] = *circuit.NewANDGate()
	}
	i.bit0NOTGate = *circuit.NewNOTGate()
	return i
}

type CPU struct {
	gpReg0 components.Register
	gpReg1 components.Register
	gpReg2 components.Register
	gpReg3 components.Register
	tmp    components.Register
	acc    components.Register
	ir     components.Register
	iar    components.Register
	flags  components.Register

	clockState bool
	memory     *memory.Memory64K
	alu        *alu.ALU
	stepper    *components.Stepper
	busOne     components.BusOne

	mainBus       *components.Bus
	tmpBus        *components.Bus
	busOneOutput  *components.Bus
	controlBus    *components.Bus
	accBus        *components.Bus
	aluToFlagsBus *components.Bus
	flagsBus      *components.Bus
	ioBus         *components.IOBus

	step4Gates     [8]circuit.ANDGate
	step4Gate3And  components.ANDGate3
	step5Gates     [6]circuit.ANDGate
	step5Gate3And  components.ANDGate3
	step6Gates     [2]components.ANDGate3
	step6Gates2And circuit.ANDGate

	instrDecoder3x8              InstructionDecoder3x8
	instructionDecoderEnables2x4 [2]components.Decoder2x4
	instructionDecoderSet2x4     components.Decoder2x4

	irInstructionANDGate components.ANDGate3
	irInstructionNOTGate circuit.NOTGate

	ioBusEnableGate       circuit.ANDGate
	registerAEnableORGate components.ORGate3
	registerBEnableORGate components.ORGate4
	registerBSetORGate    components.ORGate4
	registerAEnable       circuit.Wire
	registerBEnable       circuit.Wire
	accEnableORGate       components.ORGate4
	accEnableANDGate      circuit.ANDGate
	busOneEnableORGate    components.ORGate4
	iarEnableORGate       components.ORGate4
	iarEnableANDGate      circuit.ANDGate
	ramEnableORGate       components.ORGate5
	ramEnableANDGate      circuit.ANDGate
	gpRegEnableANDGates   [8]components.ANDGate3
	gpRegEnableORGates    [4]circuit.ORGate

	gpRegSetANDGates [4]components.ANDGate3

	ioBusSetGate    circuit.ANDGate
	irBit4NOTGate   circuit.NOTGate
	irSetANDGate    circuit.ANDGate
	marSetORGate    components.ORGate6
	marSetANDGate   circuit.ANDGate
	iarSetORGate    components.ORGate6
	iarSetANDGate   circuit.ANDGate
	accSetORGate    components.ORGate4
	accSetANDGate   circuit.ANDGate
	ramSetANDGate   circuit.ANDGate
	tmpSetANDGate   circuit.ANDGate
	flagsSetORGate  circuit.ORGate
	flagsSetANDGate circuit.ANDGate
	registerBSet    circuit.Wire

	flagStateGates  [4]circuit.ANDGate
	flagStateORGate components.ORGate4

	aluOpAndGates [3]components.ANDGate3

	carryTemp    components.Bit
	carryANDGate circuit.ANDGate

	peripherals []sio.Peripheral
}

func NewCPU(mainBus *components.Bus, mem *memory.Memory64K) *CPU {
	c := new(CPU)

	c.clockState = false
	c.stepper = components.NewStepper()
	c.memory = mem

	c.controlBus = components.NewBus(BUS_WIDTH)
	c.mainBus = mainBus
	c.gpReg0 = *components.NewRegister("R0", c.mainBus, c.mainBus)
	c.gpReg1 = *components.NewRegister("R1", c.mainBus, c.mainBus)
	c.gpReg2 = *components.NewRegister("R2", c.mainBus, c.mainBus)
	c.gpReg3 = *components.NewRegister("R3", c.mainBus, c.mainBus)
	c.ir = *components.NewRegister("IR", c.mainBus, c.controlBus)
	c.ir.Disable()
	c.iar = *components.NewRegister("IAR", c.mainBus, c.mainBus)

	c.instructionDecoderEnables2x4[0] = *components.NewDecoder2x4()
	c.instructionDecoderEnables2x4[1] = *components.NewDecoder2x4()
	c.instructionDecoderSet2x4 = *components.NewDecoder2x4()

	c.instrDecoder3x8 = *NewInstructionDecoder3x8()

	c.aluToFlagsBus = components.NewBus(BUS_WIDTH)
	c.flagsBus = components.NewBus(BUS_WIDTH)
	c.flags = *components.NewRegister("FLAGS", c.aluToFlagsBus, c.flagsBus)
	updateEnableStatus(&c.flags, true)
	updateSetStatus(&c.flags, true)
	runUpdateOn(&c.flags)
	updateSetStatus(&c.flags, false)

	c.tmpBus = components.NewBus(BUS_WIDTH)
	c.tmp = *components.NewRegister("TMP", c.mainBus, c.tmpBus)
	updateEnableStatus(&c.tmp, true)
	updateSetStatus(&c.tmp, true)
	runUpdateOn(&c.tmp)
	updateSetStatus(&c.tmp, false)

	c.busOneOutput = components.NewBus(BUS_WIDTH)
	c.busOne = *components.NewBusOne(c.tmpBus, c.busOneOutput)

	c.accBus = components.NewBus(BUS_WIDTH)
	c.acc = *components.NewRegister("ACC", c.accBus, c.mainBus)

	c.alu = alu.NewALU(c.mainBus, c.busOneOutput, c.accBus, c.aluToFlagsBus)

	c.irInstructionANDGate = *components.NewANDGate3()
	c.irInstructionNOTGate = *circuit.NewNOTGate()
	c.irBit4NOTGate = *circuit.NewNOTGate()

	c.registerAEnableORGate = *components.NewORGate3()
	c.registerBEnableORGate = *components.NewORGate4()
	c.registerBSetORGate = *components.NewORGate4()
	c.accEnableORGate = *components.NewORGate4()
	c.accEnableANDGate = *circuit.NewANDGate()
	c.busOneEnableORGate = *components.NewORGate4()
	c.iarEnableORGate = *components.NewORGate4()
	c.iarEnableANDGate = *circuit.NewANDGate()
	c.ramEnableORGate = *components.NewORGate5()
	c.ramEnableANDGate = *circuit.NewANDGate()

	c.irSetANDGate = *circuit.NewANDGate()
	c.marSetORGate = *components.NewORGate6()
	c.marSetANDGate = *circuit.NewANDGate()
	c.iarSetORGate = *components.NewORGate6()
	c.iarSetANDGate = *circuit.NewANDGate()
	c.accSetORGate = *components.NewORGate4()
	c.accSetANDGate = *circuit.NewANDGate()
	c.ramSetANDGate = *circuit.NewANDGate()
	c.tmpSetANDGate = *circuit.NewANDGate()
	c.flagsSetORGate = *circuit.NewORGate()
	c.flagsSetANDGate = *circuit.NewANDGate()

	c.carryTemp = *components.NewBit()
	c.carryANDGate = *circuit.NewANDGate()

	for i := range c.step4Gates {
		c.step4Gates[i] = *circuit.NewANDGate()
	}
	c.step4Gate3And = *components.NewANDGate3()

	for i := range c.step5Gates {
		c.step5Gates[i] = *circuit.NewANDGate()
	}
	c.step5Gate3And = *components.NewANDGate3()

	for i := range c.step6Gates {
		c.step6Gates[i] = *components.NewANDGate3()
	}
	c.step6Gates2And = *circuit.NewANDGate()

	for i := range c.flagStateGates {
		c.flagStateGates[i] = *circuit.NewANDGate()
	}
	c.flagStateORGate = *components.NewORGate4()

	for i := range c.gpRegEnableORGates {
		c.gpRegEnableORGates[i] = *circuit.NewORGate()
	}

	for i := range c.gpRegEnableANDGates {
		c.gpRegEnableANDGates[i] = *components.NewANDGate3()
	}

	for i := range c.gpRegSetANDGates {
		c.gpRegSetANDGates[i] = *components.NewANDGate3()
	}

	for i := range c.aluOpAndGates {
		c.aluOpAndGates[i] = *components.NewANDGate3()
	}

	c.ioBus = components.NewIOBus()
	c.ioBusEnableGate = *circuit.NewANDGate()
	c.ioBusSetGate = *circuit.NewANDGate()

	c.peripherals = make([]sio.Peripheral, 0)

	return c
}

func (c *CPU) ConnectPeripheral(p sio.Peripheral) {
	p.Connect(c.ioBus, c.mainBus)
	c.peripherals = append(c.peripherals, p)
}

func (c *CPU) SetIAR(address uint16) {
	c.mainBus.SetValue(address)
	updateSetStatus(&c.iar, true)
	runUpdateOn(&c.iar)
	updateSetStatus(&c.iar, false)
	runUpdateOn(&c.iar)
	c.clearMainBus()
}

func (c *CPU) Step() {
	for i := 0; i < 2; i++ {
		c.clockState = !c.clockState
		c.step(c.clockState)
	}
}

func (c *CPU) step(clockState bool) {
	c.stepper.Update(clockState)
	c.runStep4Gates()
	c.runStep5Gates()
	c.runStep6Gates()

	c.runEnable(clockState)
	c.updateStates()
	if clockState {
		c.runEnable(false)
		c.updateStates()
	}

	c.runSet(clockState)
	c.updateStates()
	if clockState {
		c.runSet(false)
		c.updateStates()
	}

	c.clearMainBus()
}

func (c *CPU) clearMainBus() {
	for i := 0; i < BUS_WIDTH; i++ {
		c.mainBus.SetInputWire(i, false)
	}
}

func (c *CPU) updateStates() {
	runUpdateOn(&c.iar)
	runUpdateOn(&c.memory.AddressRegister)
	runUpdateOn(&c.ir)
	runUpdateOn(c.memory)
	runUpdateOn(&c.tmp)
	runUpdateOn(&c.flags)
	runUpdateOn(&c.busOne)
	c.updateALU()
	runUpdateOn(&c.acc)
	runUpdateOn(&c.gpReg0)
	runUpdateOn(&c.gpReg1)
	runUpdateOn(&c.gpReg2)
	runUpdateOn(&c.gpReg3)
	c.updateInstructionDecoder3x8()
	c.updateIOBus()
	c.updatePeripherals()
}

func (c *CPU) updatePeripherals() {
	for _, p := range c.peripherals {
		p.Update()
	}
}

func (c *CPU) updateIOBus() {
	c.ioBus.Update(c.ir.Bit(12), c.ir.Bit(13))
}

func (c *CPU) updateALU() {
	c.aluOpAndGates[2].Update(c.ir.Bit(9), c.ir.Bit(8), c.stepper.GetOutputWire(5))
	c.aluOpAndGates[1].Update(c.ir.Bit(10), c.ir.Bit(8), c.stepper.GetOutputWire(5))
	c.aluOpAndGates[0].Update(c.ir.Bit(11), c.ir.Bit(8), c.stepper.GetOutputWire(5))

	c.alu.Op[2].Update(c.aluOpAndGates[2].Output())
	c.alu.Op[1].Update(c.aluOpAndGates[1].Output())
	c.alu.Op[0].Update(c.aluOpAndGates[0].Output())

	c.alu.CarryIn.Update(c.carryANDGate.Output())
	c.alu.Update()
}

func (c *CPU) updateInstructionDecoder3x8() {
	c.instrDecoder3x8.bit0NOTGate.Update(c.ir.Bit(8))
	c.instrDecoder3x8.decoder.Update(c.ir.Bit(9), c.ir.Bit(10), c.ir.Bit(11))
	for i := 0; i < 8; i++ {
		c.instrDecoder3x8.selectorGates[i].Update(c.instrDecoder3x8.decoder.GetOutputWire(i), c.instrDecoder3x8.bit0NOTGate.Output())
	}
}

func (c *CPU) runStep4Gates() {
	c.step4Gates[0].Update(c.stepper.GetOutputWire(4), c.ir.Bit(8))

	gate := 1
	for selector := 0; selector < 7; selector++ {
		c.step4Gates[gate].Update(c.stepper.GetOutputWire(4), c.instrDecoder3x8.selectorGates[selector].Output())
		gate++
	}

	c.step4Gate3And.Update(c.stepper.GetOutputWire(4), c.instrDecoder3x8.selectorGates[7].Output(), c.ir.Bit(12))
	c.irBit4NOTGate.Update(c.ir.Bit(12))
}

func (c *CPU) runStep5Gates() {
	c.step5Gates[0].Update(c.stepper.GetOutputWire(5), c.ir.Bit(8))
	c.step5Gates[1].Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[0].Output())
	c.step5Gates[2].Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[1].Output())
	c.step5Gates[3].Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[2].Output())
	c.step5Gates[4].Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[4].Output())
	c.step5Gates[5].Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[5].Output())
	c.step5Gate3And.Update(c.stepper.GetOutputWire(5), c.instrDecoder3x8.selectorGates[7].Output(), c.irBit4NOTGate.Output())
}

func (c *CPU) runStep6Gates() {
	c.step6Gates[0].Update(c.stepper.GetOutputWire(0), c.ir.Bit(8), c.irInstructionNOTGate.Output())
	c.step6Gates2And.Update(c.stepper.GetOutputWire(0), c.instrDecoder3x8.selectorGates[2].Output())
	c.step6Gates[1].Update(c.stepper.GetOutputWire(0), c.instrDecoder3x8.selectorGates[5].Output(), c.flagStateORGate.Output())
}

func (c *CPU) runEnable(state bool) {
	c.runEnableOnIO(state)
	c.runEnableOnIAR(state)
	c.runEnableOnBusOne(state)
	c.runEnableOnACC(state)
	c.runEnableOnRAM(state)
	c.runEnableOnRegisterB()
	c.runEnableOnRegisterA()
	c.runEnableGeneralPurposeRegisters(state)
}

func (c *CPU) runEnableOnIO(state bool) {
	c.ioBusEnableGate.Update(state, c.step5Gate3And.Output())
	updateEnableStatus(c.ioBus, c.ioBusEnableGate.Output())
}

func (c *CPU) runEnableOnRegisterB() {
	c.registerBEnableORGate.Update(c.step4Gates[0].Output(), c.step5Gates[2].Output(), c.step4Gates[4].Output(), c.step4Gate3And.Output())
	c.registerBEnable.Update(c.registerBEnableORGate.Output())
}

func (c *CPU) runEnableOnRegisterA() {
	c.registerAEnableORGate.Update(c.step4Gates[1].Output(), c.step4Gates[2].Output(), c.step5Gates[0].Output())
	c.registerAEnable.Update(c.registerAEnableORGate.Output())
}

func (c *CPU) runEnableOnBusOne(state bool) {
	c.busOneEnableORGate.Update(c.stepper.GetOutputWire(1), c.step4Gates[7].Output(), c.step4Gates[6].Output(), c.step4Gates[3].Output())
	updateEnableStatus(&c.busOne, c.busOneEnableORGate.Output())
}

func (c *CPU) runEnableOnACC(state bool) {
	c.accEnableORGate.Update(c.stepper.GetOutputWire(3), c.step5Gates[5].Output(), c.step6Gates2And.Output(), c.step6Gates[0].Output())
	c.accEnableANDGate.Update(state, c.accEnableORGate.Output())
	updateEnableStatus(&c.acc, c.accEnableANDGate.Output())
}

func (c *CPU) runEnableOnIAR(state bool) {
	c.iarEnableORGate.Update(c.stepper.GetOutputWire(1), c.step4Gates[3].Output(), c.step4Gates[5].Output(), c.step4Gates[6].Output())
	c.iarEnableANDGate.Update(state, c.iarEnableORGate.Output())
	updateEnableStatus(&c.iar, c.iarEnableANDGate.Output())
}

func (c *CPU) runEnableOnRAM(state bool) {
	c.ramEnableORGate.Update(
		c.stepper.GetOutputWire(2),
		c.step6Gates[1].Output(),
		c.step5Gates[4].Output(),
		c.step5Gates[3].Output(),
		c.step5Gates[1].Output(),
	)
	c.ramEnableANDGate.Update(state, c.ramEnableORGate.Output())
	updateEnableStatus(c.memory, c.ramEnableANDGate.Output())
}

func (c *CPU) runEnableGeneralPurposeRegisters(state bool) {
	c.instructionDecoderEnables2x4[0].Update(c.ir.Bit(14), c.ir.Bit(15))
	c.instructionDecoderEnables2x4[1].Update(c.ir.Bit(12), c.ir.Bit(13))

	c.gpRegEnableANDGates[0].Update(state, c.registerBEnable.Get(), c.instructionDecoderEnables2x4[0].GetOutputWire(0))
	c.gpRegEnableANDGates[4].Update(state, c.registerAEnable.Get(), c.instructionDecoderEnables2x4[1].GetOutputWire(0))
	c.gpRegEnableORGates[0].Update(c.gpRegEnableANDGates[4].Output(), c.gpRegEnableANDGates[0].Output())
	updateEnableStatus(&c.gpReg0, c.gpRegEnableORGates[0].Output())

	c.gpRegEnableANDGates[1].Update(state, c.registerBEnable.Get(), c.instructionDecoderEnables2x4[0].GetOutputWire(1))
	c.gpRegEnableANDGates[5].Update(state, c.registerAEnable.Get(), c.instructionDecoderEnables2x4[1].GetOutputWire(1))
	c.gpRegEnableORGates[1].Update(c.gpRegEnableANDGates[5].Output(), c.gpRegEnableANDGates[1].Output())
	updateEnableStatus(&c.gpReg1, c.gpRegEnableORGates[1].Output())

	c.gpRegEnableANDGates[2].Update(state, c.registerBEnable.Get(), c.instructionDecoderEnables2x4[0].GetOutputWire(2))
	c.gpRegEnableANDGates[6].Update(state, c.registerAEnable.Get(), c.instructionDecoderEnables2x4[1].GetOutputWire(2))
	c.gpRegEnableORGates[2].Update(c.gpRegEnableANDGates[6].Output(), c.gpRegEnableANDGates[2].Output())
	updateEnableStatus(&c.gpReg2, c.gpRegEnableORGates[2].Output())

	c.gpRegEnableANDGates[3].Update(state, c.registerBEnable.Get(), c.instructionDecoderEnables2x4[0].GetOutputWire(3))
	c.gpRegEnableANDGates[7].Update(state, c.registerAEnable.Get(), c.instructionDecoderEnables2x4[1].GetOutputWire(3))
	c.gpRegEnableORGates[3].Update(c.gpRegEnableANDGates[7].Output(), c.gpRegEnableANDGates[3].Output())
	updateEnableStatus(&c.gpReg3, c.gpRegEnableORGates[3].Output())
}

func (c *CPU) runSet(state bool) {
	c.irInstructionANDGate.Update(c.ir.Bit(11), c.ir.Bit(10), c.ir.Bit(9))
	c.irInstructionNOTGate.Update(c.irInstructionANDGate.Output())

	c.refreshFlagStateGates()

	c.runSetOnIO(state)
	c.runSetOnMAR(state)
	c.runSetOnIAR(state)
	c.runSetOnIR(state)
	c.runSetOnACC(state)
	c.runSetOnRAM(state)
	c.runSetOnTMP(state)
	c.runSetOnFLAGS(state)
	c.runSetOnRegisterB()
	c.runSetGeneralPurposeRegisters(state)
}

func (c *CPU) refreshFlagStateGates() {
	c.flagStateGates[0].Update(c.ir.Bit(12), c.flagsBus.GetOutputWire(FLAGS_BUS_CARRY))
	c.flagStateGates[1].Update(c.ir.Bit(13), c.flagsBus.GetOutputWire(FLAGS_BUS_A_LARGER))
	c.flagStateGates[2].Update(c.ir.Bit(14), c.flagsBus.GetOutputWire(FLAGS_BUS_EQUAL))
	c.flagStateGates[3].Update(c.ir.Bit(15), c.flagsBus.GetOutputWire(FLAGS_BUS_ZERO))
	c.flagStateORGate.Update(
		c.flagStateGates[0].Output(),
		c.flagStateGates[1].Output(),
		c.flagStateGates[2].Output(),
		c.flagStateGates[3].Output(),
	)
}

func (c *CPU) runSetOnIO(state bool) {
	c.ioBusSetGate.Update(state, c.step4Gate3And.Output())
	updateSetStatus(c.ioBus, c.ioBusSetGate.Output())
}

func (c *CPU) runSetOnMAR(state bool) {
	c.marSetORGate.Update(
		c.stepper.GetOutputWire(1),
		c.step4Gates[3].Output(),
		c.step4Gates[6].Output(),
		c.step4Gates[1].Output(),
		c.step4Gates[2].Output(),
		c.step4Gates[5].Output(),
	)
	c.marSetANDGate.Update(state, c.marSetORGate.Output())
	updateSetStatus(&c.memory.AddressRegister, c.marSetANDGate.Output())
}

func (c *CPU) runSetOnIAR(state bool) {
	c.iarSetORGate.Update(
		c.stepper.GetOutputWire(3),
		c.step4Gates[4].Output(),
		c.step5Gates[4].Output(),
		c.step5Gates[5].Output(),
		c.step6Gates2And.Output(),
		c.step6Gates[1].Output(),
	)
	c.iarSetANDGate.Update(state, c.iarSetORGate.Output())
	updateSetStatus(&c.iar, c.iarSetANDGate.Output())
}

func (c *CPU) runSetOnIR(state bool) {
	c.irSetANDGate.Update(state, c.stepper.GetOutputWire(2))
	updateSetStatus(&c.ir, c.irSetANDGate.Output())
}

func (c *CPU) runSetOnACC(state bool) {
	c.accSetORGate.Update(
		c.stepper.GetOutputWire(1),
		c.step4Gates[3].Output(),
		c.step4Gates[6].Output(),
		c.step5Gates[0].Output(),
	)
	c.accSetANDGate.Update(state, c.accSetORGate.Output())
	updateSetStatus(&c.acc, c.accSetANDGate.Output())
}

func (c *CPU) runSetOnFLAGS(state bool) {
	c.flagsSetORGate.Update(
		c.step5Gates[0].Output(),
		c.step4Gates[7].Output(),
	)
	c.flagsSetANDGate.Update(state, c.flagsSetORGate.Output())
	updateSetStatus(&c.flags, c.flagsSetANDGate.Output())
}

func (c *CPU) runSetOnRAM(state bool) {
	c.ramSetANDGate.Update(state, c.step5Gates[2].Output())
	updateSetStatus(c.memory, c.ramSetANDGate.Output())
}

func (c *CPU) runSetOnTMP(state bool) {
	c.tmpSetANDGate.Update(state, c.step4Gates[0].Output())
	updateSetStatus(&c.tmp, c.tmpSetANDGate.Output())

	c.carryTemp.Update(c.flagsBus.GetOutputWire(FLAGS_BUS_CARRY), c.tmpSetANDGate.Output())
	c.carryANDGate.Update(c.carryTemp.Get(), c.step5Gates[0].Output())
}

func (c *CPU) runSetOnRegisterB() {
	c.registerBSetORGate.Update(
		c.step5Gates[1].Output(),
		c.step6Gates[0].Output(),
		c.step5Gates[3].Output(),
		c.step5Gate3And.Output(),
	)
	c.registerBSet.Update(c.registerBSetORGate.Output())
}

func (c *CPU) runSetGeneralPurposeRegisters(state bool) {
	c.instructionDecoderSet2x4.Update(c.ir.Bit(14), c.ir.Bit(15))

	c.gpRegSetANDGates[0].Update(state, c.registerBSet.Get(), c.instructionDecoderSet2x4.GetOutputWire(0))
	updateSetStatus(&c.gpReg0, c.gpRegSetANDGates[0].Output())

	c.gpRegSetANDGates[1].Update(state, c.registerBSet.Get(), c.instructionDecoderSet2x4.GetOutputWire(1))
	updateSetStatus(&c.gpReg1, c.gpRegSetANDGates[1].Output())

	c.gpRegSetANDGates[2].Update(state, c.registerBSet.Get(), c.instructionDecoderSet2x4.GetOutputWire(2))
	updateSetStatus(&c.gpReg2, c.gpRegSetANDGates[2].Output())

	c.gpRegSetANDGates[3].Update(state, c.registerBSet.Get(), c.instructionDecoderSet2x4.GetOutputWire(3))
	updateSetStatus(&c.gpReg3, c.gpRegSetANDGates[3].Output())
}

func (c *CPU) GPReg(n int) uint16 {
	switch n {
	case 0:
		return c.gpReg0.Value()
	case 1:
		return c.gpReg1.Value()
	case 2:
		return c.gpReg2.Value()
	case 3:
		return c.gpReg3.Value()
	}
	return 0
}

func (c *CPU) EqualFlag() bool {
	return c.flagsBus.GetOutputWire(FLAGS_BUS_EQUAL)
}

func runUpdateOn(component Updatable) {
	component.Update()
}

func updateEnableStatus(component Enableable, state bool) {
	if state {
		component.Enable()
	} else {
		component.Disable()
	}
}

func updateSetStatus(component Settable, state bool) {
	if state {
		component.Set()
	} else {
		component.Unset()
	}
}
