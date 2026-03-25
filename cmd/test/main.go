package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/Senoue/tinygo-nes/nes"
)

func main() {
	// Load ROM
	data, err := os.ReadFile("roms/DragonQuest.nes")
	if err != nil {
		fmt.Println("Error reading ROM:", err)
		os.Exit(1)
	}

	// Parse ROM header
	if len(data) < 16 {
		fmt.Println("ROM too short")
		os.Exit(1)
	}
	magic := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	fmt.Printf("Magic: 0x%08X\n", magic)
	fmt.Printf("PRG banks: %d (%dKB)\n", data[4], int(data[4])*16)
	fmt.Printf("CHR banks: %d (%dKB)\n", data[5], int(data[5])*8)
	fmt.Printf("Control1: 0x%02X\n", data[6])
	fmt.Printf("Control2: 0x%02X\n", data[7])

	numPRG := data[4]
	numCHR := data[5]
	control1 := data[6]
	control2 := data[7]

	mapperNum := (control1 >> 4) | (control2>>4)<<4
	mirror := (control1 & 1) | ((control1>>3)&1)<<1
	battery := (control1 >> 1) & 1
	fmt.Printf("Mapper: %d\n", mapperNum)
	fmt.Printf("Mirror: %d\n", mirror)
	fmt.Printf("Battery: %d\n", battery)

	offset := 16
	if control1&4 == 4 {
		offset += 512
	}

	prgSize := int(numPRG) * 16384
	chrSize := int(numCHR) * 8192

	prg := make([]byte, prgSize)
	copy(prg, data[offset:offset+prgSize])
	offset += prgSize

	var chr []byte
	if numCHR == 0 {
		chr = make([]byte, 8192)
	} else {
		chr = make([]byte, chrSize)
		copy(chr, data[offset:offset+chrSize])
	}

	cartridge := nes.NewCartridge(prg, chr, mapperNum, mirror, battery)
	console, err := nes.NewConsole(cartridge)
	if err != nil {
		fmt.Println("Error creating console:", err)
		os.Exit(1)
	}

	// Run 300 frames
	for i := 0; i < 300; i++ {
		cycles := console.StepFrame()
		if i < 10 || i%50 == 0 {
			fmt.Printf("Frame %d: %d cycles, PPU.Frame=%d\n", i, cycles, console.PPU.Frame)
		}
	}

	// Save frame as PNG
	buf := console.Buffer()
	img := image.NewRGBA(image.Rect(0, 0, 256, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 256; x++ {
			idx := buf[y*256+x]
			c := nes.Palette[idx]
			img.SetRGBA(x, y, c)
		}
	}

	f, _ := os.Create("frame300.png")
	png.Encode(f, img)
	f.Close()
	fmt.Println("Saved frame300.png")
}
