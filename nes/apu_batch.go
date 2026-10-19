package nes

// StepCycles advances the APU after one CPU instruction. Split at frame-counter
// ticks so envelope, length and sweep changes affect precisely the same timers
// as repeated Step calls. The CPU must not execute or write registers mid-batch.
func (apu *APU) StepCycles(cycles int) {
	for cycles > 0 {
		n := cycles
		if apu.frameCounter != 0 && n > int(apu.frameCounter) {
			n = int(apu.frameCounter)
		}
		halfTicks := (n + int(apu.cycle&1)) / 2
		apu.pulse1.stepTimerCycles(halfTicks)
		apu.pulse2.stepTimerCycles(halfTicks)
		apu.noise.stepTimerCycles(halfTicks)
		apu.dmc.stepTimerCycles(halfTicks)
		apu.triangle.stepTimerCycles(n)
		apu.cycle += uint32(n)
		apu.frameCounter -= uint32(n)
		if apu.frameCounter == 0 {
			apu.frameCounter = frameCounterPeriod
			apu.stepFrameCounter()
		}
		cycles -= n
	}
}

// advanceTimer returns the countdown and number of expirations after ticks.
// A countdown of zero fires on the next tick, then reloads to period.
func advanceTimer(value, period uint16, ticks int) (uint16, int) {
	if ticks <= int(value) {
		return value - uint16(ticks), 0
	}
	remaining := ticks - int(value) - 1
	interval := int(period) + 1 // 65536 must not wrap when period is 65535.
	reloads := remaining / interval
	return period - uint16(remaining-reloads*interval), reloads + 1
}

func (p *Pulse) stepTimerCycles(ticks int) {
	value, expirations := advanceTimer(p.timerValue, p.timerPeriod, ticks)
	p.timerValue = value
	p.dutyValue = (p.dutyValue + byte(expirations)) & 7
}

func (t *Triangle) stepTimerCycles(ticks int) {
	value, expirations := advanceTimer(t.timerValue, t.timerPeriod, ticks)
	t.timerValue = value
	if t.lengthValue > 0 && t.counterValue > 0 {
		t.dutyValue = (t.dutyValue + byte(expirations)) & 31
	}
}

func (n *Noise) stepTimerCycles(ticks int) {
	value, expirations := advanceTimer(n.timerValue, n.timerPeriod, ticks)
	n.timerValue = value
	shift := uint(1)
	if n.mode {
		shift = 6
	}
	for i := 0; i < expirations; i++ {
		feedback := (n.shiftRegister ^ (n.shiftRegister >> shift)) & 1
		n.shiftRegister = (n.shiftRegister >> 1) | (feedback << 14)
	}
}

func (d *DMC) stepTimerCycles(ticks int) {
	if !d.enabled {
		return
	}
	for ticks > 0 {
		// Preserve read-before-shift ordering, including the refill on the
		// first tick after the shifter consumes the final buffered bit.
		d.stepReader()
		if d.tickValue != 0 {
			n := ticks
			if n > int(d.tickValue) {
				n = int(d.tickValue)
			}
			d.tickValue -= byte(n)
			ticks -= n
			continue
		}
		d.tickValue = d.tickPeriod
		d.stepShifter()
		ticks--
	}
}
