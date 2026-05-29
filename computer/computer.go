package computer

import (
	"time"

	"simple-computer/components"
	"simple-computer/cpu"
	sio "simple-computer/io"
	"simple-computer/memory"
)

const CODE_REGION_START = uint16(0x0500)

type PrintStateConfig struct {
	PrintState      bool
	PrintStateEvery int
}

type SimpleComputer struct {
	memory  *memory.Memory64K
	cpu     *cpu.CPU
	mainBus *components.Bus

	displayAdapter  *sio.DisplayAdapter
	screenControl   *sio.ScreenControl
	keyboardAdapter *sio.KeyboardAdapter

	screenChannel chan *[160][240]byte
	quitChannel   chan bool
}

func NewComputer(screenChannel chan *[160][240]byte, quitChannel chan bool) *SimpleComputer {
	c := new(SimpleComputer)
	c.screenChannel = screenChannel
	c.quitChannel = quitChannel

	c.mainBus = components.NewBus(16)
	c.memory = memory.NewMemory64K(c.mainBus)
	c.cpu = cpu.NewCPU(c.mainBus, c.memory)

	c.keyboardAdapter = sio.NewKeyboardAdapter()
	c.cpu.ConnectPeripheral(c.keyboardAdapter)

	c.displayAdapter = sio.NewDisplayAdapter()
	c.screenControl = sio.NewScreenControl(c.displayAdapter, c.screenChannel, c.quitChannel)
	c.cpu.ConnectPeripheral(c.displayAdapter)

	return c
}

func (c *SimpleComputer) ConnectKeyboard(keyboard *sio.Keyboard) {
	keyboard.ConnectTo(c.keyboardAdapter.KeyboardInBus)
}

func (c *SimpleComputer) LoadToRAM(offset uint16, values []uint16) {
	if offset < 0x0500 {
		panic("0x0000 - 0x04FF is a reserved memory area")
	}
	if offset > 0xFEFF {
		panic("0xFEFF - 0xFFFF is a reserved memory area")
	}
	for i, v := range values {
		c.putValueInRAM(offset+uint16(i), v)
	}
}

func (c *SimpleComputer) putValueInRAM(address, value uint16) {
	c.memory.AddressRegister.Set()
	c.mainBus.SetValue(address)
	c.memory.Update()

	c.memory.AddressRegister.Unset()
	c.memory.Update()

	c.mainBus.SetValue(value)
	c.memory.Set()
	c.memory.Update()

	c.memory.Unset()
	c.memory.Update()
}

func (c *SimpleComputer) Run(tickInterval <-chan time.Time, printStateConfig PrintStateConfig) {
	c.putValueInRAM(0xFEFE, 0x0040) // JMP back to code region start
	c.putValueInRAM(0xFEFF, CODE_REGION_START)
	c.cpu.SetIAR(CODE_REGION_START)
	go c.screenControl.Run()

	for {
		select {
		case <-c.quitChannel:
			return
		case <-tickInterval:
			c.cpu.Step()
		}
	}
}
