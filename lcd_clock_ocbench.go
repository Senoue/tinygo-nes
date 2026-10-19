//go:build wioterminal && ocbench

package main

// Both clock experiments select the independent 48 MHz GCLK1 divided by two.
const lcdSPIFrequency = 24000000
