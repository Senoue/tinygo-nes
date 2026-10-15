package nes

import "testing"

func TestCPURAMMirrorsAndStackWrap(t *testing.T) {
	c := frameIRQTestConsole(t)
	for address := uint16(0); address < 0x2000; address++ {
		c.CPU.Write(address, byte(address>>3))
		for base := uint16(0); base < 0x2000; base += 0x800 {
			if got := c.CPU.Read(base | (address & 0x7ff)); got != byte(address>>3) {
				t.Fatalf("RAM mirror at %04x: %02x", address, got)
			}
		}
	}
	c.CPU.SP = 0
	c.CPU.push16(0x1234)
	if c.CPU.SP != 0xfe || c.RAM[0x100] != 0x12 || c.RAM[0x1ff] != 0x34 {
		t.Fatal("stack push did not wrap within page 1")
	}
	if got := c.CPU.pull16(); got != 0x1234 || c.CPU.SP != 0 {
		t.Fatalf("stack pull: %04x, SP=%02x", got, c.CPU.SP)
	}
}

func TestCPUZeroPageIndirectWrap(t *testing.T) {
	for _, opcode := range []byte{0xa1, 0xb1} { // LDA ($fe,X), LDA ($ff),Y
		c := frameIRQTestConsole(t)
		operand := byte(0xfe)
		c.CPU.X = 1
		if opcode == 0xb1 {
			operand = 0xff
			c.CPU.Y = 2 // Also cross a page, requiring the extra CPU cycle.
		}
		copy(c.Cartridge.PRG, []byte{opcode, operand})
		c.RAM[0xff], c.RAM[0], c.RAM[0x100] = 0xff, 0x12, 0x99
		c.RAM[0x2ff], c.RAM[0x301] = 0x42, 0x81
		want := byte(0x42)
		if opcode == 0xb1 {
			want = 0x81
		}
		if cycles := c.CPU.Step(); cycles != 6 || c.CPU.A != want || c.CPU.PC != 0x8002 {
			t.Fatalf("opcode %02x: cycles=%d A=%02x PC=%04x", opcode, cycles, c.CPU.A, c.CPU.PC)
		}
	}
}

func TestCPUMemoryDeviceSideEffects(t *testing.T) {
	c := frameIRQTestConsole(t)
	// PPU registers are mirrored through $3FFF; status reads clear vblank.
	c.CPU.Write(0x3ffb, 7)
	c.CPU.Write(0x3ffc, 0xab)
	if c.PPU.oamData[7] != 0xab {
		t.Fatal("mirrored OAM write lost")
	}
	c.PPU.nmiOccurred, c.PPU.w = true, 1
	if c.CPU.Read(0x3ffa)&0x80 == 0 || c.PPU.nmiOccurred || c.PPU.w != 0 {
		t.Fatal("mirrored PPU status read lost its side effects")
	}
	c.APU.frameIRQPending = true
	if c.CPU.Read(0x4015)&0x40 == 0 || c.APU.frameIRQPending {
		t.Fatal("APU status read did not acknowledge IRQ")
	}
	c.APU.frameIRQPending = true
	c.CPU.Write(0x4017, 0x40)
	if c.APU.frameIRQPending {
		t.Fatal("APU IRQ inhibit write lost")
	}
	c.Controller1.buttons[ButtonA] = true
	c.Controller2.buttons[ButtonB] = true
	c.CPU.Write(0x4016, 1)
	c.CPU.Write(0x4016, 0)
	if c.CPU.Read(0x4016) != 1 || c.CPU.Read(0x4016) != 0 ||
		c.CPU.Read(0x4017) != 0 || c.CPU.Read(0x4017) != 1 {
		t.Fatal("controller strobe or serial reads changed")
	}
	// RAM-sourced DMA still copies the entire page and stalls the CPU.
	for i := 0; i < 256; i++ {
		c.RAM[0x200+i] = byte(i)
	}
	c.CPU.Write(0x2003, 0)
	c.CPU.Write(0x4014, 2)
	for i, got := range c.PPU.oamData {
		if got != byte(i) {
			t.Fatalf("DMA byte %d: %02x", i, got)
		}
	}
	if c.CPU.stall != 513 {
		t.Fatalf("DMA stall = %d", c.CPU.stall)
	}
}

func TestCPUCartridgeMemoryBoundaries(t *testing.T) {
	c := outputTestConsole(t, 2)
	for bank := 0; bank < 4; bank++ {
		for i := 0; i < 0x4000; i++ {
			c.Cartridge.PRG[bank*0x4000+i] = byte(bank + 1)
		}
	}
	for _, address := range []uint16{0x6000, 0x7fff} {
		c.CPU.Write(address, 0xa5)
		if c.CPU.Read(address) != 0xa5 {
			t.Fatalf("SRAM at %04x", address)
		}
	}
	c.CPU.Write(0x8000, 2)
	for _, address := range []uint16{0x8000, 0xbfff, 0xc000, 0xffff} {
		want := byte(3)
		if address >= 0xc000 {
			want = 4
		}
		if got := c.CPU.Read(address); got != want {
			t.Fatalf("PRG at %04x: %d != %d", address, got, want)
		}
	}
	if got := c.CPU.Read16(0xbfff); got != 0x0403 {
		t.Fatalf("bank boundary: %04x", got)
	}
	c.RAM[0] = 0x56
	if got := c.CPU.Read16(0xffff); got != 0x5604 {
		t.Fatalf("address wrap: %04x", got)
	}
	for _, address := range []uint16{0x4018, 0x5fff} {
		c.CPU.Write(address, 0xff)
		if c.CPU.Read(address) != 0 {
			t.Fatalf("unmapped address %04x changed", address)
		}
	}
}
