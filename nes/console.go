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
	console.CPU = NewCPU(&console)
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
	for i := 0; i < cpuCycles; i++ {
		console.APU.Step()
	}
	return cpuCycles
}

func (console *Console) StepFrame() int {
	cpuCycles := 0
	frame := console.PPU.Frame
	for frame == console.PPU.Frame {
		cpuCycles += console.Step()
	}
	return cpuCycles
}

func (console *Console) StepSeconds(seconds float64) {
	cycles := int(CPUFrequency * seconds)
	for cycles > 0 {
		cycles -= console.Step()
	}
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
