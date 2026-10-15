package nes

import (
	"reflect"
	"testing"
)

func TestBuzzerToneSelection(t *testing.T) {
	pulse := Pulse{enabled: true, lengthValue: 1, timerPeriod: 253, constantVolume: 12}
	triangle := Triangle{enabled: true, lengthValue: 1, counterValue: 1, timerPeriod: 253}
	cases := []struct {
		name string
		apu  APU
		want BuzzerTone
	}{
		{"silence", APU{}, BuzzerTone{}},
		{"pulse1", APU{pulse1: pulse}, BuzzerTone{440, 12}},
		{"pulse2 fallback", APU{pulse2: pulse}, BuzzerTone{440, 12}},
		{"triangle fallback", APU{triangle: triangle}, BuzzerTone{220, 8}},
		{"pulse priority", APU{pulse1: pulse, triangle: triangle}, BuzzerTone{440, 12}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := tc.apu
			if got := tc.apu.BuzzerTone(); got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			if !reflect.DeepEqual(before, tc.apu) {
				t.Fatal("tone extraction mutated emulated audio state")
			}
		})
	}
}

func TestBuzzerToneIgnoresInstantaneousPhase(t *testing.T) {
	a := APU{pulse1: Pulse{enabled: true, lengthValue: 1, timerPeriod: 253, envelopeEnabled: true, envelopeVolume: 9}}
	// Duty mode 0 / phase 0 is currently low, but the note must keep sounding.
	if a.pulse1.output() != 0 {
		t.Fatal("test must start in low phase")
	}
	if got := a.BuzzerTone(); got != (BuzzerTone{440, 9}) {
		t.Fatalf("phase gated note: %+v", got)
	}
	a.pulse1.envelopeVolume = 0
	a.pulse1.constantVolume = 15
	if a.BuzzerTone() != (BuzzerTone{}) {
		t.Fatal("zero envelope should mute")
	}
}

func TestBuzzerToneGates(t *testing.T) {
	for _, period := range []uint16{0, 7, 8, 2047, 2048, 65535} {
		a := APU{pulse1: Pulse{enabled: true, lengthValue: 1, timerPeriod: period, constantVolume: 15}}
		got := a.BuzzerTone()
		if period < 8 || period > 2047 {
			if got != (BuzzerTone{}) {
				t.Fatalf("invalid period %d sounded", period)
			}
		} else if got.Frequency != CPUFrequency/(16*(uint32(period)+1)) || got.Volume != 15 {
			t.Fatal("incorrect valid pulse pitch")
		}
	}
	a := APU{pulse1: Pulse{enabled: true, lengthValue: 1, timerPeriod: 253, constantVolume: 15}}
	a.pulse1.enabled = false
	if a.BuzzerTone() != (BuzzerTone{}) {
		t.Fatal("disabled pulse sounded")
	}
	a.pulse1.enabled = true
	a.pulse1.lengthValue = 0
	if a.BuzzerTone() != (BuzzerTone{}) {
		t.Fatal("expired pulse sounded")
	}
	a.triangle = Triangle{enabled: true, lengthValue: 1, counterValue: 0, timerPeriod: 253}
	if a.BuzzerTone() != (BuzzerTone{}) {
		t.Fatal("triangle with zero linear counter sounded")
	}
	a.triangle.counterValue = 1
	a.triangle.timerPeriod = 2
	if a.BuzzerTone() != (BuzzerTone{}) {
		t.Fatal("ultrasonic triangle sounded")
	}
}

func TestBuzzerToneStableVoicePriority(t *testing.T) {
	a := APU{
		pulse1: Pulse{enabled: true, lengthValue: 1, timerPeriod: 253, constantVolume: 2},
		pulse2: Pulse{enabled: true, lengthValue: 1, timerPeriod: 126, constantVolume: 15},
	}
	if got := a.BuzzerTone(); got != (BuzzerTone{440, 2}) {
		t.Fatal("louder secondary voice stole melody")
	}
	a.pulse1.constantVolume = 0
	if got := a.BuzzerTone(); got != (BuzzerTone{880, 15}) {
		t.Fatalf("fallback failed: %+v", got)
	}
}
