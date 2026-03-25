package main

import (
	"errors"

	"github.com/Senoue/tinygo-nes/nes"
)

const iNESFileMagic = 0x1a53454e

func NewConsoleFromString(data string) (*nes.Console, error) {
	if len(data) < 16 {
		return nil, errors.New("ROM too short")
	}

	magic := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	if magic != iNESFileMagic {
		return nil, errors.New("invalid .nes file")
	}

	numPRG := data[4]
	numCHR := data[5]
	control1 := data[6]
	control2 := data[7]

	mapper := (control1 >> 4) | (control2>>4)<<4
	mirror := (control1 & 1) | ((control1 >> 3) & 1 << 1)
	battery := (control1 >> 1) & 1

	offset := 16
	if control1&4 == 4 {
		offset += 512
	}

	prgSize := int(numPRG) * 16384
	chrSize := int(numCHR) * 8192

	if len(data) < offset+prgSize+chrSize {
		return nil, errors.New("ROM data truncated")
	}


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

	cartridge := nes.NewCartridge(prg, chr, mapper, mirror, battery)

	return nes.NewConsole(cartridge)
}
