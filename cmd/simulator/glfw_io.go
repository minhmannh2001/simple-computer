package main

import (
	"log"
	"time"

	sio "simple-computer/io"

	"github.com/go-gl/gl/v3.2-compatibility/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// GlfwIO drives the system using a GLFW window (requires libglfw3).
type GlfwIO struct {
	glfwDisplay     *glfwDisplay
	screenChannel   chan *[160][240]byte
	keyPressChannel chan *sio.KeyPress
	quitChannel     chan bool
}

func NewGlfwIO(screenChannel chan *[160][240]byte, keyPressChannel chan *sio.KeyPress, quitChannel chan bool) *GlfwIO {
	log.Println("Creating GLFW based IO Handler")
	d := &glfwDisplay{onCloseHandler: func() { close(quitChannel) }}
	return &GlfwIO{d, screenChannel, keyPressChannel, quitChannel}
}

func (i *GlfwIO) Init(title string) error {
	if err := i.glfwDisplay.init(title); err != nil {
		return err
	}

	i.glfwDisplay.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Repeat {
			i.keyPressChannel <- &down_key_presses[int(key)]
			return
		}
		if action == glfw.Press {
			i.keyPressChannel <- &down_key_presses[int(key)]
		} else {
			i.keyPressChannel <- &up_key_presses[int(key)]
		}
	})

	return nil
}

func (i *GlfwIO) Run() {
	clock := time.Tick(33 * time.Millisecond)
	for {
		<-clock
		select {
		case <-i.quitChannel:
			i.glfwDisplay.Destroy()
			return
		case frame := <-i.screenChannel:
			i.glfwDisplay.DrawFrame(frame)
		}
	}
}

type glfwDisplay struct {
	onCloseHandler func()
	window         *glfw.Window
}

func (s *glfwDisplay) init(title string) error {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	window, err := glfw.CreateWindow(240, 160, title, nil, nil)
	if err != nil {
		return err
	}

	if monitor := glfw.GetPrimaryMonitor(); monitor != nil {
		if vidMode := monitor.GetVideoMode(); vidMode != nil {
			window.SetPos(vidMode.Width/3, vidMode.Height/3)
		}
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return err
	}

	gl.ClearColor(0.255, 0.255, 0.255, 0)
	window.SetCloseCallback(func(w *glfw.Window) { s.onCloseHandler() })
	s.window = window
	return nil
}

func (s *glfwDisplay) Destroy() {
	log.Println("Destroying window")
	s.window.Destroy()
	log.Println("Destroying GLFW instance")
	glfw.Terminate()
}

func (s *glfwDisplay) DrawFrame(screenData *[160][240]byte) {
	fw, fh := s.window.GetFramebufferSize()
	gl.Viewport(0, 0, int32(fw), int32(fh))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(240), float64(160), 0, -1, 1)
	gl.ClearColor(0.255, 0.255, 0.255, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Disable(gl.DEPTH_TEST)
	gl.PointSize(2.0)
	gl.Begin(gl.POINTS)
	for y := 0; y < 160; y++ {
		for x := 0; x < 240; x++ {
			if screenData[y][x] > 0 {
				gl.Color3ub(220, 220, 220)
			} else {
				gl.Color3ub(50, 50, 50)
			}
			gl.Vertex2i(int32(x), int32(y))
		}
	}
	gl.End()
	glfw.PollEvents()
	s.window.SwapBuffers()
}

func makeKeyPresses(isDown bool) []sio.KeyPress {
	presses := make([]sio.KeyPress, 1024)
	for i := range presses {
		presses[i] = sio.KeyPress{Value: i, IsDown: isDown}
	}
	return presses
}

var up_key_presses = makeKeyPresses(false)
var down_key_presses = makeKeyPresses(true)
