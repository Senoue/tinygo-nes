//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/Senoue/tinygo-nes/nes"
)

// The browser owns timing and input; the same NES core renders one frame per call.
func main() {
	var console *nes.Console
	pixels := make([]byte, 256*240*4)
	load := js.FuncOf(func(this js.Value, args []js.Value) (result interface{}) {
		defer func() {
			if err := recover(); err != nil {
				result = fmt.Sprint("ROM load failed: ", err)
			}
		}()
		if len(args) != 1 {
			return "ROM data required"
		}
		data := make([]byte, args[0].Get("byteLength").Int())
		js.CopyBytesToGo(data, args[0])
		next, err := NewConsoleFromString(string(data))
		if err != nil {
			return err.Error()
		}
		console = next
		return ""
	})
	frame := js.FuncOf(func(this js.Value, args []js.Value) (result interface{}) {
		defer func() {
			if err := recover(); err != nil {
				result = fmt.Sprint("Emulation stopped: ", err)
			}
		}()
		if console == nil {
			return "Load a ROM first"
		}
		var buttons [8]bool
		mask := args[0].Int()
		for i := range buttons {
			buttons[i] = mask&(1<<i) != 0
		}
		console.SetButtons1(buttons)
		console.SetFrameOutput(true)
		console.StepFrame()
		for i, index := range console.Buffer() {
			c := nes.Palette[index&63]
			pixels[4*i], pixels[4*i+1], pixels[4*i+2], pixels[4*i+3] = c.R, c.G, c.B, 255
		}
		js.CopyBytesToJS(args[1], pixels)
		return ""
	})
	js.Global().Set("nesLoadROM", load)
	js.Global().Set("nesFrame", frame)
	select {}
}
