package main

import (
	_ "embed"

	"github.com/Senoue/tinygo-nes/nes"
)

// Place your .nes ROM file in roms/ and update the path below.
//
//go:embed roms/game.nes
var romData string

func main() {
	console, err := NewConsoleFromString(romData)
	if err != nil {
		panic(err.Error())
	}

	// Transfer every other frame in 24-line RGB565 batches.
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
