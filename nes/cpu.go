package nes


const CPUFrequency = 1789773

// interrupt types
const (
	_ = iota
	interruptNone
	interruptNMI
	interruptIRQ
)

// addressing modes
const (
	_ = iota
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeAccumulator
	modeImmediate
	modeImplied
	modeIndexedIndirect
	modeIndirect
	modeIndirectIndexed
	modeRelative
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
)

// instructionModes indicates the addressing mode for each instruction
var instructionModes = [256]byte{
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	1, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 8, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
}

// instructionSizes indicates the size of each instruction in bytes
var instructionSizes = [256]byte{
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	3, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 0, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 0, 3, 0, 0,
	2, 2, 2, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
}

// instructionCycles indicates the number of cycles used by each instruction,
// not including conditional cycles
var instructionCycles = [256]byte{
	7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 6, 2, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 5, 2, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
}

// instructionPageCycles indicates the number of cycles used by each
// instruction when a page is crossed
var instructionPageCycles = [256]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
}


type CPU struct {
	console   *Console // direct console pointer (no interface)
	Cycles    uint64   // number of cycles
	PC        uint16 // program counter
	SP        byte   // stack pointer
	A         byte   // accumulator
	X         byte   // x register
	Y         byte   // y register
	C         byte   // carry flag
	Z         byte   // zero flag
	I         byte   // interrupt disable flag
	D         byte   // decimal mode flag
	B         byte   // break command flag
	U         byte   // unused flag
	V         byte   // overflow flag
	N         byte   // negative flag
	interrupt   byte // interrupt type to perform
	stall       int  // number of cycles to stall
	stepInfoBuf stepInfo // reusable buffer to avoid heap allocation
}

func NewCPU(console *Console) *CPU {
	cpu := CPU{console: console, interrupt: interruptNone}
	cpu.Reset()
	return &cpu
}

func NewCPUNoReset(console *Console) *CPU {
	cpu := CPU{console: console, interrupt: interruptNone}
	return &cpu
}

func (cpu *CPU) Read(address uint16) byte {
	// RAM and cartridge reads dominate instruction execution. Keep device
	// register decoding out of that path, including instruction fetches.
	if address < 0x2000 {
		return cpu.console.RAM[address&0x07ff]
	}
	if address >= 0x6000 {
		// Common discrete mappers can decode CPU memory without a virtual
		// call or a CHR address check. Keep other mappers on their own path.
		switch m := cpu.console.Mapper.(type) {
		case *Mapper2: // Mapper 0 also uses Mapper2.
			return m.readCPU(address)
		case *Mapper3:
			return m.readCPU(address)
		}
		return cpu.console.Mapper.Read(address)
	}
	return cpu.readIO(address)
}

func (cpu *CPU) readIO(address uint16) byte {
	switch {
	case address < 0x4000:
		return cpu.console.PPU.readRegister(0x2000 + address%8)
	case address == 0x4014:
		return cpu.console.PPU.readRegister(address)
	case address == 0x4015:
		return cpu.console.APU.readRegister(address)
	case address == 0x4016:
		return cpu.console.Controller1.Read()
	case address == 0x4017:
		return cpu.console.Controller2.Read()
	default:
		return 0
	}
	return 0
}

func (cpu *CPU) Write(address uint16, value byte) {
	if address < 0x2000 {
		cpu.console.RAM[address&0x07ff] = value
		return
	}
	if address >= 0x6000 {
		cpu.console.Mapper.Write(address, value)
		return
	}
	cpu.writeIO(address, value)
}

func (cpu *CPU) writeIO(address uint16, value byte) {
	switch {
	case address < 0x4000:
		cpu.console.PPU.writeRegister(0x2000+address%8, value)
	case address < 0x4014:
		cpu.console.APU.writeRegister(address, value)
	case address == 0x4014:
		cpu.console.PPU.writeRegister(address, value)
	case address == 0x4015:
		cpu.console.APU.writeRegister(address, value)
	case address == 0x4016:
		cpu.console.Controller1.Write(value)
		cpu.console.Controller2.Write(value)
	case address == 0x4017:
		cpu.console.APU.writeRegister(address, value)
	default:
		return
	}
}


// Reset resets the CPU to its initial powerup state
func (cpu *CPU) Reset() {
	cpu.PC = cpu.Read16(0xFFFC)
	cpu.SP = 0xFD
	cpu.SetFlags(0x24)
}


// pagesDiffer returns true if the two addresses reference different pages
func pagesDiffer(a, b uint16) bool {
	return a&0xFF00 != b&0xFF00
}

// addBranchCycles adds a cycle for taking a branch and adds another cycle
// if the branch jumps to a new page
func (cpu *CPU) addBranchCycles(info *stepInfo) {
	cpu.Cycles++
	if pagesDiffer(info.pc, info.address) {
		cpu.Cycles++
	}
}

func (cpu *CPU) compare(a, b byte) {
	cpu.setZN(a - b)
	if a >= b {
		cpu.C = 1
	} else {
		cpu.C = 0
	}
}

// Read16 reads two bytes using Read to return a double-word value
func (cpu *CPU) Read16(address uint16) uint16 {
	lo := uint16(cpu.Read(address))
	hi := uint16(cpu.Read(address + 1))
	return hi<<8 | lo
}

// read16bug emulates a 6502 bug that caused the low byte to wrap without
// incrementing the high byte
func (cpu *CPU) read16bug(address uint16) uint16 {
	a := address
	b := (a & 0xFF00) | uint16(byte(a)+1)
	lo := cpu.Read(a)
	hi := cpu.Read(b)
	return uint16(hi)<<8 | uint16(lo)
}

// push pushes a byte onto the stack
func (cpu *CPU) push(value byte) {
	cpu.console.RAM[0x100|uint16(cpu.SP)] = value
	cpu.SP--
}

// pull pops a byte from the stack
func (cpu *CPU) pull() byte {
	cpu.SP++
	return cpu.console.RAM[0x100|uint16(cpu.SP)]
}

// Zero-page pointers always read internal RAM and wrap from $FF to $00.
func (cpu *CPU) read16ZeroPage(address byte) uint16 {
	ram := cpu.console.RAM
	return uint16(ram[address]) | uint16(ram[address+1])<<8
}

// push16 pushes two bytes onto the stack
func (cpu *CPU) push16(value uint16) {
	hi := byte(value >> 8)
	lo := byte(value & 0xFF)
	cpu.push(hi)
	cpu.push(lo)
}

// pull16 pops two bytes from the stack
func (cpu *CPU) pull16() uint16 {
	lo := uint16(cpu.pull())
	hi := uint16(cpu.pull())
	return hi<<8 | lo
}

// Flags returns the processor status flags
func (cpu *CPU) Flags() byte {
	var flags byte
	flags |= cpu.C << 0
	flags |= cpu.Z << 1
	flags |= cpu.I << 2
	flags |= cpu.D << 3
	flags |= cpu.B << 4
	flags |= cpu.U << 5
	flags |= cpu.V << 6
	flags |= cpu.N << 7
	return flags
}

// SetFlags sets the processor status flags
func (cpu *CPU) SetFlags(flags byte) {
	cpu.C = (flags >> 0) & 1
	cpu.Z = (flags >> 1) & 1
	cpu.I = (flags >> 2) & 1
	cpu.D = (flags >> 3) & 1
	cpu.B = (flags >> 4) & 1
	cpu.U = (flags >> 5) & 1
	cpu.V = (flags >> 6) & 1
	cpu.N = (flags >> 7) & 1
}

// setZ sets the zero flag if the argument is zero
func (cpu *CPU) setZ(value byte) {
	if value == 0 {
		cpu.Z = 1
	} else {
		cpu.Z = 0
	}
}

// setN sets the negative flag if the argument is negative (high bit is set)
func (cpu *CPU) setN(value byte) {
	if value&0x80 != 0 {
		cpu.N = 1
	} else {
		cpu.N = 0
	}
}

// setZN sets the zero flag and the negative flag
func (cpu *CPU) setZN(value byte) {
	cpu.setZ(value)
	cpu.setN(value)
}

// triggerNMI causes a non-maskable interrupt to occur on the next cycle
func (cpu *CPU) triggerNMI() {
	cpu.interrupt = interruptNMI
}

// triggerIRQ causes an IRQ interrupt to occur on the next cycle
func (cpu *CPU) triggerIRQ() {
	if cpu.I == 0 {
		cpu.interrupt = interruptIRQ
	}
}

// stepInfo contains information that the instruction functions use
type stepInfo struct {
	address uint16
	pc      uint16
	mode    byte
}

// Step executes a single CPU instruction
func (cpu *CPU) Step() int {
	if cpu.stall > 0 {
		cpu.stall--
		return 1
	}

	cycles := cpu.Cycles

	// The APU frame IRQ remains asserted until $4015 is read or $4017
	// inhibits it. CPU interrupt masking must not discard that request.
	if cpu.interrupt == interruptNone && cpu.I == 0 && cpu.console.APU != nil && cpu.console.APU.frameIRQPending {
		cpu.interrupt = interruptIRQ
	}
	switch cpu.interrupt {
	case interruptNMI:
		cpu.nmi()
	case interruptIRQ:
		cpu.irq()
	}
	cpu.interrupt = interruptNone

	opcode := cpu.Read(cpu.PC)
	mode := instructionModes[opcode]

	var address uint16
	var pageCrossed bool
	switch mode {
	case modeAbsolute:
		address = cpu.Read16(cpu.PC + 1)
	case modeAbsoluteX:
		address = cpu.Read16(cpu.PC+1) + uint16(cpu.X)
		pageCrossed = pagesDiffer(address-uint16(cpu.X), address)
	case modeAbsoluteY:
		address = cpu.Read16(cpu.PC+1) + uint16(cpu.Y)
		pageCrossed = pagesDiffer(address-uint16(cpu.Y), address)
	case modeAccumulator:
		address = 0
	case modeImmediate:
		address = cpu.PC + 1
	case modeImplied:
		address = 0
	case modeIndexedIndirect:
		address = cpu.read16ZeroPage(cpu.Read(cpu.PC+1) + cpu.X)
	case modeIndirect:
		address = cpu.read16bug(cpu.Read16(cpu.PC + 1))
	case modeIndirectIndexed:
		address = cpu.read16ZeroPage(cpu.Read(cpu.PC+1)) + uint16(cpu.Y)
		pageCrossed = pagesDiffer(address-uint16(cpu.Y), address)
	case modeRelative:
		offset := uint16(cpu.Read(cpu.PC + 1))
		if offset < 0x80 {
			address = cpu.PC + 2 + offset
		} else {
			address = cpu.PC + 2 + offset - 0x100
		}
	case modeZeroPage:
		address = uint16(cpu.Read(cpu.PC + 1))
	case modeZeroPageX:
		address = uint16(cpu.Read(cpu.PC+1)+cpu.X) & 0xff
	case modeZeroPageY:
		address = uint16(cpu.Read(cpu.PC+1)+cpu.Y) & 0xff
	}

	cpu.PC += uint16(instructionSizes[opcode])
	cpu.Cycles += uint64(instructionCycles[opcode])
	if pageCrossed {
		cpu.Cycles += uint64(instructionPageCycles[opcode])
	}
	cpu.stepInfoBuf.address = address
	cpu.stepInfoBuf.pc = cpu.PC
	cpu.stepInfoBuf.mode = mode
	switch opcode {
	case 0x00:
		cpu.brk(&cpu.stepInfoBuf)
	case 0x01:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x02:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x03:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x04:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x05:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x06:
		cpu.asl(&cpu.stepInfoBuf)
	case 0x07:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x08:
		cpu.php(&cpu.stepInfoBuf)
	case 0x09:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x0A:
		cpu.asl(&cpu.stepInfoBuf)
	case 0x0B:
		cpu.anc(&cpu.stepInfoBuf)
	case 0x0C:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x0D:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x0E:
		cpu.asl(&cpu.stepInfoBuf)
	case 0x0F:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x10:
		cpu.bpl(&cpu.stepInfoBuf)
	case 0x11:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x12:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x13:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x14:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x15:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x16:
		cpu.asl(&cpu.stepInfoBuf)
	case 0x17:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x18:
		cpu.clc(&cpu.stepInfoBuf)
	case 0x19:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x1A:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x1B:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x1C:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x1D:
		cpu.ora(&cpu.stepInfoBuf)
	case 0x1E:
		cpu.asl(&cpu.stepInfoBuf)
	case 0x1F:
		cpu.slo(&cpu.stepInfoBuf)
	case 0x20:
		cpu.jsr(&cpu.stepInfoBuf)
	case 0x21:
		cpu.and(&cpu.stepInfoBuf)
	case 0x22:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x23:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x24:
		cpu.bit(&cpu.stepInfoBuf)
	case 0x25:
		cpu.and(&cpu.stepInfoBuf)
	case 0x26:
		cpu.rol(&cpu.stepInfoBuf)
	case 0x27:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x28:
		cpu.plp(&cpu.stepInfoBuf)
	case 0x29:
		cpu.and(&cpu.stepInfoBuf)
	case 0x2A:
		cpu.rol(&cpu.stepInfoBuf)
	case 0x2B:
		cpu.anc(&cpu.stepInfoBuf)
	case 0x2C:
		cpu.bit(&cpu.stepInfoBuf)
	case 0x2D:
		cpu.and(&cpu.stepInfoBuf)
	case 0x2E:
		cpu.rol(&cpu.stepInfoBuf)
	case 0x2F:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x30:
		cpu.bmi(&cpu.stepInfoBuf)
	case 0x31:
		cpu.and(&cpu.stepInfoBuf)
	case 0x32:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x33:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x34:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x35:
		cpu.and(&cpu.stepInfoBuf)
	case 0x36:
		cpu.rol(&cpu.stepInfoBuf)
	case 0x37:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x38:
		cpu.sec(&cpu.stepInfoBuf)
	case 0x39:
		cpu.and(&cpu.stepInfoBuf)
	case 0x3A:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x3B:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x3C:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x3D:
		cpu.and(&cpu.stepInfoBuf)
	case 0x3E:
		cpu.rol(&cpu.stepInfoBuf)
	case 0x3F:
		cpu.rla(&cpu.stepInfoBuf)
	case 0x40:
		cpu.rti(&cpu.stepInfoBuf)
	case 0x41:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x42:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x43:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x44:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x45:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x46:
		cpu.lsr(&cpu.stepInfoBuf)
	case 0x47:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x48:
		cpu.pha(&cpu.stepInfoBuf)
	case 0x49:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x4A:
		cpu.lsr(&cpu.stepInfoBuf)
	case 0x4B:
		cpu.alr(&cpu.stepInfoBuf)
	case 0x4C:
		cpu.jmp(&cpu.stepInfoBuf)
	case 0x4D:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x4E:
		cpu.lsr(&cpu.stepInfoBuf)
	case 0x4F:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x50:
		cpu.bvc(&cpu.stepInfoBuf)
	case 0x51:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x52:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x53:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x54:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x55:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x56:
		cpu.lsr(&cpu.stepInfoBuf)
	case 0x57:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x58:
		cpu.cli(&cpu.stepInfoBuf)
	case 0x59:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x5A:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x5B:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x5C:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x5D:
		cpu.eor(&cpu.stepInfoBuf)
	case 0x5E:
		cpu.lsr(&cpu.stepInfoBuf)
	case 0x5F:
		cpu.sre(&cpu.stepInfoBuf)
	case 0x60:
		cpu.rts(&cpu.stepInfoBuf)
	case 0x61:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x62:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x63:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x64:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x65:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x66:
		cpu.ror(&cpu.stepInfoBuf)
	case 0x67:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x68:
		cpu.pla(&cpu.stepInfoBuf)
	case 0x69:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x6A:
		cpu.ror(&cpu.stepInfoBuf)
	case 0x6B:
		cpu.arr(&cpu.stepInfoBuf)
	case 0x6C:
		cpu.jmp(&cpu.stepInfoBuf)
	case 0x6D:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x6E:
		cpu.ror(&cpu.stepInfoBuf)
	case 0x6F:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x70:
		cpu.bvs(&cpu.stepInfoBuf)
	case 0x71:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x72:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x73:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x74:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x75:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x76:
		cpu.ror(&cpu.stepInfoBuf)
	case 0x77:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x78:
		cpu.sei(&cpu.stepInfoBuf)
	case 0x79:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x7A:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x7B:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x7C:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x7D:
		cpu.adc(&cpu.stepInfoBuf)
	case 0x7E:
		cpu.ror(&cpu.stepInfoBuf)
	case 0x7F:
		cpu.rra(&cpu.stepInfoBuf)
	case 0x80:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x81:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x82:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x83:
		cpu.sax(&cpu.stepInfoBuf)
	case 0x84:
		cpu.sty(&cpu.stepInfoBuf)
	case 0x85:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x86:
		cpu.stx(&cpu.stepInfoBuf)
	case 0x87:
		cpu.sax(&cpu.stepInfoBuf)
	case 0x88:
		cpu.dey(&cpu.stepInfoBuf)
	case 0x89:
		cpu.nop(&cpu.stepInfoBuf)
	case 0x8A:
		cpu.txa(&cpu.stepInfoBuf)
	case 0x8B:
		cpu.xaa(&cpu.stepInfoBuf)
	case 0x8C:
		cpu.sty(&cpu.stepInfoBuf)
	case 0x8D:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x8E:
		cpu.stx(&cpu.stepInfoBuf)
	case 0x8F:
		cpu.sax(&cpu.stepInfoBuf)
	case 0x90:
		cpu.bcc(&cpu.stepInfoBuf)
	case 0x91:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x92:
		cpu.kil(&cpu.stepInfoBuf)
	case 0x93:
		cpu.ahx(&cpu.stepInfoBuf)
	case 0x94:
		cpu.sty(&cpu.stepInfoBuf)
	case 0x95:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x96:
		cpu.stx(&cpu.stepInfoBuf)
	case 0x97:
		cpu.sax(&cpu.stepInfoBuf)
	case 0x98:
		cpu.tya(&cpu.stepInfoBuf)
	case 0x99:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x9A:
		cpu.txs(&cpu.stepInfoBuf)
	case 0x9B:
		cpu.tas(&cpu.stepInfoBuf)
	case 0x9C:
		cpu.shy(&cpu.stepInfoBuf)
	case 0x9D:
		cpu.sta(&cpu.stepInfoBuf)
	case 0x9E:
		cpu.shx(&cpu.stepInfoBuf)
	case 0x9F:
		cpu.ahx(&cpu.stepInfoBuf)
	case 0xA0:
		cpu.ldy(&cpu.stepInfoBuf)
	case 0xA1:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xA2:
		cpu.ldx(&cpu.stepInfoBuf)
	case 0xA3:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xA4:
		cpu.ldy(&cpu.stepInfoBuf)
	case 0xA5:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xA6:
		cpu.ldx(&cpu.stepInfoBuf)
	case 0xA7:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xA8:
		cpu.tay(&cpu.stepInfoBuf)
	case 0xA9:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xAA:
		cpu.tax(&cpu.stepInfoBuf)
	case 0xAB:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xAC:
		cpu.ldy(&cpu.stepInfoBuf)
	case 0xAD:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xAE:
		cpu.ldx(&cpu.stepInfoBuf)
	case 0xAF:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xB0:
		cpu.bcs(&cpu.stepInfoBuf)
	case 0xB1:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xB2:
		cpu.kil(&cpu.stepInfoBuf)
	case 0xB3:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xB4:
		cpu.ldy(&cpu.stepInfoBuf)
	case 0xB5:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xB6:
		cpu.ldx(&cpu.stepInfoBuf)
	case 0xB7:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xB8:
		cpu.clv(&cpu.stepInfoBuf)
	case 0xB9:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xBA:
		cpu.tsx(&cpu.stepInfoBuf)
	case 0xBB:
		cpu.las(&cpu.stepInfoBuf)
	case 0xBC:
		cpu.ldy(&cpu.stepInfoBuf)
	case 0xBD:
		cpu.lda(&cpu.stepInfoBuf)
	case 0xBE:
		cpu.ldx(&cpu.stepInfoBuf)
	case 0xBF:
		cpu.lax(&cpu.stepInfoBuf)
	case 0xC0:
		cpu.cpy(&cpu.stepInfoBuf)
	case 0xC1:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xC2:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xC3:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xC4:
		cpu.cpy(&cpu.stepInfoBuf)
	case 0xC5:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xC6:
		cpu.dec(&cpu.stepInfoBuf)
	case 0xC7:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xC8:
		cpu.iny(&cpu.stepInfoBuf)
	case 0xC9:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xCA:
		cpu.dex(&cpu.stepInfoBuf)
	case 0xCB:
		cpu.axs(&cpu.stepInfoBuf)
	case 0xCC:
		cpu.cpy(&cpu.stepInfoBuf)
	case 0xCD:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xCE:
		cpu.dec(&cpu.stepInfoBuf)
	case 0xCF:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xD0:
		cpu.bne(&cpu.stepInfoBuf)
	case 0xD1:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xD2:
		cpu.kil(&cpu.stepInfoBuf)
	case 0xD3:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xD4:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xD5:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xD6:
		cpu.dec(&cpu.stepInfoBuf)
	case 0xD7:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xD8:
		cpu.cld(&cpu.stepInfoBuf)
	case 0xD9:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xDA:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xDB:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xDC:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xDD:
		cpu.cmp(&cpu.stepInfoBuf)
	case 0xDE:
		cpu.dec(&cpu.stepInfoBuf)
	case 0xDF:
		cpu.dcp(&cpu.stepInfoBuf)
	case 0xE0:
		cpu.cpx(&cpu.stepInfoBuf)
	case 0xE1:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xE2:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xE3:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xE4:
		cpu.cpx(&cpu.stepInfoBuf)
	case 0xE5:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xE6:
		cpu.inc(&cpu.stepInfoBuf)
	case 0xE7:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xE8:
		cpu.inx(&cpu.stepInfoBuf)
	case 0xE9:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xEA:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xEB:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xEC:
		cpu.cpx(&cpu.stepInfoBuf)
	case 0xED:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xEE:
		cpu.inc(&cpu.stepInfoBuf)
	case 0xEF:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xF0:
		cpu.beq(&cpu.stepInfoBuf)
	case 0xF1:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xF2:
		cpu.kil(&cpu.stepInfoBuf)
	case 0xF3:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xF4:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xF5:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xF6:
		cpu.inc(&cpu.stepInfoBuf)
	case 0xF7:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xF8:
		cpu.sed(&cpu.stepInfoBuf)
	case 0xF9:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xFA:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xFB:
		cpu.isc(&cpu.stepInfoBuf)
	case 0xFC:
		cpu.nop(&cpu.stepInfoBuf)
	case 0xFD:
		cpu.sbc(&cpu.stepInfoBuf)
	case 0xFE:
		cpu.inc(&cpu.stepInfoBuf)
	case 0xFF:
		cpu.isc(&cpu.stepInfoBuf)
	}

	return int(cpu.Cycles - cycles)
}

// NMI - Non-Maskable Interrupt
func (cpu *CPU) nmi() {
	cpu.push16(cpu.PC)
	cpu.php(nil)
	cpu.PC = cpu.Read16(0xFFFA)
	cpu.I = 1
	cpu.Cycles += 7
}

// IRQ - IRQ Interrupt
func (cpu *CPU) irq() {
	cpu.push16(cpu.PC)
	cpu.php(nil)
	cpu.PC = cpu.Read16(0xFFFE)
	cpu.I = 1
	cpu.Cycles += 7
}

// ADC - Add with Carry
func (cpu *CPU) adc(info *stepInfo) {
	a := cpu.A
	b := cpu.Read(info.address)
	c := cpu.C
	cpu.A = a + b + c
	cpu.setZN(cpu.A)
	if int(a)+int(b)+int(c) > 0xFF {
		cpu.C = 1
	} else {
		cpu.C = 0
	}
	if (a^b)&0x80 == 0 && (a^cpu.A)&0x80 != 0 {
		cpu.V = 1
	} else {
		cpu.V = 0
	}
}

// AND - Logical AND
func (cpu *CPU) and(info *stepInfo) {
	cpu.A = cpu.A & cpu.Read(info.address)
	cpu.setZN(cpu.A)
}

// ASL - Arithmetic Shift Left
func (cpu *CPU) asl(info *stepInfo) {
	if info.mode == modeAccumulator {
		cpu.C = (cpu.A >> 7) & 1
		cpu.A <<= 1
		cpu.setZN(cpu.A)
	} else {
		value := cpu.Read(info.address)
		cpu.C = (value >> 7) & 1
		value <<= 1
		cpu.Write(info.address, value)
		cpu.setZN(value)
	}
}

// BCC - Branch if Carry Clear
func (cpu *CPU) bcc(info *stepInfo) {
	if cpu.C == 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BCS - Branch if Carry Set
func (cpu *CPU) bcs(info *stepInfo) {
	if cpu.C != 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BEQ - Branch if Equal
func (cpu *CPU) beq(info *stepInfo) {
	if cpu.Z != 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BIT - Bit Test
func (cpu *CPU) bit(info *stepInfo) {
	value := cpu.Read(info.address)
	cpu.V = (value >> 6) & 1
	cpu.setZ(value & cpu.A)
	cpu.setN(value)
}

// BMI - Branch if Minus
func (cpu *CPU) bmi(info *stepInfo) {
	if cpu.N != 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BNE - Branch if Not Equal
func (cpu *CPU) bne(info *stepInfo) {
	if cpu.Z == 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BPL - Branch if Positive
func (cpu *CPU) bpl(info *stepInfo) {
	if cpu.N == 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BRK - Force Interrupt
func (cpu *CPU) brk(info *stepInfo) {
	cpu.push16(cpu.PC)
	cpu.php(info)
	cpu.sei(info)
	cpu.PC = cpu.Read16(0xFFFE)
}

// BVC - Branch if Overflow Clear
func (cpu *CPU) bvc(info *stepInfo) {
	if cpu.V == 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// BVS - Branch if Overflow Set
func (cpu *CPU) bvs(info *stepInfo) {
	if cpu.V != 0 {
		cpu.PC = info.address
		cpu.addBranchCycles(info)
	}
}

// CLC - Clear Carry Flag
func (cpu *CPU) clc(info *stepInfo) {
	cpu.C = 0
}

// CLD - Clear Decimal Mode
func (cpu *CPU) cld(info *stepInfo) {
	cpu.D = 0
}

// CLI - Clear Interrupt Disable
func (cpu *CPU) cli(info *stepInfo) {
	cpu.I = 0
}

// CLV - Clear Overflow Flag
func (cpu *CPU) clv(info *stepInfo) {
	cpu.V = 0
}

// CMP - Compare
func (cpu *CPU) cmp(info *stepInfo) {
	value := cpu.Read(info.address)
	cpu.compare(cpu.A, value)
}

// CPX - Compare X Register
func (cpu *CPU) cpx(info *stepInfo) {
	value := cpu.Read(info.address)
	cpu.compare(cpu.X, value)
}

// CPY - Compare Y Register
func (cpu *CPU) cpy(info *stepInfo) {
	value := cpu.Read(info.address)
	cpu.compare(cpu.Y, value)
}

// DEC - Decrement Memory
func (cpu *CPU) dec(info *stepInfo) {
	value := cpu.Read(info.address) - 1
	cpu.Write(info.address, value)
	cpu.setZN(value)
}

// DEX - Decrement X Register
func (cpu *CPU) dex(info *stepInfo) {
	cpu.X--
	cpu.setZN(cpu.X)
}

// DEY - Decrement Y Register
func (cpu *CPU) dey(info *stepInfo) {
	cpu.Y--
	cpu.setZN(cpu.Y)
}

// EOR - Exclusive OR
func (cpu *CPU) eor(info *stepInfo) {
	cpu.A = cpu.A ^ cpu.Read(info.address)
	cpu.setZN(cpu.A)
}

// INC - Increment Memory
func (cpu *CPU) inc(info *stepInfo) {
	value := cpu.Read(info.address) + 1
	cpu.Write(info.address, value)
	cpu.setZN(value)
}

// INX - Increment X Register
func (cpu *CPU) inx(info *stepInfo) {
	cpu.X++
	cpu.setZN(cpu.X)
}

// INY - Increment Y Register
func (cpu *CPU) iny(info *stepInfo) {
	cpu.Y++
	cpu.setZN(cpu.Y)
}

// JMP - Jump
func (cpu *CPU) jmp(info *stepInfo) {
	cpu.PC = info.address
}

// JSR - Jump to Subroutine
func (cpu *CPU) jsr(info *stepInfo) {
	cpu.push16(cpu.PC - 1)
	cpu.PC = info.address
}

// LDA - Load Accumulator
func (cpu *CPU) lda(info *stepInfo) {
	cpu.A = cpu.Read(info.address)
	cpu.setZN(cpu.A)
}

// LDX - Load X Register
func (cpu *CPU) ldx(info *stepInfo) {
	cpu.X = cpu.Read(info.address)
	cpu.setZN(cpu.X)
}

// LDY - Load Y Register
func (cpu *CPU) ldy(info *stepInfo) {
	cpu.Y = cpu.Read(info.address)
	cpu.setZN(cpu.Y)
}

// LSR - Logical Shift Right
func (cpu *CPU) lsr(info *stepInfo) {
	if info.mode == modeAccumulator {
		cpu.C = cpu.A & 1
		cpu.A >>= 1
		cpu.setZN(cpu.A)
	} else {
		value := cpu.Read(info.address)
		cpu.C = value & 1
		value >>= 1
		cpu.Write(info.address, value)
		cpu.setZN(value)
	}
}

// NOP - No Operation
func (cpu *CPU) nop(info *stepInfo) {
}

// ORA - Logical Inclusive OR
func (cpu *CPU) ora(info *stepInfo) {
	cpu.A = cpu.A | cpu.Read(info.address)
	cpu.setZN(cpu.A)
}

// PHA - Push Accumulator
func (cpu *CPU) pha(info *stepInfo) {
	cpu.push(cpu.A)
}

// PHP - Push Processor Status
func (cpu *CPU) php(info *stepInfo) {
	cpu.push(cpu.Flags() | 0x10)
}

// PLA - Pull Accumulator
func (cpu *CPU) pla(info *stepInfo) {
	cpu.A = cpu.pull()
	cpu.setZN(cpu.A)
}

// PLP - Pull Processor Status
func (cpu *CPU) plp(info *stepInfo) {
	cpu.SetFlags(cpu.pull()&0xEF | 0x20)
}

// ROL - Rotate Left
func (cpu *CPU) rol(info *stepInfo) {
	if info.mode == modeAccumulator {
		c := cpu.C
		cpu.C = (cpu.A >> 7) & 1
		cpu.A = (cpu.A << 1) | c
		cpu.setZN(cpu.A)
	} else {
		c := cpu.C
		value := cpu.Read(info.address)
		cpu.C = (value >> 7) & 1
		value = (value << 1) | c
		cpu.Write(info.address, value)
		cpu.setZN(value)
	}
}

// ROR - Rotate Right
func (cpu *CPU) ror(info *stepInfo) {
	if info.mode == modeAccumulator {
		c := cpu.C
		cpu.C = cpu.A & 1
		cpu.A = (cpu.A >> 1) | (c << 7)
		cpu.setZN(cpu.A)
	} else {
		c := cpu.C
		value := cpu.Read(info.address)
		cpu.C = value & 1
		value = (value >> 1) | (c << 7)
		cpu.Write(info.address, value)
		cpu.setZN(value)
	}
}

// RTI - Return from Interrupt
func (cpu *CPU) rti(info *stepInfo) {
	cpu.SetFlags(cpu.pull()&0xEF | 0x20)
	cpu.PC = cpu.pull16()
}

// RTS - Return from Subroutine
func (cpu *CPU) rts(info *stepInfo) {
	cpu.PC = cpu.pull16() + 1
}

// SBC - Subtract with Carry
func (cpu *CPU) sbc(info *stepInfo) {
	a := cpu.A
	b := cpu.Read(info.address)
	c := cpu.C
	cpu.A = a - b - (1 - c)
	cpu.setZN(cpu.A)
	if int(a)-int(b)-int(1-c) >= 0 {
		cpu.C = 1
	} else {
		cpu.C = 0
	}
	if (a^b)&0x80 != 0 && (a^cpu.A)&0x80 != 0 {
		cpu.V = 1
	} else {
		cpu.V = 0
	}
}

// SEC - Set Carry Flag
func (cpu *CPU) sec(info *stepInfo) {
	cpu.C = 1
}

// SED - Set Decimal Flag
func (cpu *CPU) sed(info *stepInfo) {
	cpu.D = 1
}

// SEI - Set Interrupt Disable
func (cpu *CPU) sei(info *stepInfo) {
	cpu.I = 1
}

// STA - Store Accumulator
func (cpu *CPU) sta(info *stepInfo) {
	cpu.Write(info.address, cpu.A)
}

// STX - Store X Register
func (cpu *CPU) stx(info *stepInfo) {
	cpu.Write(info.address, cpu.X)
}

// STY - Store Y Register
func (cpu *CPU) sty(info *stepInfo) {
	cpu.Write(info.address, cpu.Y)
}

// TAX - Transfer Accumulator to X
func (cpu *CPU) tax(info *stepInfo) {
	cpu.X = cpu.A
	cpu.setZN(cpu.X)
}

// TAY - Transfer Accumulator to Y
func (cpu *CPU) tay(info *stepInfo) {
	cpu.Y = cpu.A
	cpu.setZN(cpu.Y)
}

// TSX - Transfer Stack Pointer to X
func (cpu *CPU) tsx(info *stepInfo) {
	cpu.X = cpu.SP
	cpu.setZN(cpu.X)
}

// TXA - Transfer X to Accumulator
func (cpu *CPU) txa(info *stepInfo) {
	cpu.A = cpu.X
	cpu.setZN(cpu.A)
}

// TXS - Transfer X to Stack Pointer
func (cpu *CPU) txs(info *stepInfo) {
	cpu.SP = cpu.X
}

// TYA - Transfer Y to Accumulator
func (cpu *CPU) tya(info *stepInfo) {
	cpu.A = cpu.Y
	cpu.setZN(cpu.A)
}

// illegal opcodes below

func (cpu *CPU) ahx(info *stepInfo) {
}

func (cpu *CPU) alr(info *stepInfo) {
}

func (cpu *CPU) anc(info *stepInfo) {
}

func (cpu *CPU) arr(info *stepInfo) {
}

func (cpu *CPU) axs(info *stepInfo) {
}

func (cpu *CPU) dcp(info *stepInfo) {
}

func (cpu *CPU) isc(info *stepInfo) {
}

func (cpu *CPU) kil(info *stepInfo) {
}

func (cpu *CPU) las(info *stepInfo) {
}

func (cpu *CPU) lax(info *stepInfo) {
}

func (cpu *CPU) rla(info *stepInfo) {
}

func (cpu *CPU) rra(info *stepInfo) {
}

func (cpu *CPU) sax(info *stepInfo) {
}

func (cpu *CPU) shx(info *stepInfo) {
}

func (cpu *CPU) shy(info *stepInfo) {
}

func (cpu *CPU) slo(info *stepInfo) {
}

func (cpu *CPU) sre(info *stepInfo) {
}

func (cpu *CPU) tas(info *stepInfo) {
}

func (cpu *CPU) xaa(info *stepInfo) {
}
