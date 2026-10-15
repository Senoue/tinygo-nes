package nes

import (
	"errors"
)

type Mapper interface {
	Read(address uint16) byte
	Write(address uint16, value byte)
	Step()
}

func NewMapper(console *Console) (Mapper, error) {
	cartridge := console.Cartridge
	if len(cartridge.PRG) == 0 || len(cartridge.PRG)%0x4000 != 0 {
		return nil, errors.New("invalid PRG size")
	}
	if len(cartridge.CHR) < 0x2000 || len(cartridge.CHR)%0x2000 != 0 {
		return nil, errors.New("invalid CHR size")
	}
	if cartridge.Mapper == 7 && len(cartridge.PRG)%0x8000 != 0 {
		return nil, errors.New("mapper 7 requires 32KB PRG banks")
	}
	if cartridge.Mapper == 40 && len(cartridge.PRG) < 0x10000 {
		return nil, errors.New("mapper 40 requires at least 64KB PRG")
	}
	switch cartridge.Mapper {
	case 0:
		return NewMapper2(cartridge), nil
	case 1:
		return NewMapper1(cartridge), nil
	case 2:
		return NewMapper2(cartridge), nil
	case 3:
		return NewMapper3(cartridge), nil
	case 4:
		return NewMapper4(console, cartridge), nil
	case 7:
		return NewMapper7(cartridge), nil
	case 40:
		return NewMapper40(console, cartridge), nil
	case 225:
		return NewMapper225(cartridge), nil
	}
	return nil, errors.New("unsupported mapper")
}
