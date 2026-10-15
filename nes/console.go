package nes

import (
	"image/color"
)

type Console struct {
	CPU           *CPU
	APU           *APU
	PPU           *PPU
	Cartridge     *Cartridge
	Controller1   *Controller
	Controller2   *Controller
	Mapper        Mapper
	RAM           []byte
	mapperHasStep bool
}

func NewConsole(cartridge *Cartridge) (*Console, error) {
	ram := make([]byte, 2048)
	controller1 := NewController()
	controller2 := NewController()
	console := Console{
		nil, nil, nil, cartridge, controller1, controller2, nil, ram, false}
	mapper, err := NewMapper(&console)
	if err != nil {
		return nil, err
	}
	console.Mapper = mapper
	// Only mappers 4 and 40 have non-empty Step() methods
	switch cartridge.Mapper {
	case 4, 40:
		console.mapperHasStep = true
	}
	// Keep startup independent of interface dispatch for the TinyGo target.
	// Most mappers initially map the last PRG bank at $FFFC; AxROM maps
	// the first 32KB bank, and mapper 40 maps its fixed eighth 8KB bank.
	vectorEnd := len(cartridge.PRG)
	switch cartridge.Mapper {
	case 7:
		vectorEnd = 0x8000
	case 40:
		vectorEnd = 0x10000
	}
	console.CPU = NewCPUNoReset(&console)
	console.CPU.PC = uint16(cartridge.PRG[vectorEnd-4]) | uint16(cartridge.PRG[vectorEnd-3])<<8
	console.CPU.SP = 0xFD
	console.CPU.SetFlags(0x24)
	console.APU = NewAPU(&console)
	console.PPU = NewPPU(&console)
	return &console, nil
}

func (console *Console) Reset() {
	console.CPU.Reset()
}

func (console *Console) Step() int {
	cpuCycles := console.CPU.Step()
	ppuCycles := cpuCycles * 3
	for i := 0; i < ppuCycles; i++ {
		console.PPU.Step()
		if console.mapperHasStep {
			console.Mapper.Step()
		}
	}
	console.APU.StepCycles(cpuCycles)
	return cpuCycles
}

func (console *Console) StepFrame() int {
	cpuCycles := 0
	startFrame := console.PPU.Frame
	// Mapper 40 counts PPU cycles, so scanline-only mapper calls are insufficient.
	if console.Cartridge.Mapper == 40 {
		for startFrame == console.PPU.Frame {
			cpuCycles += console.Step()
		}
		return cpuCycles
	}
	for startFrame == console.PPU.Frame {
		// Run CPU for approximately one scanline (~114 CPU cycles = ~341 PPU cycles)
		scanlineCycles := 0
		for scanlineCycles < 114 {
			c := console.CPU.Step()
			scanlineCycles += c
			cpuCycles += c
			console.APU.StepCycles(c)
		}
		// Process one scanline
		console.PPU.RenderScanline()
	}
	return cpuCycles
}

func (console *Console) StepSeconds(seconds float64) {
	cycles := int(CPUFrequency * seconds)
	for cycles > 0 {
		cycles -= console.Step()
	}
}

// SetFrameOutput controls pixel generation only; CPU/APU, PPU timing and
// status flags continue to advance. Re-enable before a frame that is displayed.
func (console *Console) SetFrameOutput(enabled bool) {
	console.PPU.skipPixels = !enabled
}

func (console *Console) Buffer() *[256 * 240]byte {
	return console.PPU.front
}

func (console *Console) BackgroundColor() color.RGBA {
	return Palette[console.PPU.readPalette(0)%64]
}

func (console *Console) SetButtons1(buttons [8]bool) {
	console.Controller1.SetButtons(buttons)
}

func (console *Console) SetButtons2(buttons [8]bool) {
	console.Controller2.SetButtons(buttons)
}

func (console *Console) SetAudioChannel(channel chan float32) {
	console.APU.channel = channel
}

func (console *Console) SetAudioSampleRate(sampleRate float64) {
	if sampleRate != 0 {
		// Convert samples per second to cpu steps per sample
		console.APU.sampleRate = CPUFrequency / sampleRate
		// Initialize filters
		console.APU.filterChain = FilterChain{
			HighPassFilter(float32(sampleRate), 90),
			HighPassFilter(float32(sampleRate), 440),
			LowPassFilter(float32(sampleRate), 14000),
		}
	} else {
		console.APU.filterChain = nil
	}
}
