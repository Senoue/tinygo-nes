# Local ILI9341 driver

Copied from tinygo.org/x/drivers/ili9341 v0.34.0 (BSD-3-Clause; see LICENSE).

ATSAMD51 SPI writes wait for INTFLAG.TXC before returning, allowing the
caller to change DC or CS only after the final bit has been transmitted.
Empty slices and nonpositive repeat counts return without waiting.
Other platform implementations are unchanged. This local copy avoids
modifying the Go module cache. With `wioterminal,lcddma`, bulk byte slices use synchronous DMAC channel 0
on SPI3/SERCOM7. This experiment owns the entire DMAC descriptor tables;
it refuses an already enabled controller. Commands and other write types
still use polling. See OVERCLOCK.md for build and validation instructions.
