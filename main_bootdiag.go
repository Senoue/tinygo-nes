//go:build wioterminal && bootdiag

package main

import (
	_ "embed"
	"image/color"
	"time"

	"github.com/Senoue/tinygo-nes/nes"
)

//go:embed roms/game.nes
var diagnosticROM string

// No USB output, audio init, or performance reporting in this diagnostic.
func main() {
	stage := func(c color.RGBA) {
		display.FillScreen(c)
		time.Sleep(700 * time.Millisecond)
	}
	stage(color.RGBA{0, 0, 255, 255}) // Entered main; display init returned.
	console, err := NewConsoleFromString(diagnosticROM)
	if err != nil {
		display.FillScreen(color.RGBA{255, 0, 0, 255})
		for {
			time.Sleep(time.Second)
		}
	}
	stage(color.RGBA{0, 255, 0, 255}) // ROM and console constructed.
	console.SetFrameOutput(true)
	stage(color.RGBA{255, 255, 0, 255}) // About to execute first frame.
	console.StepFrame()
	stage(color.RGBA{255, 0, 255, 255}) // First frame returned.
	line := make([]byte, 256*2)
	for {
		console.SetButtons1(GetPressedButtons())
		console.StepFrame()
		buf := console.Buffer()
		for y := 0; y < 240; y++ {
			for x := 0; x < 256; x++ {
				c := nes.PaletteRGB565[buf[y*256+x]&63]
				line[2*x], line[2*x+1] = byte(c>>8), byte(c)
			}
			display.DrawRGBBitmap8(32, int16(y), line, 256, 1)
		}
	}
}
