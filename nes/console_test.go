package nes

import "testing"

func TestMapper4ScanlineIRQ(t *testing.T) {
	c, err := NewConsole(NewCartridge(make([]byte, 32768), make([]byte, 8192), 4, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	c.CPU.I = 0
	c.PPU.ScanLine = 0
	c.PPU.flagShowBackground = 1
	c.Mapper.Write(0xc000, 1)
	c.Mapper.Write(0xe001, 0)
	c.PPU.RenderScanline()
	c.PPU.RenderScanline()
	if c.CPU.interrupt != interruptIRQ {
		t.Fatal("scanlines did not trigger mapper IRQ")
	}
}

func TestMapper40FrameClock(t *testing.T) {
	prg := make([]byte, 65536)
	for i := range prg {
		prg[i] = 0xea
	}
	prg[0xfffc], prg[0xfffd] = 0, 0x80
	c, err := NewConsole(NewCartridge(prg, make([]byte, 8192), 40, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	cycles := c.StepFrame()
	m := c.Mapper.(*Mapper40)
	if m.cycles != cycles*3%(4096*3) {
		t.Fatalf("mapper cycles = %d, CPU cycles = %d", m.cycles, cycles)
	}
}
