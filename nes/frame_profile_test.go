//go:build perf

package nes

import (
	"reflect"
	"testing"
)

func TestProfileFramePreservesExecution(t *testing.T) {
	for name, mapper := range map[string]byte{"NROM": 0, "MMC3": 4, "Mapper40": 40} {
		t.Run(name, func(t *testing.T) {
			makeConsole := func() *Console {
				prg := make([]byte, 65536)
				for bank := 0; bank < len(prg); bank += 8192 {
					// JMP $8000, with valid interrupt/reset vectors in each bank.
					copy(prg[bank:], []byte{0x4c, 0x00, 0x80})
					for vector := bank + 8192 - 6; vector < bank+8192; vector += 2 {
						prg[vector], prg[vector+1] = 0, 0x80
					}
				}
				c, err := NewConsole(NewCartridge(prg, make([]byte, 8192), mapper, 0, 0))
				if err != nil {
					t.Fatal(err)
				}
				c.PPU.flagShowBackground = 1
				c.PPU.flagShowSprites = 1
				c.CPU.Write(0x4015, 0x0f)
				if mapper == 4 {
					c.Mapper.Write(0xc000, 3)
					c.Mapper.Write(0xe001, 0)
				}
				return c
			}
			normal, profiled := makeConsole(), makeConsole()
			for i := 0; i < 3; i++ {
				normal.SetFrameOutput(i&1 != 0)
				profiled.SetFrameOutput(i&1 != 0)
				cycles := normal.StepFrame()
				p := profiled.ProfileFrame()
				if p.CPUCycles != cycles {
					t.Fatalf("cycles = %d, want %d", p.CPUCycles, cycles)
				}
				// DeepEqual follows cyclic console pointers and compares the complete
				// CPU/APU/PPU state, framebuffer, cartridge memory and mapper state.
				if !reflect.DeepEqual(normal, profiled) {
					t.Fatalf("state differs after frame %d", i)
				}
				if p.Scanline != (mapper != 40) {
					t.Fatal("incorrect profiling mode")
				}
				if mapper == 40 && (p.CPUAPU != 0 || p.PPU != 0) {
					t.Fatal("cycle-based path must report total time only")
				}
				if mapper != 40 && (p.CPUAPU <= 0 || p.PPU <= 0) {
					t.Fatal("missing scanline section timings")
				}
				if p.Total < p.CPUAPU+p.PPU {
					t.Fatal("section times exceed total")
				}
			}
		})
	}
}
