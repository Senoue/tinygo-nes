//go:build wioterminal
// +build wioterminal

package main

import (
	"image/color"
	"machine"

	"github.com/Senoue/tinygo-nes/nes"
	"tinygo.org/x/drivers/ili9341"
)

var (
	display *ili9341.Device
	dac     machine.DAC
)

func init() {
	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		SDO:       machine.LCD_SDO_PIN,
		SDI:       machine.LCD_SDI_PIN,
		Frequency: 80000000,
	})

	display = ili9341.NewSPI(
		machine.SPI3,
		machine.LCD_DC,
		machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)
	display.Configure(ili9341.Config{
		Rotation: ili9341.Rotation270,
	})

	machine.LCD_BACKLIGHT.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LCD_BACKLIGHT.High()

	display.FillScreen(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})

	// DAC for audio output
	dac = machine.DAC{Channel: 0}
	dac.Configure(machine.DACConfig{})

	machine.BUTTON_1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.BUTTON_2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.BUTTON_3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	machine.WIO_5S_UP.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.WIO_5S_LEFT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.WIO_5S_RIGHT.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.WIO_5S_DOWN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	machine.WIO_5S_PRESS.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

// GetPressedButtons returns NES controller button states.
func GetPressedButtons() [8]bool {
	var buttons [8]bool
	buttons[nes.ButtonSelect] = !machine.BUTTON_1.Get()
	buttons[nes.ButtonB] = !machine.BUTTON_2.Get()
	buttons[nes.ButtonStart] = !machine.BUTTON_3.Get()
	buttons[nes.ButtonA] = !machine.WIO_5S_PRESS.Get()
	buttons[nes.ButtonUp] = !machine.WIO_5S_UP.Get()
	buttons[nes.ButtonDown] = !machine.WIO_5S_DOWN.Get()
	buttons[nes.ButtonLeft] = !machine.WIO_5S_LEFT.Get()
	buttons[nes.ButtonRight] = !machine.WIO_5S_RIGHT.Get()
	return buttons
}

// OutputAudio sends an APU sample to the DAC.
// Input: float32 0.0~1.0 → DAC 0~65535
func OutputAudio(sample float32) {
	if sample < 0 {
		sample = 0
	}
	if sample > 1 {
		sample = 1
	}
	dac.Set(uint16(sample * 65535))
}
