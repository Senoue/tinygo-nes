package nes

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestBackgroundRendererMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(927))
	for _, mapper := range []byte{0, 1, 2, 3, 4, 7, 40, 225} {
		t.Run(fmt.Sprint(mapper), func(t *testing.T) {
			slow, fast := outputTestConsole(t, mapper), outputTestConsole(t, mapper)
			for mirror := byte(0); mirror <= MirrorFour; mirror++ {
				for fineX := byte(0); fineX < 8; fineX++ {
					for table := byte(0); table < 2; table++ {
						for _, mask := range []byte{0, 8, 10, 24, 30} {
							v := uint16(rng.Intn(0x8000))
							// Exercise crossing into the next nametable often.
							if fineX&1 != 0 {
								v |= 31
							}
							line := rng.Intn(240)
							var palette [32]byte
							for i := range palette {
								palette[i] = byte(rng.Intn(256))
							}
							for _, c := range []*Console{slow, fast} {
								p := c.PPU
								c.Cartridge.Mirror = mirror
								p.v, p.x, p.ScanLine = v, fineX, line
								p.paletteData = palette
								p.flagBackgroundTable = table
								p.writeMask(mask)
								p.evaluateSprites()
							}
							slow.PPU.renderScanlineBackgroundReference(line)
							fast.PPU.renderScanlineBackground(line)
							if !reflect.DeepEqual(slow, fast) {
								t.Fatalf("background differs: mirror=%d fineX=%d table=%d mask=%d v=%04x", mirror, fineX, table, mask, v)
							}
							slow.PPU.renderScanlineSprites(line)
							fast.PPU.renderScanlineSprites(line)
							if !reflect.DeepEqual(slow, fast) {
								t.Fatal("sprite composition or sprite-zero hit differs")
							}
						}
					}
				}
			}
		})
	}
}

func BenchmarkBackgroundRenderer(b *testing.B) {
	for _, fineX := range []byte{0, 7} {
		for _, reference := range []bool{true, false} {
			b.Run(fmt.Sprintf("fineX%d/reference=%v", fineX, reference), func(b *testing.B) {
				c := outputTestConsole(b, 3)
				p := c.PPU
				p.x = fineX
				b.ResetTimer()
				if reference {
					for i := 0; i < b.N; i++ {
						p.v = 0x73bf
						p.renderScanlineBackgroundReference(120)
					}
				} else {
					for i := 0; i < b.N; i++ {
						p.v = 0x73bf
						p.renderScanlineBackground(120)
					}
				}
			})
		}
	}
}

// Retain the previous pixel-by-pixel renderer as the differential test oracle.
func (ppu *PPU) renderScanlineBackgroundReference(scanline int) {
	if ppu.flagShowBackground == 0 {
		// Fill with background color
		bgColor := ppu.readPalette(0) % 64
		for x := 0; x < 256; x++ {
			ppu.backgroundColors[x] = 0
			ppu.front[scanline*256+x] = bgColor
		}
		return
	}

	fineY := (ppu.v >> 12) & 7
	table := uint16(ppu.flagBackgroundTable)

	// We need to render 256 pixels. Due to fine X scroll, we may need 33 tiles.
	// Save v so we can restore coarse X iteration state properly.
	fineX := ppu.x
	pixelX := 0

	for pixelX < 256 {
		// Fetch tile data from current v
		ntAddr := 0x2000 | (ppu.v & 0x0FFF)
		nameTableByte := ppu.Read(ntAddr)

		atAddr := uint16(0x23C0) | (ppu.v & 0x0C00) | ((ppu.v >> 4) & 0x38) | ((ppu.v >> 2) & 0x07)
		shift := ((ppu.v >> 4) & 4) | (ppu.v & 2)
		attribute := ((ppu.Read(atAddr) >> shift) & 3) << 2

		patAddr := 0x1000*table + uint16(nameTableByte)*16 + fineY
		lowByte := ppu.Read(patAddr)
		highByte := ppu.Read(patAddr + 8)

		// Render pixels from this tile
		startBit := int(fineX)
		for bit := startBit; bit < 8 && pixelX < 256; bit++ {
			shift := uint(7 - bit)
			p1 := (lowByte >> shift) & 1
			p2 := (highByte >> shift) & 1
			color := attribute | p2<<1 | p1

			if color&3 == 0 || (pixelX < 8 && ppu.flagShowLeftBackground == 0) {
				color = 0
			}

			ppu.backgroundColors[pixelX] = color
			ppu.front[scanline*256+pixelX] = ppu.readPalette(uint16(color)) % 64
			pixelX++
		}

		// After first tile, fine X is 0
		fineX = 0

		// Increment coarse X in v
		ppu.incrementX()
	}
}
