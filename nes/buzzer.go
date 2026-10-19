package nes

// BuzzerTone is a monophonic approximation for a hardware PWM buzzer. It uses
// register/envelope state, not the instantaneous phase of APU.Output().
type BuzzerTone struct {
	Frequency uint32 // Hz; zero means silence
	Volume    byte   // 0..15; PWM pulse width approximates loudness
}

// BuzzerTone chooses pulse 1, then pulse 2, then triangle. This preserves one
// melodic voice without requiring a real-time PCM stream from a slow emulator.
// Noise, DMC, original duty cycles and simultaneous voices are not reproduced.
func (apu *APU) BuzzerTone() BuzzerTone {
	if tone := apu.pulse1.buzzerTone(); tone.Volume != 0 {
		return tone
	}
	if tone := apu.pulse2.buzzerTone(); tone.Volume != 0 {
		return tone
	}
	t := &apu.triangle
	if t.enabled && t.lengthValue != 0 && t.counterValue != 0 && t.timerPeriod >= 3 && t.timerPeriod <= 0x7ff {
		return BuzzerTone{Frequency: CPUFrequency / (32 * (uint32(t.timerPeriod) + 1)), Volume: 8}
	}
	return BuzzerTone{}
}

func (p *Pulse) buzzerTone() BuzzerTone {
	if !p.enabled || p.lengthValue == 0 || p.timerPeriod < 8 || p.timerPeriod > 0x7ff {
		return BuzzerTone{}
	}
	volume := p.constantVolume
	if p.envelopeEnabled {
		volume = p.envelopeVolume
	}
	if volume == 0 {
		return BuzzerTone{}
	}
	return BuzzerTone{Frequency: CPUFrequency / (16 * (uint32(p.timerPeriod) + 1)), Volume: volume}
}
