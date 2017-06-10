package main

import "log"

var sound_regs [16]byte
var sound_regNext byte

/*

	Exemple en msx basic:

	sound 0,105         // Set tone A frequency
	sound 1,0
	sound 7, &B10111110 // enable tone generator A
	sound 8, &B00001111 // Set amplitude for channel A

*/

func sound_writePort(ad byte, val byte) {
	switch {
	case ad == 0xa0:
		// Register write port
		sound_regNext = val
		return

	case ad == 0xa1:
		// Write value to port
		sound_regs[sound_regNext] = val
		if sound_regNext < 14 {
			sound_work()
		}
		return
	}

	log.Fatalf("Sound, not implemented: out(%02x,%02x)", ad, val)
}

func sound_readPort(ad byte) byte {
	switch {
	case ad == 0xa2:
		// Read value from port
		if sound_regNext == 0x0e {
			// joystick triggers.
			// Per ara ho posem a 1 (no moviment de joystick)
			return 0x3f
		}
		if sound_regNext == 0x0f {
			// PSG port 15 (joystick select)
			// TODO: millorar
			return 0
		}
		return sound_regs[sound_regNext]
	}

	log.Fatalf("Sound, not implemented: in(%02x)", ad)
	return 0
}

func sound_work() {
	log.Println(sound_regs)
	freqA := (uint(sound_regs[1]&0x0f) << 8) | uint(sound_regs[0])
	if freqA > 0 {
		realFreqA := 111861 / freqA
		log.Printf("Freq A: %d, real: %d\n", freqA, realFreqA)
	}
}
