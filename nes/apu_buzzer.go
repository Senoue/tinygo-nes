package nes

// StepBuzzerCycles advances the state used by BuzzerTone and by the CPU.
// Pulse/triangle waveform phases and the noise shift register are deliberately
// not maintained, so this path cannot supply PCM samples via Output().
// Register writes, envelopes, sweeps, length/linear counters and frame IRQs
// still use the normal implementation. DMC reads and CPU stalls are unchanged.
func (apu *APU) StepBuzzerCycles(cycles int) {
	for cycles > 0 {
		n := cycles
		if apu.frameCounter != 0 && n > int(apu.frameCounter) {
			n = int(apu.frameCounter)
		}
		apu.dmc.stepTimerCycles((n + int(apu.cycle&1)) / 2)
		apu.cycle += uint32(n)
		apu.frameCounter -= uint32(n)
		if apu.frameCounter == 0 {
			apu.frameCounter = frameCounterPeriod
			apu.stepFrameCounter()
		}
		cycles -= n
	}
}
