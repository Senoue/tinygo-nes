package nes

import "testing"

func frameIRQTestConsole(t *testing.T) *Console {
	t.Helper()
	prg := make([]byte, 16384)
	for i := range prg {
		prg[i] = 0xea
	}
	prg[0x3ffc], prg[0x3ffd] = 0, 0x80
	prg[0x3ffe], prg[0x3fff] = 0, 0x90
	prg[0x3ffa], prg[0x3ffb] = 0, 0xa0
	c, err := NewConsole(NewCartridge(prg, make([]byte, 8192), 0, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestAPUFrameIRQStatusWhileCPUMasked(t *testing.T) {
	for _, batch := range []bool{false, true} {
		c := frameIRQTestConsole(t)
		a := c.APU
		c.CPU.I = 1
		a.writeFrameCounter(0)
		a.frameValue, a.frameCounter = 2, 1
		if batch {
			a.StepCycles(1)
		} else {
			a.Step()
		}
		if !a.frameIRQPending || c.CPU.interrupt != interruptNone {
			t.Fatal("masked IRQ must remain in APU status, not be discarded or queued as an edge")
		}
		if status := c.CPU.Read(0x4015); status&0x40 == 0 {
			t.Fatalf("frame IRQ missing from status: %02x", status)
		}
		if a.frameIRQPending || c.CPU.Read(0x4015)&0x40 != 0 {
			t.Fatal("status read did not acknowledge frame IRQ")
		}
	}
}

func TestAPUFrameIRQAcknowledgement(t *testing.T) {
	c := frameIRQTestConsole(t)
	a := c.APU
	a.writeFrameCounter(0)
	a.fireIRQ()
	a.writeControl(15)
	if !a.frameIRQPending {
		t.Fatal("$4015 write must not acknowledge frame IRQ")
	}
	a.writeFrameCounter(0x80)
	if !a.frameIRQPending {
		t.Fatal("5-step mode without inhibit must not acknowledge existing IRQ")
	}
	a.writeFrameCounter(0x40)
	if a.frameIRQPending {
		t.Fatal("$4017 IRQ inhibit must acknowledge IRQ")
	}
	a.fireIRQ()
	if a.frameIRQPending {
		t.Fatal("inhibited frame counter generated IRQ")
	}
}

func TestAPUFrameIRQDeliveredAfterCLI(t *testing.T) {
	c := frameIRQTestConsole(t)
	c.Cartridge.PRG[0] = 0x58 // CLI
	c.APU.frameIRQPending = true
	c.CPU.Step()
	if c.CPU.PC != 0x8001 || c.CPU.SP != 0xfd {
		t.Fatal("IRQ serviced while masked")
	}
	c.CPU.Step()
	if c.CPU.PC != 0x9001 || c.CPU.SP != 0xfa {
		t.Fatal("pending APU IRQ not serviced after CLI")
	}
	if !c.APU.frameIRQPending {
		t.Fatal("entering handler must not acknowledge source")
	}
	c.CPU.Read(0x4015)
	c.CPU.I = 0
	c.CPU.Step()
	if c.CPU.PC != 0x9002 || c.CPU.SP != 0xfa {
		t.Fatal("acknowledged IRQ was serviced again")
	}
}

func TestAPUFrameIRQDoesNotOverrideNMI(t *testing.T) {
	c := frameIRQTestConsole(t)
	c.CPU.I = 0
	c.APU.frameIRQPending = true
	c.CPU.triggerNMI()
	c.CPU.Step()
	if c.CPU.PC != 0xa001 || !c.APU.frameIRQPending {
		t.Fatal("NMI priority or pending IRQ was lost")
	}
}

func TestPollingAPUFrameIRQStartsTone(t *testing.T) {
	c := frameIRQTestConsole(t)
	// Synthetic sound driver: keep IRQ masked, poll $4015.6, then play A4.
	copy(c.Cartridge.PRG, []byte{
		0x78,                         // SEI
		0xa9, 0x0f, 0x8d, 0x15, 0x40, // Enable channels.
		0xa9, 0x00, 0x8d, 0x17, 0x40, // Enable frame IRQ generation.
		0xad, 0x15, 0x40, 0x29, 0x40, // Poll and mask frame IRQ status.
		0xf0, 0xf9, // BEQ back to poll.
		0xa9, 0x9f, 0x8d, 0x00, 0x40, // Pulse volume 15.
		0xa9, 0xfd, 0x8d, 0x02, 0x40, // Timer period 253 -> 440 Hz.
		0xa9, 0x08, 0x8d, 0x03, 0x40, // Load length counter.
		0x4c, 0x0b, 0x80, // Wait for next frame IRQ.
	})
	for i := 0; i < 5; i++ {
		c.StepFrame()
	}
	if c.CPU.I != 1 {
		t.Fatal("test driver must keep CPU IRQ masked")
	}
	if got := c.APU.BuzzerTone(); got != (BuzzerTone{440, 15}) {
		t.Fatalf("polling sound driver did not start tone: %+v", got)
	}
}
