package main

import (
	_ "embed"

	"github.com/Senoue/tinygo-nes/nes"
)

//go:embed roms/DragonQuest.nes
var romData string

func main() {
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

	prg := make([]byte, prgSize)
	copy(prg, d[offset:offset+prgSize])
	offset += prgSize

	chr := make([]byte, chrSize)
	copy(chr, d[offset:offset+chrSize])

	cartridge := nes.NewCartridge(prg, chr, mapperNum, mirror, battery)
	ram := make([]byte, 2048)
	console := &nes.Console{
		Cartridge:   cartridge,
		Controller1: nes.NewController(),
		Controller2: nes.NewController(),
		RAM:         ram,
	}

	mapper, _ := nes.NewMapper(console)
	console.Mapper = mapper
	console.CPU = nes.NewCPUNoReset(console)
	console.APU = nes.NewAPU(console)
	console.PPU = nes.NewPPU(console)

	// Manual CPU reset (bypass interface)
	console.CPU.PC = uint16(prg[len(prg)-3])<<8 | uint16(prg[len(prg)-4])
	console.CPU.SP = 0xFD
	console.CPU.SetFlags(0x24)

	// Emulation loop - render every frame, 256px wide, 16-line batches
	const batchLines = 24
	const nesW = 256
	const offsetX = 32
	var batchBuf [nesW * 2 * batchLines]byte

	frameCount := byte(0)
	for {
		buttons := GetPressedButtons()
		console.SetButtons1(buttons)
		console.StepFrame()
		frameCount++

		OutputAudio(console.APU.Output())

		if frameCount&1 != 0 {
			continue
		}

		buf := console.Buffer()
		if buf == nil {
			continue
		}
		for yBatch := 0; yBatch < 240; yBatch += batchLines {
			for row := 0; row < batchLines; row++ {
				y := yBatch + row
				lineOff := row * nesW * 2
				for x := 0; x < nesW; x++ {
					rgb565 := nes.PaletteRGB565[buf[y*nesW+x]]
					batchBuf[lineOff] = byte(rgb565 >> 8)
					batchBuf[lineOff+1] = byte(rgb565)
					lineOff += 2
				}
			}
			display.DrawRGBBitmap8(int16(offsetX), int16(yBatch), batchBuf[:], nesW, batchLines)
		}
	}
}

