package nes


type Mapper3 struct {
	*Cartridge
	chrBank  int
	prgBank1 int
	prgBank2 int
}

func NewMapper3(cartridge *Cartridge) Mapper {
	prgBanks := len(cartridge.PRG) / 0x4000
	return &Mapper3{cartridge, 0, 0, prgBanks - 1}
}

func (m *Mapper3) Step() {
}

// readCPU is the direct CPU path for addresses $6000-$FFFF. CHR bank
// switching still goes through Write and does not change the PRG mapping.
func (m *Mapper3) readCPU(address uint16) byte {
	if address >= 0xc000 {
		return m.PRG[m.prgBank2*0x4000+int(address&0x3fff)]
	}
	if address >= 0x8000 {
		return m.PRG[m.prgBank1*0x4000+int(address&0x3fff)]
	}
	return m.SRAM[address-0x6000]
}

func (m *Mapper3) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		index := m.chrBank*0x2000 + int(address)
		return m.CHR[index]
	case address >= 0xC000:
		index := m.prgBank2*0x4000 + int(address-0xC000)
		return m.PRG[index]
	case address >= 0x8000:
		index := m.prgBank1*0x4000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		return 0
	}
	return 0
}

func (m *Mapper3) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		index := m.chrBank*0x2000 + int(address)
		m.CHR[index] = value
	case address >= 0x8000:
		m.chrBank = int(value & 3)
	case address >= 0x6000:
		index := int(address) - 0x6000
		m.SRAM[index] = value
	default:
		return
	}
}
