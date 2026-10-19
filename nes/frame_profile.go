//go:build perf

package nes

import "time"

// FrameProfile measures one frame. CPUAPU and PPU are available only for the
// scanline path. Total includes clock-reading and loop bookkeeping overhead.
type FrameProfile struct {
	CPUAPU, PPU, Total time.Duration
	CPUCycles          int
	Scanline           bool
}

// ProfileFrame follows StepFrame's execution order without changing the normal
// hot path. Clock reads stay outside the CPU instruction / APU cycle loops.
func (console *Console) ProfileFrame() FrameProfile {
	var p FrameProfile
	started := time.Now()
	if console.Cartridge.Mapper == 40 {
		// CPU/APU and PPU interleave per instruction on this path. Report only
		// total time rather than adding expensive instruction-level timers.
		p.CPUCycles = console.StepFrame()
		p.Total = time.Since(started)
		return p
	}

	p.Scanline = true
	startFrame := console.PPU.Frame
	for startFrame == console.PPU.Frame {
		// Match StepFrame: interleave CPU and APU, then render a scanline.
		t := time.Now()
		scanlineCycles := 0
		for scanlineCycles < 114 {
			c := console.CPU.Step()
			scanlineCycles += c
			p.CPUCycles += c
			console.APU.stepForConsole(c)
		}
		p.CPUAPU += time.Since(t)

		t = time.Now()
		console.PPU.RenderScanline()
		p.PPU += time.Since(t)
	}
	p.Total = time.Since(started)
	return p
}
