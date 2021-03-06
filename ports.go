package main

import "log"

type Ports struct {
	vdp *Vdp
	ppi *PPI
	psg *PSG
}

func (self *Ports) ReadPort(address uint16) byte {
	ad := byte(address & 0xFF)
	switch {
	case ad >= 0xa8 && ad <= 0xab:
		return self.ppi.readPort(ad)

	case ad >= 0xa0 && ad <= 0xa2:
		return self.psg.readPort(ad)

	case ad >= 0x98 && ad <= 0x9b:
		return self.vdp.readPort(ad)
	}

	log.Printf("ReadPort: %02x\n", ad)
	return 0
}

func (self *Ports) WritePort(address uint16, b byte) {
	ad := byte(address & 0xFF)
	switch {
	case ad >= 0xa8 && ad <= 0xab:
		self.ppi.writePort(ad, b)
		return

	case ad >= 0xa0 && ad <= 0xa2:
		self.psg.writePort(ad, b)
		return

	case ad >= 0x90 && ad <= 0x91:
		// Printer. Do nothing
		return

	case ad >= 0x98 && ad <= 0x9b:
		self.vdp.writePort(ad, b)
		return

	case ad >= 0x00 && ad <= 0x01:
		// MIDI / Sensor Kid
		return
	}

	log.Printf("Writeport: %02x -> %02x\n", ad, b)
}
