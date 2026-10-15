//go:build !wioterminal || !buzzer

package main

import "github.com/Senoue/tinygo-nes/nes"

func InitAudio()            {}
func OutputAudio(*nes.APU)  {}
func HandleAudioInput(bool) {}
