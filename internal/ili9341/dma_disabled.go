//go:build (atsamd51 || atsame5x) && (!wioterminal || !lcddma)

package ili9341

func (pd *spiDriver) dmaWrite(b []byte) bool { return false }
