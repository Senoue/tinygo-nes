//go:build !perf

package main

import "github.com/Senoue/tinygo-nes/nes"

// Empty hooks let the compiler remove instrumentation from ordinary builds.
type performanceStats struct{}

func performanceNow() int64                      { return 0 }
func (*performanceStats) startWindow()           {}
func (*performanceStats) recordEmulation(int64)  {}
func (*performanceStats) recordConversion(int64) {}
func (*performanceStats) recordTransfer(int64)   {}
func (*performanceStats) finishFrame(bool)       {}

func (*performanceStats) stepFrame(console *nes.Console, _ bool) { console.StepFrame() }
