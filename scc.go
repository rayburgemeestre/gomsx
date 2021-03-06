package main

import "log"

const (
	SCC_NCHANNELS = 5
)

type SCCChannel struct {
	tonegenerator *ToneGenerator
	waveform      [32]byte
	sccRegFreq    [2]byte
}

var scc_channels []*SCCChannel

func init() {
	for i := 0; i < SCC_NCHANNELS; i++ {
		scc_channels = append(scc_channels, NewSCCChannel())
	}
}

func NewSCCChannel() *SCCChannel {
	return &SCCChannel{tonegenerator: NewToneGenerator()}
}

func (self *SCCChannel) feedSamples(data []int16) {
	self.tonegenerator.feedSamples(data)
}

func (self *SCCChannel) setWaveform(k uint16, b byte) {
	self.waveform[k] = b
	self.tonegenerator.updateWaveform(self.waveform[:])
}

func scc_feedSamples(data []int16) {
	for _, c := range scc_channels {
		c.feedSamples(data)
	}
}

func scc_write(n uint16, b byte) {
	switch {
	case n >= 0x00 && n < 0x80:
		// Waveform (32 bits)
		nch := n / 0x20
		k := n % 0x20
		scc_channels[nch].setWaveform(k, b)
		// Waveform is shared for ch4 and ch5
		if nch == 3 {
			scc_channels[4].setWaveform(k, b)
		}

	case n >= 0x80 && n < 0x8a:
		// Frequency
		nch := (n - 0x80) / 2
		scc_channels[nch].sccRegFreq[n%2] = b
		freq := int(scc_channels[nch].sccRegFreq[0]) | (int(scc_channels[nch].sccRegFreq[1]&0x0f) << 8)
		if freq > 0 {
			realFreq := float32(111861) / float32(freq)
			scc_channels[nch].tonegenerator.setFrequency(realFreq)
		}

	case n >= 0x8a && n < 0x8f:
		// Volume
		nch := n - 0x8a
		scc_channels[nch].tonegenerator.setVolume(float32(b))

	case n == 0x8f:
		// ON/OFF switch channel 1 to 5
		for i := 0; i < 5; i++ {
			var m byte = 0x01 << uint(i)
			if m&b != 0 {
				scc_channels[i].tonegenerator.activate(true)
			} else {
				scc_channels[i].tonegenerator.activate(false)
			}
		}
	default:
		log.Printf("SCC: %x -> %d\n", n, b)
	}
}
