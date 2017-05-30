package main

import "log"

type Ports struct {
}

func (self *Ports) ReadPort(address uint16) byte {
	ad := byte(address & 0xFF)
	switch {
	case ad >= 0xa8 && ad <= 0xab:
		return ppi_readPort(ad)
	}

	log.Fatalf("ReadPort: %02x\n", ad)
	return 0
}

func (self *Ports) WritePort(address uint16, b byte) {
	ad := byte(address & 0xFF)
	switch {
	case ad >= 0xa8 && ad <= 0xab:
		ppi_writePort(ad, b)
		return
	}

	log.Fatalf("Writeport: %02x -> %02x\n", ad, b)
}

func (self *Ports) ReadPortInternal(address uint16, contend bool) byte {
	panic("ReadPortInternal")
}

func (self *Ports) WritePortInternal(address uint16, b byte, contend bool) {
	panic("WritePortInternal")
}

func (self *Ports) ContendPortPreio(address uint16) {
	panic("ContendPortPreio")

}

func (self *Ports) ContendPortPostio(address uint16) {
	panic("ContendPortPostio")
}
