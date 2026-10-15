package nes

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDirectCPUMapperReadsMatchGeneralReads(t *testing.T) {
	for _, mapper := range []byte{0, 2, 3} {
		for _, size := range []int{0x4000, 0x8000, 0x10000} {
			t.Run(fmt.Sprintf("mapper%d/PRG%d", mapper, size), func(t *testing.T) {
				prg := make([]byte, size)
				for i := range prg {
					prg[i] = byte(i*37 + i/256 + i/0x4000*13)
				}
				c, err := NewConsole(NewCartridge(prg, make([]byte, 0x8000), mapper, 0, 0))
				if err != nil {
					t.Fatal(err)
				}
				for i := range c.Cartridge.SRAM {
					c.Cartridge.SRAM[i] = byte(i*7 + i/256)
				}
				for _, value := range []byte{0, 1, 2, 3, 127, 255} {
					c.CPU.Write(0x8000, value)
					for address := 0x6000; address <= 0xffff; address++ {
						got, want := c.CPU.Read(uint16(address)), c.Mapper.Read(uint16(address))
						if got != want {
							t.Fatalf("bank write %02x, address %04x: %02x != %02x", value, address, got, want)
						}
					}
					// A cached or copied ROM must not hide changes in the backing data.
					c.Cartridge.PRG[0]++
				}
			})
		}
	}
}

// Hide the concrete mapper type so CPU reads use the original general path.
type generalReadMapper struct{ Mapper }

func TestDirectCPUMapperFrameEquivalence(t *testing.T) {
	for _, mapper := range []byte{0, 2, 3} {
		t.Run(fmt.Sprint(mapper), func(t *testing.T) {
			slow, fast := outputTestConsole(t, mapper), outputTestConsole(t, mapper)
			for _, c := range []*Console{slow, fast} {
				chr := make([]byte, 0x8000)
				for i := range chr {
					chr[i] = byte(i*17 + i/0x2000*29)
				}
				c.Cartridge.CHR = chr
				// Switch banks while executing; every PRG bank contains the loop.
				program := []byte{
					0xa2, 0x00, // LDX #0
					0xe8, 0x8a, // INX; TXA
					0x8d, 0x00, 0x80, // STA $8000 (PRG or CHR bank)
					0xad, 0x00, 0x80, // LDA $8000
					0xad, 0x00, 0xc0, // LDA $C000
					0x48, 0x68, // PHA; PLA
					0x8d, 0x00, 0x60, // STA $6000
					0xad, 0x15, 0x40, // LDA $4015
					0x4c, 0x02, 0x80, // JMP loop
				}
				for bank := 0; bank < len(c.Cartridge.PRG); bank += 0x4000 {
					copy(c.Cartridge.PRG[bank:], program)
				}
			}
			original := slow.Mapper
			for frame := 0; frame < 8; frame++ {
				slow.Mapper = &generalReadMapper{original}
				slow.SetFrameOutput(frame&1 == 0)
				fast.SetFrameOutput(frame&1 == 0)
				want, got := slow.StepFrame(), fast.StepFrame()
				slow.Mapper = original
				if got != want || !reflect.DeepEqual(slow, fast) {
					t.Fatalf("frame %d: direct reads changed execution or frame output", frame)
				}
			}
		})
	}
}
