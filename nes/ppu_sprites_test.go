package nes

import (
	"math/rand"
	"reflect"
	"testing"
)

func (ppu *PPU) renderScanlineSpritesReference(scanline int) {
	if ppu.flagShowSprites == 0 {
		return
	}

	// Process sprites in reverse order so lower-index sprites have priority
	for i := ppu.spriteCount - 1; i >= 0; i-- {
		spriteX := int(ppu.spritePositions[i])
		pattern := ppu.spritePatterns[i]
		priority := ppu.spritePriorities[i]
		spriteIdx := ppu.spriteIndexes[i]

		for j := 0; j < 8; j++ {
			px := spriteX + j
			if px < 0 || px >= 256 {
				continue
			}

			// Extract sprite pixel color from pattern (MSB first)
			offset := uint(7 - j)
			spriteColor := byte((pattern >> (offset * 4)) & 0x0F)

			if spriteColor%4 == 0 {
				// Transparent sprite pixel
				continue
			}

			if px < 8 && ppu.flagShowLeftSprites == 0 {
				continue
			}

			// Transparency is determined by pattern bits, independent of palette
			// colors and any higher-index sprites already drawn at this pixel.
			background := ppu.backgroundColors[px]
			bgColor := background & 3

			// Sprite 0 hit detection
			if spriteIdx == 0 && bgColor != 0 && px < 255 {
				ppu.flagSpriteZeroHit = 1
			}

			// Priority: 0 = sprite in front of bg, 1 = sprite behind bg
			if priority == 0 || bgColor == 0 {
				ppu.front[scanline*256+px] = ppu.readPalette(uint16(spriteColor)|0x10) % 64
			} else {
				// A lower-index sprite behind the background also hides higher-index sprites.
				ppu.front[scanline*256+px] = ppu.readPalette(uint16(background)) % 64
			}
		}
	}
}

func TestSpriteRendererMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(71))
	slow, fast := outputTestConsole(t, 3), outputTestConsole(t, 3)
	for trial := 0; trial < 2000; trial++ {
		p := slow.PPU
		p.spriteCount = rng.Intn(9)
		p.flagShowSprites = byte(rng.Intn(2))
		p.flagShowLeftSprites = byte(rng.Intn(2))
		p.flagSpriteZeroHit = byte(rng.Intn(2))
		for i := range p.paletteData {
			p.paletteData[i] = byte(rng.Intn(256))
		}
		for i := range p.backgroundColors {
			p.backgroundColors[i] = byte(rng.Intn(16))
		}
		for i := 0; i < p.spriteCount; i++ {
			p.spritePositions[i] = byte(rng.Intn(256))
			p.spritePatterns[i] = rng.Uint32()
			p.spritePriorities[i] = byte(rng.Intn(2))
			p.spriteIndexes[i] = byte(i)
		}
		frame := fast.PPU.front
		*fast.PPU = *p
		fast.PPU.console = fast
		fast.PPU.front = frame
		*frame = *p.front
		line := rng.Intn(240)
		p.renderScanlineSpritesReference(line)
		fast.PPU.renderScanlineSprites(line)
		if !reflect.DeepEqual(p.front, fast.PPU.front) || p.flagSpriteZeroHit != fast.PPU.flagSpriteZeroHit {
			t.Fatalf("trial %d: pixels or sprite zero hit differ", trial)
		}
	}
}
