package nes

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// Compare the complete console, ignoring only the eight waveform fields that
// the lightweight mode intentionally does not advance. Do not ignore DMC,
// IRQ, envelope, length, linear counter, sweep or CPU/PPU state.
func assertBuzzerConsoleEqual(t *testing.T, full, lite *Console) {
	t.Helper()
	a, b := *full.APU, *lite.APU
	if a.BuzzerTone() != b.BuzzerTone() || a.readStatus() != b.readStatus() {
		t.Fatal("lightweight APU changed buzzer tone or CPU-visible status")
	}
	saved := *lite.APU
	b = saved
	b.pulse1.timerValue, b.pulse1.dutyValue = a.pulse1.timerValue, a.pulse1.dutyValue
	b.pulse2.timerValue, b.pulse2.dutyValue = a.pulse2.timerValue, a.pulse2.dutyValue
	b.triangle.timerValue, b.triangle.dutyValue = a.triangle.timerValue, a.triangle.dutyValue
	b.noise.timerValue, b.noise.shiftRegister = a.noise.timerValue, a.noise.shiftRegister
	*lite.APU = b
	equal := reflect.DeepEqual(full, lite)
	*lite.APU = saved
	if !equal {
		t.Fatal("lightweight APU changed state beyond unused waveform fields")
	}
}

func TestBuzzerAPURegisterSequence(t *testing.T) {
	for _, seed := range []int64{1, 42, 927} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			full, lite := outputTestConsole(t, 0), outputTestConsole(t, 0)
			rng := rand.New(rand.NewSource(seed))
			for i := 0; i < 5000; i++ {
				if i%3 == 0 {
					address, value := uint16(0x4000+rng.Intn(24)), byte(rng.Intn(256))
					full.CPU.Write(address, value)
					lite.CPU.Write(address, value)
				}
				cycles := rng.Intn(16)
				if i%127 == 0 {
					cycles = frameCounterPeriod*2 + 17
				}
				full.APU.StepCycles(cycles)
				lite.APU.StepBuzzerCycles(cycles)
				assertBuzzerConsoleEqual(t, full, lite)
				if i%5 == 0 && full.CPU.Read(0x4015) != lite.CPU.Read(0x4015) {
					t.Fatal("APU status read differs")
				}
			}
		})
	}
}

func TestBuzzerAPUDMCAndCounterBoundaries(t *testing.T) {
	for _, loop := range []bool{false, true} {
		for _, period := range []byte{0, 27, 214} {
			full, lite := outputTestConsole(t, 0), outputTestConsole(t, 0)
			fullTrace, liteTrace := &apuReadTrace{}, &apuReadTrace{}
			full.Mapper, lite.Mapper = fullTrace, liteTrace
			for _, c := range []*Console{full, lite} {
				a := c.APU
				a.cycle = ^uint32(0) - 3
				a.frameCounter, a.frameValue, a.frameIRQ = 2, 2, true
				a.pulse1.sweepEnabled, a.pulse1.sweepReload = true, true
				a.pulse1.timerPeriod, a.pulse1.sweepShift = 253, 1
				a.triangle.lengthValue, a.triangle.counterValue = 1, 1
				a.triangle.lengthEnabled = true
				a.dmc.enabled, a.dmc.loop = true, loop
				a.dmc.currentAddress, a.dmc.currentLength = 0xfffe, 4
				a.dmc.sampleAddress, a.dmc.sampleLength = 0xc010, 2
				a.dmc.tickPeriod, a.dmc.tickValue = period, 1
			}
			for _, cycles := range []int{0, 1, 1, 2, 3, 255, 256, 4096, 15000} {
				full.APU.StepCycles(cycles)
				lite.APU.StepBuzzerCycles(cycles)
				assertBuzzerConsoleEqual(t, full, lite)
				if !reflect.DeepEqual(fullTrace.addresses, liteTrace.addresses) {
					t.Fatal("DMC read order differs")
				}
				if lite.CPU.stall != len(liteTrace.addresses)*4 {
					t.Fatal("DMC stall count differs")
				}
			}
			if len(liteTrace.addresses) < 4 || !reflect.DeepEqual(liteTrace.addresses[:4], []uint16{0xfffe, 0xffff, 0x8000, 0x8001}) {
				t.Fatalf("DMC wrap not exercised: %v", liteTrace.addresses)
			}
		}
	}
}

func BenchmarkBuzzerAPU(b *testing.B) {
	for _, dmc := range []bool{false, true} {
		for _, lite := range []bool{false, true} {
			b.Run(fmt.Sprintf("dmc=%v/lite=%v", dmc, lite), func(b *testing.B) {
				c := outputTestConsole(b, 0)
				a := c.APU
				// Fixed register order gives identical starting conditions.
				for _, reg := range [][2]uint16{{0x4015, 15}, {0x4000, 0xbf}, {0x4002, 200}, {0x4003, 8}, {0x4004, 0x7f}, {0x4006, 120}, {0x4007, 8}, {0x4008, 0xff}, {0x400a, 100}, {0x400b, 8}, {0x400c, 0x3f}, {0x400e, 8}, {0x400f, 8}} {
					a.writeRegister(reg[0], byte(reg[1]))
				}
				if dmc {
					a.writeRegister(0x4010, 0x4f)
					a.writeRegister(0x4012, 0)
					a.writeRegister(0x4013, 20)
					a.writeRegister(0x4015, 31)
				}
				b.ResetTimer()
				if lite {
					for i := 0; i < b.N; i++ {
						a.StepBuzzerCycles(2 + i%6)
					}
				} else {
					for i := 0; i < b.N; i++ {
						a.StepCycles(2 + i%6)
					}
				}
			})
		}
	}
}
