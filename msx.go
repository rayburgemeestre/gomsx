package main

import (
	"log"
	"time"

	"github.com/rayburgemeestre/gogame"
	"github.com/rayburgemeestre/gomsx/z80"
)

type MSX struct {
	cpuz80 *z80.Z80
	vdp    *Vdp
	memory *Memory
	ppi    *PPI
	psg    *PSG
}

func (msx *MSX) mainLoop(frameInterval int) (bool, float64) {
	log.Println("Beginning simulation...")
	state_init()
	var currentTime, elapsedTime, lag int64
	updateInterval := int64(time.Millisecond) * int64(frameInterval)
	// For testing purposes: super fast execution
	//updateInterval = int64(time.Millisecond)

	// Four times as fast
	updateInterval /= 4
	previousTime := time.Now().UnixNano()

	beginTime := time.Now()
	startTime := beginTime.UnixNano()
	nframes := 0
	gogame.SetFullScreen(true)
	for {
		currentTime = time.Now().UnixNano()
		elapsedTime = currentTime - previousTime
		// Exit the mainLoop every minute, so our caller can load a new random cartridge.
		if time.Now().Sub(beginTime) > time.Second*10 {
			return false, 0
		}
		previousTime = currentTime
		lag += elapsedTime
		for lag >= updateInterval {
			msx.cpuFrame()
			lag -= updateInterval
		}

		if quit := gogame.SlurpEvents(); quit == true {
			break
		}

		graphics_lock()
		msx.vdp.updateBuffer()
		graphics_unlock()
		graphics_render()

		nframes++

		// let's dial back the GPU or CPU usage when the screensaver is active
		time.Sleep(20 * time.Millisecond)
	}
	delta := (time.Now().UnixNano() - startTime) / int64(time.Second)
	return true, float64(nframes) / float64(delta)
}

func (msx *MSX) cpuFrame() {
	msx.cpuz80.Cycles %= CYCLESPERFRAME
	for msx.cpuz80.Cycles < CYCLESPERFRAME {
		if msx.cpuz80.Halted == true {
			break
		}
		msx.cpuz80.DoOpcode()
	}

	if msx.vdp.enabledInterrupts {
		msx.vdp.setFrameFlag()
		msx.cpuz80.Interrupt()
	}
}
