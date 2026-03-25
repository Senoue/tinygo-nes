package main

import (
	_ "embed"
	"image/color"

	"github.com/Senoue/tinygo-nes/nes"
)

//go:embed roms/DragonQuest.nes
var romData string

func main() {
	// green = start
	display.FillScreen(color.RGBA{0x00, 0xFF, 0x00, 0xFF})

	// Parse ROM header manually
	d := romData
	numPRG := d[4]
	numCHR := d[5]
	ctrl1 := d[6]
	ctrl2 := d[7]
	mapperNum := (ctrl1 >> 4) | (ctrl2>>4)<<4
	mirror := (ctrl1 & 1) | ((ctrl1>>3)&1)<<1
	battery := (ctrl1 >> 1) & 1
	offset := 16

	prgSize := int(numPRG) * 16384
	chrSize := int(numCHR) * 8192

	// dark green = PRG alloc
	display.FillScreen(color.RGBA{0x00, 0x80, 0x00, 0xFF})
	prg := make([]byte, prgSize)
	copy(prg, d[offset:offset+prgSize])
	offset += prgSize

	// dark blue = CHR alloc
	display.FillScreen(color.RGBA{0x00, 0x00, 0x80, 0xFF})
	chr := make([]byte, chrSize)
	copy(chr, d[offset:offset+chrSize])

	// purple = cartridge
	display.FillScreen(color.RGBA{0x80, 0x00, 0x80, 0xFF})
	cartridge := nes.NewCartridge(prg, chr, mapperNum, mirror, battery)

	// orange = RAM + controllers
	display.FillScreen(color.RGBA{0xFF, 0x80, 0x00, 0xFF})
	ram := make([]byte, 2048)
	controller1 := nes.NewController()
	controller2 := nes.NewController()

	// pink = building console struct
	display.FillScreen(color.RGBA{0xFF, 0x00, 0xFF, 0xFF})
	console := &nes.Console{
		Cartridge:   cartridge,
		Controller1: controller1,
		Controller2: controller2,
		RAM:         ram,
	}

	// teal = mapper
	display.FillScreen(color.RGBA{0x00, 0x80, 0x80, 0xFF})
	mapper, err := nes.NewMapper(console)
	if err != nil {
		display.FillScreen(color.RGBA{0xFF, 0x00, 0x00, 0xFF})
		for {
		}
	}
	console.Mapper = mapper

	// dark red = CPU struct (no reset yet)
	display.FillScreen(color.RGBA{0x80, 0x00, 0x00, 0xFF})
	console.CPU = nes.NewCPUNoReset(console)

	// dark cyan = APU
	display.FillScreen(color.RGBA{0x00, 0x80, 0x80, 0xFF})
	console.APU = nes.NewAPU(console)

	// white = PPU (60KB)
	display.FillScreen(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	console.PPU = nes.NewPPU(console)

	// bright red = CPU Reset (manual, bypass interface call)
	display.FillScreen(color.RGBA{0xFF, 0x40, 0x40, 0xFF})
	// Read reset vector directly from PRG ROM
	// 0xFFFC in CPU space = PRG[prgBank2*0x4000 + 0x3FFC]
	// prgBank2 = numPRG-1 = 1, so PRG[0x7FFC] and PRG[0x7FFD]
	resetLo := uint16(prg[len(prg)-4])
	resetHi := uint16(prg[len(prg)-3])
	console.CPU.PC = resetHi<<8 | resetLo
	console.CPU.SP = 0xFD
	console.CPU.SetFlags(0x24)

	// Emulation loop
	var batchBuf [320 * 2 * 8]byte
	for {
		buttons := GetPressedButtons()
		console.SetButtons1(buttons)
		console.StepFrame()

		buf := console.Buffer()
		if buf == nil {
			continue
		}
		for yBatch := 0; yBatch < 240; yBatch += 8 {
			for y := 0; y < 8; y++ {
				renderLine(buf, yBatch+y, batchBuf[:], y)
			}
			display.DrawRGBBitmap8(0, int16(yBatch), batchBuf[:], 320, 8)
		}
	}
}

func renderLine(buf *[256 * 240]byte, y int, batch []byte, batchRow int) {
	const lineStride = 320 * 2
	lineOffset := batchRow * lineStride
	// Zero the line region within the batch
	for i := lineOffset; i < lineOffset+lineStride; i++ {
		batch[i] = 0
	}
	const offsetX = 32
	for x := 0; x < 256; x++ {
		idx := buf[y*256+x]
		rgb565 := nes.PaletteRGB565[idx]
		dstIdx := lineOffset + (x+offsetX)*2
		batch[dstIdx] = byte(rgb565 >> 8)
		batch[dstIdx+1] = byte(rgb565)
	}
}

