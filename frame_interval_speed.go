//go:build speed

package main

// Trade display updates for faster game execution. CPU/APU and input still
// advance every frame. Keep this a divisor of both 120 and 256 so profiling
// windows and the byte frame counter retain their phase.
const displayFrameInterval = 4
