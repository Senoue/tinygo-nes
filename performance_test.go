//go:build perf

package main

import (
	"github.com/Senoue/tinygo-nes/nes"
	"testing"
	"time"
)

func TestPerformanceMetrics(t *testing.T) {
	p := performanceStats{
		frames: 120, displayed: 60,
		emulation:  3 * time.Second,
		conversion: 300 * time.Millisecond,
		transfer:   900 * time.Millisecond,
	}
	// Wall time includes input, audio, and instrumentation overhead too.
	got := p.metrics(5 * time.Second)
	want := performanceMetrics{240, 120, 25000, 5000, 15000}
	if got != want {
		t.Fatalf("metrics = %+v, want %+v", got, want)
	}
}

func TestPerformanceMetricsWithoutDisplay(t *testing.T) {
	p := performanceStats{frames: 120, emulation: time.Second}
	m := p.metrics(2 * time.Second)
	if m.emulationFPS10 != 600 || m.displayFPS10 != 0 || m.conversionUS != 0 || m.transferUS != 0 {
		t.Fatalf("metrics = %+v", m)
	}
	if got := new(performanceStats).metrics(0); got != (performanceMetrics{}) {
		t.Fatalf("empty metrics = %+v", got)
	}
}

func TestPerformanceWindow(t *testing.T) {
	var p performanceStats
	p.startWindow()
	for i := 0; i < performanceWindowFrames-1; i++ {
		p.finishFrame(i%2 == 0)
	}
	if p.frames != 119 || p.displayed != 60 {
		t.Fatalf("counts = %d/%d", p.frames, p.displayed)
	}
	p.emulation = time.Second
	p.finishFrame(false)
	if p.frames != 0 || p.displayed != 0 || p.emulation != 0 || p.conversion != 0 || p.transfer != 0 || p.started.IsZero() {
		t.Fatal("measurement window was not reset")
	}
}

func TestCoreSampleSelection(t *testing.T) {
	c, err := nes.NewConsole(nes.NewCartridge(make([]byte, 16384), make([]byte, 8192), 0, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	p := performanceStats{frames: performanceWindowFrames - 3}
	p.stepFrame(c, true)
	if p.core != ([2]nes.FrameProfile{}) || p.profiled {
		t.Fatal("sampled a regular frame")
	}
	p.frames++
	c.SetFrameOutput(false)
	p.stepFrame(c, false)
	if p.core[0].Total <= 0 || p.samples != ([2]int64{1, 0}) {
		t.Fatal("missing skipped sample")
	}
	p.frames++
	c.SetFrameOutput(true)
	p.stepFrame(c, true)
	if p.core[1].Total <= 0 || p.samples != ([2]int64{1, 1}) {
		t.Fatal("missing drawn sample")
	}
	p.emulation = 120*time.Millisecond + 118*80*time.Millisecond
	p.sampledEmulation = 120 * time.Millisecond
	p.frames++
	if got := p.regularEmulationUS(); got != 80000 {
		t.Fatalf("regular mean = %d, want 80000", got)
	}
	p.startWindow()
	if p.core != ([2]nes.FrameProfile{}) || p.sampledEmulation != 0 || p.samples != ([2]int64{}) || p.profiled {
		t.Fatal("core samples were not cleared")
	}
}
