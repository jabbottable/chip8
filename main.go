package main

import (
	"flag"
	"time"

	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/nsf/termbox-go"
)

// mem is the memory for CHIP-8.
//
// 0x000 to 0x1FF is reserved for the interpreter and
// should not be modified by any program.
// Program may refer to a group of sprites that represent
// the hexadecimal digits 0 - F. These sprites are 5
// bytes long, or 8x5 pixels. eg 0
// +------+----------+------+
// | "0"  | Binary   | Hex  |
// +------+----------+------+
// | **** | 11110000 | 0xF0 |
// | *  * | 10010000 | 0x90 |
// | *  * | 10010000 | 0x90 |
// | *  * | 10010000 | 0x90 |
// | **** | 11110000 | 0xF0 |
// +------+----------+------+
var mem []byte = []byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func main() {

	fileName := flag.String("f", "", "file path of rom")
	flag.Parse()

	if *fileName == "" {
		fmt.Println(
			" CHIP-8 is tool for running CHIP-8 roms.\n\n",
			"Usage:\n\n",
			"\tchip8 -f [file]\n\n",
			"Run in a 64 character per line terminal.",
		)
		return
	}

	mem = append(mem, make([]byte, 512-len(mem))...)

	rom, err := ioutil.ReadFile(*fileName)
	if err != nil {
		fmt.Printf("Can't read file: %v\n", err)
		return
	}

	mem = append(mem, rom...)

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	const spriteBytes = 5

	regs := [0xF + 1]byte{}
	i := uint16(0)
	stack := []uint16{}
	pc := uint16(0x200)
	screen := [64][32]bool{}
	delay := byte(0)

	opCode := uint16(0)
	addr := uint16(0)
	n := byte(0)
	x := byte(0)
	y := byte(0)
	kk := byte(0)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	backbuf := make([]termbox.Cell, w*h)
	copy(termbox.CellBuffer(), backbuf)

	keyboardEvents := make(chan byte)
	done := make(chan bool)

	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyEsc {
					done <- true
				}
				switch ev.Ch {
				case '0':
					keyboardEvents <- byte(0x0)
				case '1':
					keyboardEvents <- byte(0x1)
				case '2':
					keyboardEvents <- byte(0x2)
				case '3':
					keyboardEvents <- byte(0x3)
				case '4':
					keyboardEvents <- byte(0x4)
				case '5':
					keyboardEvents <- byte(0x5)
				case '6':
					keyboardEvents <- byte(0x6)
				case '7':
					keyboardEvents <- byte(0x7)
				case '8':
					keyboardEvents <- byte(0x8)
				case '9':
					keyboardEvents <- byte(0x9)
				case 'A', 'a':
					keyboardEvents <- byte(0xA)
				case 'B', 'b':
					keyboardEvents <- byte(0xB)
				case 'C', 'c':
					keyboardEvents <- byte(0xC)
				case 'D', 'd':
					keyboardEvents <- byte(0xD)
				case 'E', 'e':
					keyboardEvents <- byte(0xE)
				case 'F', 'f':
					keyboardEvents <- byte(0xF)

				}
			}
		}
	}()

	keys := [0xF + 1]bool{}

	go func() {
		for {
			select {
			case k := <-keyboardEvents:
				keys[k] = true
			case <-time.After(time.Second):
				keys = [0xF + 1]bool{}
			}
		}
	}()

	go func() {
		for {
			opCode = (uint16(mem[pc]) << 8) | uint16(mem[pc+1])
			addr = opCode & 0x0FFF
			n = byte(opCode & 0xF)
			x = byte(opCode >> 8 & 0xF)
			y = byte(opCode >> 4 & 0xF)
			kk = byte(opCode & 0xFF)

			switch opCode >> 12 {
			case 0x0:
				if x != 0x0 {
					unknowOpcode(opCode, pc)
				}
				switch mem[pc+1] {
				case 0xE0:
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				case 0xEE:
					if len(stack) == 0 {
						return
					}
					pc = stack[len(stack)-1]
					stack = stack[:len(stack)-1]
				default:
					unknowOpcode(opCode, pc)
				}
			case 0x1:
				pc = addr
				continue
			case 0x2:
				stack = append(stack, pc)
				pc = addr
				continue
			case 0x3:
				if regs[x] == kk {
					pc = pc + 2
				}
			case 0x4:
				if regs[x] != kk {
					pc = pc + 2
				}
			case 0x5:
				if regs[x] == regs[y] {
					pc = pc + 2
				}
			case 0x6:
				regs[x] = kk
			case 0x7:
				regs[x] = regs[x] + kk
			case 0x8:
				switch n {
				case 0x0:
					regs[x] = regs[y]
				case 0x1:
					regs[x] = regs[x] | regs[y]
				case 0x2:
					regs[x] = regs[x] & regs[y]
				case 0x3:
					regs[x] = regs[x] ^ regs[y]
				case 0x4:
					sum := int16(regs[x]) + int16(regs[y])
					regs[x] = byte(sum)
					regs[0xF] = byte(sum >> 8)
				case 0x5:
					diff := regs[x] - regs[y]
					regs[0xF] = ((^regs[x] & regs[y]) | (^(regs[x] ^ regs[y]) & diff)) >> 7
					regs[x] = diff
				default:
					unknowOpcode(opCode, pc)
				}
			case 0x9:
				if regs[x] != regs[y] {
					pc = pc + 2
				}
			case 0xA:
				i = uint16(opCode & 0xFFF)
			case 0xB:
				pc = uint16(regs[0x0]) + addr
			case 0xC:
				regs[x] = byte(rand.Int()>>24) & kk
			case 0xD:
				screen, regs[0xF-1] = draw(regs[x], regs[y], n, mem[i:(i+uint16(n))], screen)
				printScreen(screen)
			case 0xE:
				switch kk {
				case 0x9E:
					unimplementedOpcode(opCode, pc)
				case 0xA1:
					if !keys[regs[x]] {
						pc = pc + 2
					}
				default:
					unknowOpcode(opCode, pc)
				}
			case 0xF:
				switch kk {
				case 0x07:
					regs[x] = delay
				case 0x0A:
					unimplementedOpcode(opCode, pc)
				case 0x15:
					delay = x
					delayTicker := time.NewTicker(time.Second / 60)
					go func() {
						for {
							select {
							case <-delayTicker.C:
								delay--
								if delay == 0 {
									return
								}
							}
						}
					}()
				case 0x18:
					//TODO: Sound
				case 0x1E:
					unimplementedOpcode(opCode, pc)
				case 0x29:
					//The sprites live at addr 0 and consist of spriteBytes bytes
					i = uint16(x) * spriteBytes
				case 0x33:
					bcd := bcd(regs[x])
					mem[i] = bcd[0]
					mem[i+1] = bcd[1]
					mem[i+2] = bcd[2]
				case 0x55:
					for a := uint16(0); a <= uint16(x); a++ {
						mem[i+a] = regs[a]
					}
				case 0x65:
					for a := uint16(0); a <= uint16(x); a++ {
						regs[a] = mem[i+a]
					}
				}
			}
			pc = pc + 2
			time.Sleep(time.Second / 60)
		}
	}()

	<-done
}

// bcd returns a binary-coded decimal representaion of b.
//
// For a given decimal bytes value return a 3 length byte
// array with the hundreds digit in index 0, tens digit
// in index 1 and the ones digit in index 2.
//
// eg if b is 147 the return would be [3]byte{1, 4, 7}
func bcd(b byte) [3]byte {
	var bcd [3]byte
	bcd[0] = b / 100
	b -= b / 100 * 100
	bcd[1] = b / 10
	b -= b / 10 * 10
	bcd[2] = b
	return bcd
}

// draw s sprite at xStart,yStart with width of 8 pixels
// and a height of n pixels on screen. Sprite pixels are
// XOR'd with screen pixels. Draw returns the updated
// screen and 1 if any screen pixel was unset when the
// sprite was drawn otherwise 0.
//
// The screen is a 64x32 pixel monochrome display with
// the following co-ordinate system:
// +-------------------+
// |(0,0)        (63,0)|
// |                   |
// |(0,31)      (63,31)|
// +-------------------+
func draw(xStart, yStart, n byte, s []byte, screen [64][32]bool) ([64][32]bool, byte) {
	f := false
	for y := yStart; y < yStart+n; y++ {
		for x := xStart; x < xStart+8; x++ {
			p := (s[y-yStart] << (x - xStart) & 0x80) == 0x80
			f = f || (screen[x%64][y%32] && p)
			screen[x%64][y%32] = (p || screen[x%64][y%32]) && !(p && screen[x%64][y%32])
		}
	}
	if f {
		return screen, byte(1)
	}
	return screen, byte(0)
}

// prinScreen writes the content of the screen to the termbox CellBuffer.
// Every two screen rows are one termbox row. Three braille characters and
// empty space are used to give the four options two rows in a column can
// be - both on, both off or either one on.
func printScreen(screen [64][32]bool) {
	backbuf := []termbox.Cell{}
	for i := 0; i < 32; i = i + 2 {
		for j := 0; j < 64; j++ {
			if screen[j][i] {
				if screen[j][i+1] {
					backbuf = append(backbuf, termbox.Cell{Ch: '⣿'})
				} else {
					backbuf = append(backbuf, termbox.Cell{Ch: '⠛'})
				}
			} else if screen[j][i+1] {
				backbuf = append(backbuf, termbox.Cell{Ch: '⣤'})
			} else {
				backbuf = append(backbuf, termbox.Cell{Ch: ' '})
			}
		}
	}
	copy(termbox.CellBuffer(), backbuf)
	termbox.Flush()
}

func unknowOpcode(opCode uint16, pc uint16) {
	panic(fmt.Sprintf("%04X at %04X is not a valid opcode", opCode, pc))
}

func unimplementedOpcode(opCode uint16, pc uint16) {
	panic(fmt.Sprintf("%04X at %04X is an unimplemented opcode", opCode, pc))
}
