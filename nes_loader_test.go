package main

import "testing"

func testROM(mapper, prg, chr byte, trainer bool) []byte {
	offset := 16
	if trainer {
		offset += 512
	}
	d := make([]byte, offset+int(prg)*16384+int(chr)*8192)
	copy(d, "NES\x1a")
	d[4], d[5], d[6], d[7] = prg, chr, mapper<<4, mapper&0xf0
	if trainer {
		d[6] |= 4
		d[16] = 0xee
	}
	if prg > 0 {
		d[offset] = 0xea
		d[offset+int(prg)*16384-4] = 0x34
		d[offset+int(prg)*16384-3] = 0x92
	}
	return d
}

func TestROMLoading(t *testing.T) {
	for _, trainer := range []bool{false, true} {
		c, err := NewConsoleFromString(string(testROM(0, 1, 0, trainer)))
		if err != nil {
			t.Fatal(err)
		}
		if c.Cartridge.PRG[0] != 0xea || len(c.Cartridge.CHR) != 8192 || c.CPU.PC != 0x9234 {
			t.Fatal("incorrect ROM layout or reset vector")
		}
		c.Mapper.Write(0x123, 0x56)
		if c.Mapper.Read(0x123) != 0x56 {
			t.Fatal("CHR RAM is not writable")
		}
	}
}

func TestInvalidROM(t *testing.T) {
	nes2 := testROM(0, 1, 1, false)
	nes2[7] |= 8
	truncated := testROM(0, 1, 1, true)
	for name, d := range map[string][]byte{
		"short": {}, "magic": make([]byte, 16), "no PRG": testROM(0, 0, 1, false),
		"truncated": truncated[:len(truncated)-1], "NES2": nes2,
		"mapper": testROM(99, 1, 1, false), "AxROM size": testROM(7, 1, 0, false),
		"mapper40 size": testROM(40, 1, 1, false),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewConsoleFromString(string(d)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAxROMResetVector(t *testing.T) {
	d := testROM(7, 4, 0, false)
	d[16+0x7ffc], d[16+0x7ffd] = 0x78, 0x86
	c, err := NewConsoleFromString(string(d))
	if err != nil {
		t.Fatal(err)
	}
	if c.CPU.PC != 0x8678 {
		t.Fatalf("PC = %04x", c.CPU.PC)
	}
}
