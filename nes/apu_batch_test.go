package nes

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func assertAPUBatchEqual(t *testing.T, slow, fast *Console) {
	t.Helper()
	a, b := *slow.APU, *fast.APU
	a.console, b.console, a.dmc.cpu, b.dmc.cpu = nil, nil, nil, nil
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("APU mismatch\nStep: %+v\nBatch: %+v", a, b)
	}
	if slow.CPU.stall != fast.CPU.stall || slow.CPU.interrupt != fast.CPU.interrupt {
		t.Fatalf("CPU effects differ: stall %d/%d, interrupt %d/%d", slow.CPU.stall, fast.CPU.stall, slow.CPU.interrupt, fast.CPU.interrupt)
	}
	if slow.APU.Output() != fast.APU.Output() || slow.APU.readStatus() != fast.APU.readStatus() {
		t.Fatal("APU output or status differs")
	}
}

func TestAPUBatchMatchesCycleSteps(t *testing.T) {
	for _, seed := range []int64{1, 42, 914} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			slow, fast := outputTestConsole(t, 0), outputTestConsole(t, 0)
			slow.CPU.I, fast.CPU.I = 0, 0
			rng := rand.New(rand.NewSource(seed))
			for i := 0; i < 4000; i++ {
				if i%3 == 0 {
					address := uint16(0x4000 + rng.Intn(24))
					value := byte(rng.Intn(256))
					slow.APU.writeRegister(address, value)
					fast.APU.writeRegister(address, value)
				}
				cycles := rng.Intn(15) // CPU instructions and DMC stall-sized batches.
				if i%127 == 0 {
					cycles = frameCounterPeriod*2 + 17
				}
				for j := 0; j < cycles; j++ {
					slow.APU.Step()
				}
				fast.APU.StepCycles(cycles)
				assertAPUBatchEqual(t, slow, fast)
			}
		})
	}
}

func TestAPUBatchBoundaries(t *testing.T) {
	for _, looping := range []bool{false, true} {
		for _, period := range []byte{0, 27, 214} {
			slow, fast := outputTestConsole(t, 0), outputTestConsole(t, 0)
			for _, c := range []*Console{slow, fast} {
				c.CPU.I = 0
				a := c.APU
				a.cycle = ^uint32(0) - 3
				a.frameCounter, a.frameValue, a.frameIRQ = 2, 3, true
				a.pulse1.timerPeriod, a.pulse1.timerValue = 65535, 1
				a.pulse1.sweepEnabled, a.pulse1.sweepReload = true, true
				a.pulse1.sweepShift, a.pulse1.sweepPeriod = 1, 1
				a.triangle.lengthValue, a.triangle.counterValue = 1, 1
				a.triangle.lengthEnabled = true
				a.dmc.enabled, a.dmc.loop = true, looping
				a.dmc.currentAddress, a.dmc.currentLength = 0xffff, 3
				a.dmc.sampleAddress, a.dmc.sampleLength = 0xc000, 1
				a.dmc.tickPeriod, a.dmc.tickValue = period, 1
			}
			for _, cycles := range []int{0, 1, 1, 2, 3, 255, 256, 4096, 15000} {
				for i := 0; i < cycles; i++ {
					slow.APU.Step()
				}
				fast.APU.StepCycles(cycles)
				assertAPUBatchEqual(t, slow, fast)
			}
		}
	}
}

func TestAdvanceTimerBoundaries(t *testing.T) {
	for _, value := range []uint16{0, 1, 7, 65535} {
		for _, period := range []uint16{0, 1, 8, 65535} {
			for _, ticks := range []int{0, 1, 2, 7, 8, 65536, 65537} {
				want, count := value, 0
				for i := 0; i < ticks; i++ {
					if want == 0 {
						want = period
						count++
					} else {
						want--
					}
				}
				got, expirations := advanceTimer(value, period, ticks)
				if got != want || expirations != count {
					t.Fatalf("value=%d period=%d ticks=%d: got %d/%d want %d/%d", value, period, ticks, got, expirations, want, count)
				}
			}
		}
	}
}

func BenchmarkAPUStepCycles(b *testing.B) {
	for _, mode := range []string{"idle", "music", "dmc"} {
		for _, batch := range []bool{false, true} {
			name := mode + "/cycle"
			if batch {
				name = mode + "/batch"
			}
			b.Run(name, func(b *testing.B) {
				c, err := NewConsole(NewCartridge(make([]byte, 32768), make([]byte, 8192), 0, 0, 0))
				if err != nil {
					b.Fatal(err)
				}
				a := c.APU
				if mode != "idle" {
					for address, value := range map[uint16]byte{0x4015: 15, 0x4000: 0x9f, 0x4002: 200, 0x4003: 8, 0x4004: 0x5f, 0x4006: 120, 0x4007: 8, 0x4008: 0xff, 0x400a: 100, 0x400b: 8, 0x400c: 0x1f, 0x400e: 8, 0x400f: 8} {
						a.writeRegister(address, value)
					}
				}
				if mode == "dmc" {
					a.writeRegister(0x4010, 0x4f)
					a.writeRegister(0x4012, 0)
					a.writeRegister(0x4013, 20)
					a.writeRegister(0x4015, 31)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					cycles := 2 + i%6
					if batch {
						a.StepCycles(cycles)
					} else {
						for j := 0; j < cycles; j++ {
							a.Step()
						}
					}
				}
			})
		}
	}
}

// Reference frame scheduler with the original per-cycle APU loop.
func stepFrameCycleAPU(c *Console) int {
	frame, cycles := c.PPU.Frame, 0
	for c.PPU.Frame == frame {
		if c.Cartridge.Mapper == 40 {
			n := c.CPU.Step()
			cycles += n
			for i := 0; i < n*3; i++ {
				c.PPU.Step()
				if c.mapperHasStep {
					c.Mapper.Step()
				}
			}
			for i := 0; i < n; i++ {
				c.APU.Step()
			}
		} else {
			line := 0
			for line < 114 {
				n := c.CPU.Step()
				cycles += n
				line += n
				for i := 0; i < n; i++ {
					c.APU.Step()
				}
			}
			c.PPU.RenderScanline()
		}
	}
	return cycles
}

func TestAPUBatchFrameIntegration(t *testing.T) {
	for _, mapper := range []byte{0, 4, 40} {
		slow, fast := outputTestConsole(t, mapper), outputTestConsole(t, mapper)
		for _, c := range []*Console{slow, fast} {
			program := []byte{
				0xa9, 0x4f, 0x8d, 0x10, 0x40, // Looping DMC.
				0xa9, 0x00, 0x8d, 0x12, 0x40,
				0xa9, 0x01, 0x8d, 0x13, 0x40,
				0xa9, 0x1f, 0x8d, 0x15, 0x40, // Enable after configuring sample length.
				0xe6, 0x00, 0xa5, 0x00, // Change timer and frame control every loop.
				0x8d, 0x02, 0x40, 0x8d, 0x17, 0x40,
				0x4c, 0x14, 0x80,
			}
			for bank := 0; bank < len(c.Cartridge.PRG); bank += 8192 {
				copy(c.Cartridge.PRG[bank:], program)
			}
			c.CPU.I = 0
		}
		for frame := 0; frame < 4; frame++ {
			slow.SetFrameOutput(frame&1 == 0)
			fast.SetFrameOutput(frame&1 == 0)
			want := stepFrameCycleAPU(slow)
			if got := fast.StepFrame(); got != want {
				t.Fatalf("mapper %d cycles %d != %d", mapper, got, want)
			}
			if BuzzerAPULite {
				assertBuzzerConsoleEqual(t, slow, fast)
			} else if !reflect.DeepEqual(slow, fast) {
				t.Fatalf("mapper %d frame %d: batch changed execution", mapper, frame)
			}
		}
	}
}

type apuReadTrace struct{ addresses []uint16 }

func (m *apuReadTrace) Read(address uint16) byte {
	m.addresses = append(m.addresses, address)
	return byte(address)
}
func (*apuReadTrace) Write(uint16, byte) {}
func (*apuReadTrace) Step()              {}

func TestAPUBatchDMCReadSequence(t *testing.T) {
	slow, fast := outputTestConsole(t, 0), outputTestConsole(t, 0)
	slowTrace, fastTrace := &apuReadTrace{}, &apuReadTrace{}
	slow.Mapper, fast.Mapper = slowTrace, fastTrace
	for _, c := range []*Console{slow, fast} {
		d := &c.APU.dmc
		d.enabled, d.loop = true, true
		d.currentAddress, d.currentLength = 0xfffe, 4
		d.sampleAddress, d.sampleLength = 0xc010, 2
		d.tickPeriod = 1
	}
	for i := 0; i < 300; i++ {
		slow.APU.Step()
	}
	fast.APU.StepCycles(300)
	if !reflect.DeepEqual(slowTrace.addresses, fastTrace.addresses) {
		t.Fatal("DMC read ordering changed")
	}
	wantPrefix := []uint16{0xfffe, 0xffff, 0x8000, 0x8001, 0xc010, 0xc011}
	if len(fastTrace.addresses) < len(wantPrefix) || !reflect.DeepEqual(fastTrace.addresses[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("DMC did not wrap and loop as expected: %v", fastTrace.addresses)
	}
	if fast.CPU.stall != 4*len(fastTrace.addresses) {
		t.Fatal("DMC stall count changed")
	}
	assertAPUBatchEqual(t, slow, fast)
}
