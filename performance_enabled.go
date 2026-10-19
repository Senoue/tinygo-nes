//go:build perf

package main

import (
	"time"

	"github.com/Senoue/tinygo-nes/nes"
)

const performanceWindowFrames = 120

type performanceStats struct {
	core                            [2]nes.FrameProfile
	samples                         [2]int64
	profiled                        bool
	sampledEmulation                time.Duration
	started                         time.Time
	frames, displayed               int64
	emulation, conversion, transfer time.Duration
}

type performanceMetrics struct {
	emulationFPS10, displayFPS10          int64
	emulationUS, conversionUS, transferUS int64
}

func performanceNow() time.Time { return time.Now() }

func (p *performanceStats) startWindow() {
	*p = performanceStats{started: time.Now()}
}

func (p *performanceStats) recordEmulation(started time.Time) {
	elapsed := time.Since(started)
	p.emulation += elapsed
	if p.profiled {
		p.sampledEmulation += elapsed
	}
}

// Sample one skipped and one drawn frame at the end of each window.
func (p *performanceStats) stepFrame(console *nes.Console, renderFrame bool) {
	p.profiled = p.frames >= performanceWindowFrames-2
	if p.profiled {
		i := 0
		if renderFrame {
			i = 1
		}
		p.core[i] = console.ProfileFrame()
		p.samples[i]++
	} else {
		console.StepFrame()
	}
}

func (p *performanceStats) regularEmulationUS() int64 {
	regularFrames := p.frames - p.samples[0] - p.samples[1]
	if regularFrames <= 0 {
		return 0
	}
	return (p.emulation - p.sampledEmulation).Microseconds() / regularFrames
}

func (p *performanceStats) recordConversion(started time.Time) {
	p.conversion += time.Since(started)
}

func (p *performanceStats) recordTransfer(started time.Time) {
	p.transfer += time.Since(started)
}

func (p *performanceStats) metrics(elapsed time.Duration) performanceMetrics {
	var m performanceMetrics
	if elapsed > 0 {
		m.emulationFPS10 = p.frames * 10 * int64(time.Second) / int64(elapsed)
		m.displayFPS10 = p.displayed * 10 * int64(time.Second) / int64(elapsed)
	}
	if p.frames > 0 {
		m.emulationUS = p.emulation.Microseconds() / p.frames
	}
	if p.displayed > 0 {
		m.conversionUS = p.conversion.Microseconds() / p.displayed
		m.transferUS = p.transfer.Microseconds() / p.displayed
	}
	return m
}

func (p *performanceStats) finishFrame(displayed bool) {
	p.frames++
	if displayed {
		p.displayed++
	}
	if p.frames < performanceWindowFrames {
		return
	}
	elapsed := time.Since(p.started)
	m := p.metrics(elapsed)
	// Integer output avoids fmt and floating-point formatting on the MCU.
	println("perf frames=", p.frames, "displayed=", p.displayed,
		"window_ms=", elapsed.Milliseconds(),
		"emu_fps_x10=", m.emulationFPS10, "display_fps_x10=", m.displayFPS10,
		"emu_us/frame=", m.emulationUS,
		"convert_us/display=", m.conversionUS, "transfer_us/display=", m.transferUS)
	for output, core := range p.core {
		if p.samples[output] == 0 {
			continue
		}
		if core.Scanline {
			println("perf_core samples=", p.samples[output], "frame_output=", output,
				"mode=scanline cpu_apu_us=", core.CPUAPU.Microseconds(),
				"ppu_us=", core.PPU.Microseconds(), "total_us=", core.Total.Microseconds(),
				"regular_emu_us/frame=", p.regularEmulationUS())
		} else {
			println("perf_core samples=", p.samples[output], "frame_output=", output,
				"mode=cycle_total total_us=", core.Total.Microseconds(),
				"regular_emu_us/frame=", p.regularEmulationUS())
		}
	}
	// Serial logging is excluded from both adjacent measurement windows.
	p.startWindow()
}
