package main

import "github.com/remogatto/z80"
import "github.com/pnegre/gogame"
import "log"
import "time"
import "os"
import "bufio"

const (
	WINTITLE  = "gomsx"
	WIN_W     = 800
	WIN_H     = 600
	ROMFILE   = "msx1.rom"
	NANOS_SCR = 20000000 // 50Hz -> Interval de 20Mseg
)

func readFile(fname string) ([]byte, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	stats, statsErr := f.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(f)
	_, err = bufr.Read(bytes)

	return bytes, err
}

func main() {
	if err := gogame.Init(WINTITLE, WIN_W, WIN_H); err != nil {
		log.Fatal(err)
	}
	defer gogame.Quit()

	memory := NewMemory()
	buffer, err := readFile(ROMFILE)
	if err != nil {
		log.Fatal(err)
	}
	memory.load(buffer, 0, 0)
	memory.load(buffer[0x4000:], 1, 0)

	ports := new(Ports)
	cpuZ80 := z80.NewZ80(memory, ports)
	cpuZ80.Reset()
	cpuZ80.SetPC(0)
	log.Println("Beginning simulation...")
	lastTm := time.Now().UnixNano()
	delta := int64(0)
	logAssembler := false
	for {
		for i := 0; i < 500; i++ {
			if logAssembler {
				pc := cpuZ80.PC()
				instr, _, _ := z80.Disassemble(memory, pc, 0)
				log.Printf("%04x: %s\n", pc, instr)
			}
			cpuZ80.DoOpcode()
		}

		delta = time.Now().UnixNano() - lastTm
		if delta > NANOS_SCR {
			if quit := gogame.SlurpEvents(); quit == true {
				break
			}

			graphics_renderScreen()
			if vdp_enabledInterrupts {
				cpuZ80.Interrupt()
			}
			lastTm = time.Now().UnixNano()
			gogame.Delay(1)
		}
	}
}
