package nes

import (
	"reflect"
	"testing"
)

func outputTestConsole(t testing.TB, mapper byte) *Console {
	t.Helper()
	prg := make([]byte, 65536)
	for bank := 0; bank < len(prg); bank += 8192 {
		copy(prg[bank:], []byte{0x4c, 0, 0x80})
		for vector := bank + 8192 - 6; vector < bank+8192; vector += 2 {
			prg[vector], prg[vector+1] = 0, 0x80
		}
	}
	c, err := NewConsole(NewCartridge(prg, make([]byte, 8192), mapper, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	for i := range c.Cartridge.CHR {
		c.Cartridge.CHR[i] = byte(i*37 + 13)
	}
	for i := range c.PPU.nameTableData {
		c.PPU.nameTableData[i] = byte(i * 7)
	}
	for i := range c.PPU.paletteData {
		c.PPU.paletteData[i] = byte(i)
	}
	for i := 0; i < 64; i++ {
		c.PPU.oamData[i*4] = byte(i * 3)
		c.PPU.oamData[i*4+1] = byte(i)
		c.PPU.oamData[i*4+2] = byte(i * 31)
		c.PPU.oamData[i*4+3] = byte(i * 11)
	}
	c.PPU.writeMask(0x1e)
	c.PPU.writeControl(0x80)
	c.PPU.x = 5
	c.PPU.v = 0x73bf
	c.PPU.t = 0x73bf
	if mapper == 4 {
		c.Mapper.Write(0xc000, 2)
		c.Mapper.Write(0xe001, 0)
		c.CPU.I = 0
	}
	return c
}

// Ignore only host output state. CPU/APU, all PPU registers and sprite state,
// mapper state, interrupts and emulated memory must still match.
func assertOutputStateEqual(t *testing.T, normal, skipped *Console) {
	t.Helper()
	p := skipped.PPU
	front, scratch, skip := p.front, p.backgroundColors, p.skipPixels
	p.front, p.backgroundColors, p.skipPixels = normal.PPU.front, normal.PPU.backgroundColors, normal.PPU.skipPixels
	equal := reflect.DeepEqual(normal, skipped)
	p.front, p.backgroundColors, p.skipPixels = front, scratch, skip
	if !equal {
		t.Fatal("pixel skipping changed emulated state")
	}
}

func TestFrameOutputPreservesExecution(t *testing.T) {
	for name, mapper := range map[string]byte{"NROM": 0, "MMC1": 1, "UxROM": 2, "CNROM": 3, "MMC3": 4, "AxROM": 7, "Mapper40": 40, "Mapper225": 225} {
		t.Run(name, func(t *testing.T) {
			for name, interval := range map[string]int{"every2": 2, "every4": 4} {
				t.Run(name, func(t *testing.T) {
					normal, skipped := outputTestConsole(t, mapper), outputTestConsole(t, mapper)
					for frame := 1; frame <= interval*2; frame++ {
						render := frame%interval == 0
						skipped.SetFrameOutput(render)
						before := *skipped.Buffer()
						cycles := normal.StepFrame()
						if got := skipped.StepFrame(); got != cycles {
							t.Fatalf("CPU cycles = %d, want %d", got, cycles)
						}
						assertOutputStateEqual(t, normal, skipped)
						if render {
							if *normal.Buffer() != *skipped.Buffer() {
								t.Fatal("next displayed frame differs")
							}
						} else if *skipped.Buffer() != before {
							t.Fatal("skipped frame wrote pixels")
						}
					}
				})
			}
		})
	}
}

func TestSkippedScanlineMatchesRegisters(t *testing.T) {
	for _, fineX := range []byte{0, 1, 7} {
		for _, mask := range []byte{0, 8, 16, 24, 30} {
			normal, skipped := outputTestConsole(t, 4), outputTestConsole(t, 4)
			for _, c := range []*Console{normal, skipped} {
				c.PPU.ScanLine = 0
				c.PPU.x = fineX
				c.PPU.writeMask(mask)
			}
			skipped.SetFrameOutput(false)
			for line := 0; line < 262; line++ {
				normal.PPU.RenderScanline()
				skipped.PPU.RenderScanline()
				assertOutputStateEqual(t, normal, skipped)
			}
		}
	}
}

func TestSpriteZeroWithoutFramebuffer(t *testing.T) {
	cases := []struct {
		name                string
		x, fineX, mask, low byte
		coarse              uint16
		flip                bool
		want                byte
	}{
		{"opaque same palette color", 16, 0, 30, 0xff, 0, false, 1},
		{"transparent pattern", 16, 0, 30, 0, 0, false, 0},
		{"background disabled", 16, 0, 20, 0xff, 0, false, 0},
		{"sprites disabled", 16, 0, 10, 0xff, 0, false, 0},
		{"background clipped", 0, 0, 28, 0xff, 0, false, 0},
		{"sprites clipped", 0, 0, 26, 0xff, 0, false, 0},
		{"last pixel excluded", 255, 0, 30, 0xff, 0, false, 0},
		{"penultimate pixel", 254, 0, 30, 0xff, 0, false, 1},
		{"fine scroll and nametable wrap", 7, 7, 30, 0xff, 31, true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, skip := range []bool{false, true} {
				c := outputTestConsole(t, 0)
				p := c.PPU
				p.ScanLine, p.v, p.x = 0, tc.coarse, tc.fineX
				p.writeMask(tc.mask)
				p.paletteData = [32]byte{} // Equal display colors must not suppress hits.
				p.nameTableData = [2048]byte{}
				for i := range c.Cartridge.CHR {
					c.Cartridge.CHR[i] = 0
				}
				c.Cartridge.CHR[0] = tc.low // Background tile 0, row 0.
				c.Cartridge.CHR[16] = 0xff  // Sprite tile 1, row 0.
				for i := 0; i < 64; i++ {
					p.oamData[i*4] = 240
				}
				p.oamData[0], p.oamData[1], p.oamData[2], p.oamData[3] = 0, 1, 0, tc.x
				if tc.flip {
					p.oamData[2] = 0x40
				}
				c.SetFrameOutput(!skip)
				p.RenderScanline()
				if p.flagSpriteZeroHit != tc.want {
					t.Fatalf("skip=%v hit=%d want=%d", skip, p.flagSpriteZeroHit, tc.want)
				}
			}
		})
	}
}

func BenchmarkFrameOutput(b *testing.B) {
	for _, render := range []bool{true, false} {
		name := "draw"
		if !render {
			name = "skip"
		}
		b.Run(name, func(b *testing.B) {
			prg := make([]byte, 32768)
			copy(prg, []byte{0x4c, 0, 0x80})
			prg[0x7ffd] = 0x80
			c, err := NewConsole(NewCartridge(prg, make([]byte, 8192), 0, 0, 0))
			if err != nil {
				b.Fatal(err)
			}
			c.PPU.writeMask(30)
			c.SetFrameOutput(render)
			c.StepFrame() // Finish startup's partial frame.
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.StepFrame()
			}
		})
	}
}

func TestScanlineSpritePriorityUsesBackgroundPattern(t *testing.T) {
	c := outputTestConsole(t, 0)
	p := c.PPU
	p.spriteCount = 2
	p.spritePositions[0], p.spritePositions[1] = 16, 16
	p.spritePatterns[0], p.spritePatterns[1] = 0x10000000, 0x20000000
	p.spritePriorities[0], p.spritePriorities[1] = 1, 0
	p.spriteIndexes[0], p.spriteIndexes[1] = 0, 1
	p.paletteData[0], p.paletteData[1], p.paletteData[0x12] = 5, 5, 7
	p.backgroundColors[16] = 1 // Opaque background has the same displayed color as the backdrop.
	p.front[16] = 5
	p.renderScanlineSprites(0)
	if p.front[16] != 5 || p.flagSpriteZeroHit != 1 {
		t.Fatal("lower-index sprite must hide higher sprite behind opaque background and trigger hit")
	}
	p.flagSpriteZeroHit = 0
	p.backgroundColors[16] = 0
	p.paletteData[0x11] = 9
	p.renderScanlineSprites(0)
	if p.front[16] != 9 || p.flagSpriteZeroHit != 0 {
		t.Fatal("sprite overlap alone must not trigger sprite-zero hit")
	}
}

func TestDisabledRenderingClearsDisplayedFrame(t *testing.T) {
	c := outputTestConsole(t, 0)
	c.PPU.writeMask(0)
	c.PPU.ScanLine = 0
	c.PPU.paletteData[0] = 12
	for i := range c.PPU.front {
		c.PPU.front[i] = 3
	}
	c.SetFrameOutput(false)
	c.StepFrame()
	if c.PPU.front[0] != 3 {
		t.Fatal("disabled output changed framebuffer")
	}
	c.SetFrameOutput(true)
	c.StepFrame()
	for _, pixel := range c.PPU.front {
		if pixel != 12 {
			t.Fatal("disabled rendering left stale pixels in displayed frame")
		}
	}
}

func TestSkippedSpriteZeroAcrossNametables(t *testing.T) {
	for _, skip := range []bool{false, true} {
		c := outputTestConsole(t, 0)
		p := c.PPU
		c.Cartridge.Mirror = MirrorVertical
		p.ScanLine, p.v, p.x = 0, 31, 7
		p.nameTableData = [2048]byte{}
		p.nameTableData[1024] = 2 // First tile in the adjacent horizontal nametable.
		for i := range c.Cartridge.CHR {
			c.Cartridge.CHR[i] = 0
		}
		c.Cartridge.CHR[2*16+8] = 0x80 // Background high plane only, first pixel.
		c.Cartridge.CHR[16] = 0x01     // Sprite is horizontally flipped below.
		for i := 0; i < 64; i++ {
			p.oamData[i*4] = 240
		}
		p.oamData[0], p.oamData[1], p.oamData[2], p.oamData[3] = 0, 1, 0x40, 1
		c.SetFrameOutput(!skip)
		p.RenderScanline()
		if p.flagSpriteZeroHit != 1 {
			t.Fatalf("skip=%v: missing hit across nametable boundary", skip)
		}
	}
}
