//go:build buzzer && apulite

package nes

// BuzzerAPULite reports whether console execution omits unused waveform state.
const BuzzerAPULite = true

func (apu *APU) stepForConsole(cycles int) { apu.StepBuzzerCycles(cycles) }
