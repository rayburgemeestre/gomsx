package main

import "github.com/pnegre/gogame"
import "log"

var colors []*gogame.Color

func init() {
	colors = []*gogame.Color{
		&gogame.Color{0, 0, 0, 255},          // Transparent
		&gogame.Color{0, 0, 0, 255},          // Black
		&gogame.Color{0x20, 0xc8, 0x40, 255}, // Green
		&gogame.Color{0x58, 0xd8, 0x78, 255}, // Light Green
		&gogame.Color{0x50, 0x50, 0xe8, 255}, // Dark Blue
		&gogame.Color{0x78, 0x70, 0xf7, 255}, // Light Blue
		&gogame.Color{0xd0, 0x50, 0x48, 255}, // Dark Red
		&gogame.Color{0x40, 0xe8, 0xf0, 255}, // Cyan
		&gogame.Color{0xf7, 0x50, 0x50, 255}, // Red
		&gogame.Color{0xf7, 0x78, 0x78, 255}, // Bright Red
		&gogame.Color{0xd0, 0xc0, 0x50, 255}, // Yellow
		&gogame.Color{0xe0, 0xc8, 0x80, 255}, // Light Yellow
		&gogame.Color{0x20, 0xb0, 0x38, 255}, // Dark Green
		&gogame.Color{0xc8, 0x58, 0xb8, 255}, // Purple
		&gogame.Color{0xc8, 0xc8, 0xc8, 255}, // Gray
		&gogame.Color{0xf7, 0xf7, 0xf7, 255}, // White
	}
}

func graphics_setLogicalResolution() {
	switch vdp_screenMode {
	case SCREEN0:
		gogame.SetLogicalSize(320, 192)
		return
	case SCREEN2:
		gogame.SetLogicalSize(256, 192)
		return
	case SCREEN1:
		gogame.SetLogicalSize(256, 192)
		return
	}
	println(vdp_screenMode)
	panic("setLogicalResolution: mode not supported")
}

func graphics_renderScreen() {
	if !vdp_screenEnabled {
		return
	}
	nameTable := vdp_VRAM[(uint16(vdp_registers[2]) << 10):]
	patTable := vdp_VRAM[(uint16(vdp_registers[4]) << 11):]
	colorTable := vdp_VRAM[(uint16(vdp_registers[3]) << 6):]
	switch {
	case vdp_screenMode == SCREEN0:
		// Render SCREEN0 (40x24)
		color1 := colors[(vdp_registers[7]&0xF0)>>4]
		color2 := colors[(vdp_registers[7] & 0x0F)]
		for y := 0; y < 24; y++ {
			for x := 0; x < 40; x++ {
				graphics_drawPatternS0(x*8, y*8, int(nameTable[x+y*40])*8, patTable, color1, color2)
			}
		}
		break

	case vdp_screenMode == SCREEN1:
		// Render SCREEN1 (32x24)
		for y := 0; y < 24; y++ {
			for x := 0; x < 32; x++ {
				pat := int(nameTable[x+y*32])
				graphics_drawPatternS1(x*8, y*8, pat*8, patTable, colorTable[pat/8])
			}
		}
		graphics_drawSprites()
		break

	case vdp_screenMode == SCREEN2:
		// Render SCREEN2
		// Pattern table: 0000H to 17FFH
		// Name table: 1800H to 1AFFH
		// Color table: 2000H to 37FFH
		patTable := vdp_VRAM[(uint16(vdp_registers[4]&0x04) << 11):]
		colorTable := vdp_VRAM[(uint16(vdp_registers[3]&0x80) << 6):]
		for y := 0; y < 24; y++ {
			for x := 0; x < 32; x++ {
				pat := int(nameTable[x+y*32])
				graphics_drawPatternS2(x*8, y*8, pat*8, patTable, colorTable)
			}
		}
		graphics_drawSprites()
		break

	case vdp_screenMode == SCREEN3:
		// Render SCREEN3
		log.Println("Drawing in screen3 not implemented yet")
		break

	default:
		panic("RenderScreen: impossible mode")

	}
}

func graphics_drawPatternS0(x, y int, pt int, patTable []byte, color1, color2 *gogame.Color) {
	var mask byte
	for i := 0; i < 8; i++ {
		b := patTable[i+pt]
		xx := 0
		for mask = 0x80; mask > 0; mask >>= 1 {
			if mask&b != 0 {
				gogame.DrawPixel(x+xx, y+i, color1)
			} else {
				gogame.DrawPixel(x+xx, y+i, color2)
			}
			xx++
		}
	}
}
func graphics_drawPatternS1(x, y int, pt int, patTable []byte, color byte) {
	color1 := colors[(color&0xF0)>>4]
	color2 := colors[color&0x0F]
	var mask byte
	for i := 0; i < 8; i++ {
		b := patTable[i+pt]
		xx := 0
		for mask = 0x80; mask > 0; mask >>= 1 {
			if mask&b != 0 {
				gogame.DrawPixel(x+xx, y+i, color1)
			} else {
				gogame.DrawPixel(x+xx, y+i, color2)
			}
			xx++
		}
	}
}

func graphics_drawPatternS2(x, y int, pt int, patTable []byte, colorTable []byte) {
	var mask byte
	var b byte
	var color byte
	for i := 0; i < 8; i++ {
		if y < 64 {
			b = patTable[i+pt]
			color = colorTable[i+pt]
		} else if y < 128 {
			b = patTable[i+pt+2048]
			color = colorTable[i+pt+2048]
		} else {
			b = patTable[i+pt+2048*2]
			color = colorTable[i+pt+2048*2]
		}
		color1 := colors[(color&0xF0)>>4]
		color2 := colors[color&0x0F]
		xx := 0
		for mask = 0x80; mask > 0; mask >>= 1 {
			if mask&b != 0 {
				gogame.DrawPixel(x+xx, y+i, color1)
			} else {
				gogame.DrawPixel(x+xx, y+i, color2)
			}
			xx++
		}
	}
}

func graphics_drawSprites() {
	// Sprite name table: 1B00H to 1B7FH
	// Sprite pattern table: 3800H to 3FFFH
	sprTable := vdp_VRAM[(uint16(vdp_registers[5]) << 7):]
	sprPatTable := vdp_VRAM[(uint16(vdp_registers[6]) << 11):]
	magnif := (vdp_registers[1] & 0x01) != 0
	spr16x16 := (vdp_registers[1] & 0x02) != 0
	for i, j := 0, 0; i < 32; i, j = i+1, j+4 {
		ypos := int(sprTable[j])
		xpos := int(sprTable[j+1])
		patn := sprTable[j+2]
		ec := (sprTable[j+3] & 0x80) != 0
		color := colors[sprTable[j+3]&0x0F]
		if !spr16x16 {
			patt := sprPatTable[uint16(patn)*8:]
			drawSpr(magnif, xpos, ypos, patt, ec, color)
		} else {
			patt := sprPatTable[uint16((patn>>2))*8*4:]
			drawSpr(magnif, xpos, ypos, patt, ec, color)
			drawSpr(magnif, xpos, ypos+8, patt[8:], ec, color)
			drawSpr(magnif, xpos+8, ypos, patt[16:], ec, color)
			drawSpr(magnif, xpos+8, ypos+8, patt[24:], ec, color)
		}
	}
}

// TODO: sprite magnification not implemented
func drawSpr(magnif bool, xpos, ypos int, patt []byte, ec bool, color *gogame.Color) {
	for y := 0; y < 8; y++ {
		b := patt[y]
		for x, mask := 0, byte(0x80); mask > 0; mask >>= 1 {
			if mask&b != 0 {
				gogame.DrawPixel(xpos+x, ypos+y, color)
			}
			x++
		}
	}
}
