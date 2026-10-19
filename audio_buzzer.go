//go:build wioterminal && buzzer

package main

import (
	"machine"
	"time"

	"github.com/Senoue/tinygo-nes/nes"
)

var (
	buzzerPWM        = machine.TCC0 // WIO_BUZZER/PD11 maps to TCC0 channel 4.
	buzzerChannel    uint8
	buzzerFrequency  uint32
	buzzerVolume     byte
	buzzerMuted      bool
	buzzerSelectHeld bool
)

func InitAudio() {
	// Configure for the lowest frequency first, so SetPeriod can subsequently
	// cover all pulse/triangle pitches without changing the prescaler.
	if err := buzzerPWM.Configure(machine.PWMConfig{Period: 40_000_000}); err != nil {
		panic(err.Error())
	}
	channel, err := buzzerPWM.Channel(machine.WIO_BUZZER)
	if err != nil {
		panic(err.Error())
	}
	buzzerChannel = channel
	buzzerPWM.Set(buzzerChannel, 0)

	// A short two-note startup test distinguishes output wiring/configuration
	// from a ROM that has not yet enabled its sound channels.
	setBuzzerTone(nes.BuzzerTone{Frequency: 523, Volume: 8})
	time.Sleep(100 * time.Millisecond)
	setBuzzerTone(nes.BuzzerTone{})
	time.Sleep(50 * time.Millisecond)
	setBuzzerTone(nes.BuzzerTone{Frequency: 784, Volume: 8})
	time.Sleep(100 * time.Millisecond)
	setBuzzerTone(nes.BuzzerTone{})
	println("audio: built-in buzzer enabled (mono: pulse1 > pulse2 > triangle)")
	if nes.BuzzerAPULite {
		println("audio: lightweight APU enabled (waveform generation skipped)")
	}
}

// OutputAudio updates the note once per emulated frame. Hardware PWM keeps
// sounding during emulation and LCD transfers, without a sample-rate interrupt.
func OutputAudio(apu *nes.APU) {
	if buzzerMuted {
		return
	}
	setBuzzerTone(apu.BuzzerTone())
}

// HandleAudioInput toggles only on a new press; holding Select keeps the state.
func HandleAudioInput(selectPressed bool) {
	if selectPressed && !buzzerSelectHeld {
		buzzerMuted = !buzzerMuted
		if buzzerMuted {
			setBuzzerTone(nes.BuzzerTone{})
			println("audio: muted")
		} else {
			println("audio: unmuted")
		}
	}
	buzzerSelectHeld = selectPressed
}

func setBuzzerTone(tone nes.BuzzerTone) {
	if tone.Frequency == 0 || tone.Volume == 0 {
		if buzzerVolume != 0 {
			buzzerPWM.Set(buzzerChannel, 0)
		}
		buzzerVolume = 0
		return
	}
	if tone.Frequency == buzzerFrequency && tone.Volume == buzzerVolume {
		return
	}
	if tone.Frequency != buzzerFrequency {
		if err := buzzerPWM.SetPeriod(1_000_000_000 / uint64(tone.Frequency)); err != nil {
			panic(err.Error())
		}
		buzzerFrequency = tone.Frequency
	}
	// Limit the high pulse to 1/8 of a period at maximum NES volume.
	// This is a quiet buzzer approximation, not a reconstruction of NES duty.
	buzzerPWM.Set(buzzerChannel, buzzerPWM.Top()*uint32(tone.Volume)/(15*8))
	buzzerVolume = tone.Volume
}
