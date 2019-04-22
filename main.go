package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/pnegre/gogame"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pnegre/gomsx/z80"
)

var (
	random        bool
	index         int
	cart          string
	earlyExit     bool
	visitedPPI    int
	font          *gogame.Font
	SYSTEMROMFILE string
	XMLDATABASE   string
)

const (
	// 60Hz -> Interval de 16mseg
	INTERVAL = 16
	// EL z80 va a 3.58 Mhz. Cada 16mseg passen 57280 cicles
	CYCLESPERFRAME = 60000
)

func forEachLineInFile(filename string, callback func(string) bool) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !callback(scanner.Text()) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

var lastProcess *os.Process

func main() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		_ = lastProcess.Kill()
		os.Exit(0)
	}()

	// Sometimes, especially with a multi-monitor setup the MSX emulator can crash, at least on my laptop
	// Unfortunately XScreensaver won't automatically restart it, so we build this workaround. Infinitely restart :)
	if os.Getenv("MSX_STARTED_SUBPROCESS") == "" {
		for {
			executable, err := os.Executable()
			if err != nil {
				panic(err.Error())
			}
			cmd := exec.Command(executable)
			cmd.Env = append(os.Environ(), "MSX_STARTED_SUBPROCESS=true")

			// Make sure when this process is killed, our spawned child is as well!
			//cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				// This is fine!
				log.Println(err)
				continue
			}
			lastProcess = cmd.Process
			if err := cmd.Wait(); err != nil {
				// This is fine!
				log.Println(err)
			}
		}
	}

	// Random game selection (sequential is handy when testing lots of roms)
	random = true

	// Get User Home directory
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	homedir := u.HomeDir

	msxFilesPath := fmt.Sprintf("%s/.msxsaver", homedir)
	if _, err := os.Stat(msxFilesPath); os.IsNotExist(err) {
		msxFilesPath = "/usr/share/msxsaver"
	}

	// Use Panasonic MSX1 CF-2700 ROM
	SYSTEMROMFILE = fmt.Sprintf("%s/cf-2700_basic-bios1.rom", msxFilesPath)

	// Use XML Database
	XMLDATABASE = fmt.Sprintf("%s/softwaredb.xml", msxFilesPath)

	// Load all the known to be working games
	var games []string
	forEachLineInFile(fmt.Sprintf("%s/games.txt", msxFilesPath), func(game string) bool {
		if len(game) == 0 || strings.Contains(game, "//") {
			return true
		}
		games = append(games, fmt.Sprintf("%s/roms/%s", msxFilesPath, game))
		return true
	})

	runtime.LockOSThread() // Assure SDL works...
	var systemRom string
	var quality bool
	var frameInterval int
	flag.StringVar(&cart, "cart", "", "ROM in SLOT 1")
	flag.StringVar(&systemRom, "sys", SYSTEMROMFILE, "System file")
	flag.BoolVar(&quality, "quality", true, "Best quality rendering")
	flag.IntVar(&frameInterval, "fint", INTERVAL, "Frame interval in milliseconds")
	flag.Parse()

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Seeding like this will make all screens show the same random game at once.
	rand.Seed(time.Now().Unix())

	// Seeding like this, including the pid, prevents this and all screens show a different random game.
	rand.Seed(time.Now().Unix() * int64(os.Getpid()))

	// Only initialize graphics once, we can re-use the SDL window for each game we load.
	initializeGraphics := true

	index = 0
	for {
		ppi := NewPPI()
		memory := NewMemory(ppi)
		memory.loadBiosBasic(systemRom)

		if random {
			index = rand.Intn(len(games))
			cart = games[index]
		} else {
			cart = games[index]
			index++
			if index == len(games) {
				index = 0
			}
		}

		// Debug log which rom is loaded
		f, err := os.OpenFile("/tmp/game_loaded.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err = f.WriteString(fmt.Sprintf("%d - %s\n", index, cart)); err != nil {
			panic(err)
		}

		memory.loadRom(cart, 1)

		psg := NewPSG()
		vdp := NewVdp()
		ports := &Ports{vdp: vdp, ppi: ppi, psg: psg}
		cpuZ80 := z80.NewZ80(memory, ports)
		cpuZ80.Reset()
		cpuZ80.SetPC(0)
		msx := &MSX{cpuz80: cpuZ80, vdp: vdp, memory: memory, ppi: ppi, psg: psg}

		if initializeGraphics {
			if errg := graphics_init(quality); errg != nil {
				log.Fatalf("Error initalizing graphics: %v", errg.Error())
			}

			// Load font to print current cartridge on screen
			font = gogame.NewFont(fmt.Sprintf("%s/Monaco_Linux-Powerline.ttf", msxFilesPath), 16)

			initializeGraphics = false
		}

		quit, avgFPS := msx.mainLoop(frameInterval)
		if quit {
			break
		}
		log.Printf("Avg FPS: %.2f\n", avgFPS)
	}
}
