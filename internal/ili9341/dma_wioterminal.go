//go:build wioterminal && lcddma

package ili9341

import (
	"device/arm"
	"device/sam"
	"machine"
	"runtime"
	"runtime/volatile"
	"time"
	"unsafe"
)

// This experiment owns DMAC channel 0 and the controller's descriptor tables.
// TinyGo 0.43.0's Wio Terminal SPI and the buzzer do not use DMAC.
// Refuse an already enabled controller rather than overwrite another owner.
// SAMD51 descriptors and writeback tables must be 16-byte aligned in SRAM.
var lcdDescriptors [32*16 + 15]byte
var lcdWriteback [32*16 + 15]byte
var lcdDMAReady bool

type lcdDescriptor struct {
	control volatile.Register16
	count   volatile.Register16
	source  volatile.Register32
	dest    volatile.Register32
	next    volatile.Register32
}

func alignedLCDTable(storage *[32*16 + 15]byte) uintptr {
	return (uintptr(unsafe.Pointer(&storage[0])) + 15) &^ uintptr(15)
}

func initLCDDMA() {
	if lcdDMAReady {
		return
	}
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_DMAC_)
	if sam.DMAC.CTRL.HasBits(sam.DMAC_CTRL_DMAENABLE) {
		panic("LCD DMA: controller already in use")
	}
	sam.DMAC.CTRL.Set(sam.DMAC_CTRL_SWRST)
	for sam.DMAC.CTRL.HasBits(sam.DMAC_CTRL_SWRST) {
	}
	sam.DMAC.BASEADDR.Set(uint32(alignedLCDTable(&lcdDescriptors)))
	sam.DMAC.WRBADDR.Set(uint32(alignedLCDTable(&lcdWriteback)))
	sam.DMAC.CTRL.Set(sam.DMAC_CTRL_DMAENABLE | sam.DMAC_CTRL_LVLEN0)
	ch := &sam.DMAC.CHANNEL[0]
	// SERCOM7 TX request = 19 (SAMD51 instance/sercom7.h).
	// One byte per DRE request: BURST action, SINGLE burst length.
	ch.CHCTRLA.Set(19<<sam.DMAC_CHANNEL_CHCTRLA_TRIGSRC_Pos |
		sam.DMAC_CHANNEL_CHCTRLA_TRIGACT_BURST<<sam.DMAC_CHANNEL_CHCTRLA_TRIGACT_Pos)
	ch.CHINTENCLR.Set(0xff) // Poll completion; do not enable a DMA IRQ.
	lcdDMAReady = true
}

// dmaWrite sends only bulk LCD data. Commands keep the polling path.
// The caller keeps CS asserted and waits for SPI TXC after this returns.
func (pd *spiDriver) dmaWrite(b []byte) bool {
	if pd.bus != machine.SPI3 || len(b) < 32 {
		return false
	}
	initLCDDMA()
	ch := &sam.DMAC.CHANNEL[0]
	desc := (*lcdDescriptor)(unsafe.Pointer(alignedLCDTable(&lcdDescriptors)))
	for offset := 0; offset < len(b); {
		n := len(b) - offset
		if n > 65535 {
			n = 65535
		}
		// Channel is disabled between blocks. Source increments; destination
		// stays at DATA. An incrementing address is stored as its end address.
		desc.control.Set(0)
		desc.count.Set(uint16(n))
		desc.source.Set(uint32(uintptr(unsafe.Pointer(&b[offset])) + uintptr(n)))
		desc.dest.Set(uint32(uintptr(unsafe.Pointer(&pd.bus.Bus.DATA))))
		desc.next.Set(0)
		// VALID | SRCINC, byte beat size, no linked descriptor.
		desc.control.Set(1 | 1<<10)
		ch.CHINTFLAG.Set(0x07)
		arm.Asm("dmb sy")
		ch.CHCTRLA.SetBits(sam.DMAC_CHANNEL_CHCTRLA_ENABLE)
		deadline := time.Now().Add(100 * time.Millisecond)
		for {
			flags := ch.CHINTFLAG.Get()
			if flags&sam.DMAC_CHANNEL_CHINTFLAG_TERR != 0 {
				ch.CHCTRLA.ClearBits(sam.DMAC_CHANNEL_CHCTRLA_ENABLE)
				panic("LCD DMA: transfer error")
			}
			if flags&sam.DMAC_CHANNEL_CHINTFLAG_TCMPL != 0 {
				break
			}
			if time.Now().After(deadline) {
				ch.CHCTRLA.ClearBits(sam.DMAC_CHANNEL_CHCTRLA_ENABLE)
				panic("LCD DMA: timeout")
			}
		}
		ch.CHCTRLA.ClearBits(sam.DMAC_CHANNEL_CHCTRLA_ENABLE)
		for ch.CHCTRLA.HasBits(sam.DMAC_CHANNEL_CHCTRLA_ENABLE) {
		}
		arm.Asm("dmb sy")
		offset += n
	}
	runtime.KeepAlive(b)
	return true
}
