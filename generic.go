//go:build !baremetal && !js
// +build !baremetal,!js

package main

import (
	"image/color"

	"github.com/Senoue/tinygo-nes/nes"
	"github.com/sago35/tinydisplay/examples/initdisplay"
)

var (
	display *initdisplay.TinyDisplay
)

func init() {
	display = initdisplay.InitDisplay()
	display.FillScreen(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
}

// GetPressedButtons returns NES controller button states from keyboard input.
func GetPressedButtons() [8]bool {
	var buttons [8]bool
	key := display.GetPressedKey()
	switch key {
	case 0x109: // Up
		buttons[nes.ButtonUp] = true
	case 0x108: // Down
		buttons[nes.ButtonDown] = true
	case 0x107: // Left
		buttons[nes.ButtonLeft] = true
	case 0x106: // Right
		buttons[nes.ButtonRight] = true
	case 0x020: // Space → A
		buttons[nes.ButtonA] = true
	case 0x103: // Backspace → B
		buttons[nes.ButtonB] = true
	case 0x101: // Return → Start
		buttons[nes.ButtonStart] = true
	}
	return buttons
}
