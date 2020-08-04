package main

import (
	"os"
	"time"

	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	mem, err := ioutil.ReadFile("./sprites.rom")
	if err != nil {
		fmt.Printf("can't read file: %v\n", err)
	}

	f, err := os.Create("instr.dump")
	if err != nil {
		panic(err)
	}

	mf, err := os.Create("mem.dump")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	// for i, m := range mem {
	// 	if i%5 == 0 {
	// 		fmt.Println()
	// 	}
	// 	fmt.Printf("%08b\n", m)
	// }

	mem = append(mem, make([]byte, 512-len(mem))...)

	rom, err := ioutil.ReadFile("./pong.rom")
	if err != nil {
		fmt.Printf("can't read file: %v\n", err)
	}

	mem = append(mem, rom...)

	const spriteBytes = 5

	regs := [0xF + 1]byte{}
	i := uint16(0)
	stack := []uint16{}
	pc := uint16(0x200)
	screen := [64][32]bool{}
	delay := byte(0)
	// bbw := 64
	// bbh := 32
	// termbox.Size()

	// for i := 0; i < len(rom); i = i + 2 {
	// 	fmt.Printf("%04X\n", (uint16(rom[i])<<8)|uint16(rom[i+1]))
	// }

	for i := 0; i < len(mem); i = i + 2 {
		s := fmt.Sprintf("%04X = %04X\n", i, (uint16(mem[i])<<8)|uint16(mem[i+1]))
		if _, err = mf.WriteString(s); err != nil {
			panic(err)
		}
	}

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

	byteChan := make(chan byte)
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
					byteChan <- byte(0x0)
				case '1':
					byteChan <- byte(0x1)
				case '2':
					byteChan <- byte(0x2)
				case '3':
					byteChan <- byte(0x3)
				case '4':
					byteChan <- byte(0x4)
				case '5':
					byteChan <- byte(0x5)
				case '6':
					byteChan <- byte(0x6)
				case '7':
					byteChan <- byte(0x7)
				case '8':
					byteChan <- byte(0x8)
				case '9':
					byteChan <- byte(0x9)
				case 'A', 'a':
					byteChan <- byte(0xA)
				case 'B', 'b':
					byteChan <- byte(0xB)
				case 'C', 'c':
					byteChan <- byte(0xC)
				case 'D', 'd':
					byteChan <- byte(0xD)
				case 'E', 'e':
					byteChan <- byte(0xE)
				case 'F', 'f':
					byteChan <- byte(0xF)

				}
			}
		}
	}()

	keys := [0xF + 1]bool{}

	go func() {
		for {
			select {
			case b := <-byteChan:
				keys[b] = true
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
			//fmt.Printf("%04X = %v\n", opCode, mem[pc]>>4)
			s := fmt.Sprintf("%04X", opCode)
			_ = s

			if _, err = f.WriteString(s + "\n"); err != nil {
				panic(err)
			}

			//time.Sleep(time.Millisecond * 100)

			performed := false
			switch opCode >> 12 {
			case 0x0:
				if x != 0x0 {
					unknowOpcode(opCode, pc)
				}
				switch mem[pc+1] {
				case 0xE0:
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
					performed = true
				case 0xEE:
					performed = true
					if len(stack) == 0 {
						return
					}
					pc = stack[len(stack)-1]
					stack = stack[:len(stack)-1]
				default:
					unknowOpcode(opCode, pc)
				}
			case 0x1:
				performed = true
				pc = addr
				continue
			case 0x2:
				performed = true
				stack = append(stack, pc)
				pc = addr
				continue
			case 0x3:
				if regs[x] == kk {
					pc = pc + 2
				}
				performed = true
			case 0x4:
				if regs[x] != kk {
					pc = pc + 2
				}
				performed = true
			case 0x5:
				if regs[x] == regs[y] {
					pc = pc + 2
				}
				performed = true
			case 0x6:
				regs[x] = kk
				performed = true
			case 0x7:
				regs[x] = regs[x] + kk
				performed = true
			case 0x8:
				switch n {
				case 0x0:
					regs[x] = regs[y]
					performed = true
				case 0x1:
					regs[x] = regs[x] | regs[y]
					performed = true
				case 0x2:
					regs[x] = regs[x] & regs[y]
					performed = true
				case 0x3:
					regs[x] = regs[x] ^ regs[y]
					performed = true
				case 0x4:
					sum := int16(regs[x]) + int16(regs[y])
					regs[x] = byte(sum)
					regs[0xF] = byte(sum >> 8)
					performed = true
				case 0x5:
					diff := regs[x] - regs[y]
					regs[0xF] = ((^regs[x] & regs[y]) | (^(regs[x] ^ regs[y]) & diff)) >> 7
					regs[x] = diff
					performed = true
				default:
					unknowOpcode(opCode, pc)
				}
			case 0x9:
				if regs[x] != regs[y] {
					pc = pc + 2
				}
				performed = true
			case 0xA:
				i = uint16(opCode & 0xFFF)
				performed = true
			case 0xB:
				pc = uint16(regs[0x0]) + addr
				performed = true
			case 0xC:
				performed = true
				regs[x] = byte(rand.Int()>>24) & kk
			case 0xD:
				screen, regs[0xF-1] = draw(regs[x], regs[y], n, mem[i:(i+uint16(n))], screen)
				printScreen(screen)
				performed = true
			case 0xE:
				switch kk {
				case 0x9E:
				case 0xA1:
					if !keys[regs[x]] {
						pc = pc + 2
					}
					performed = true
				default:
					unknowOpcode(opCode, pc)
				}
			case 0xF:
				switch kk {
				case 0x07:
					regs[x] = delay
					performed = true
				case 0x0A:
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
					performed = true
				case 0x18:
					//TODO: Sound
					performed = true
				case 0x1E:
				case 0x29:
					//The sprites live at addr 0 and consist of spriteBytes bytes
					i = uint16(x) * spriteBytes
					performed = true
				case 0x33:
					bcd := bcd(regs[x])
					mem[i] = bcd[0]
					mem[i+1] = bcd[1]
					mem[i+2] = bcd[2]
					performed = true
				case 0x55:
					for a := uint16(0); a <= uint16(x); a++ {
						mem[i+a] = regs[a]
					}
					performed = true
				case 0x65:
					for a := uint16(0); a <= uint16(x); a++ {
						regs[a] = mem[i+a]
					}
					performed = true
				}
			}
			if !performed {
				unknowOpcode(opCode, pc)
			}
			pc = pc + 2
			time.Sleep(time.Second / 60)
		}
	}()

	<-done
}

func bcd(b byte) [3]byte {
	var bcd [3]byte
	bcd[0] = b / 100
	b -= b / 100 * 100
	bcd[1] = b / 10
	b -= b / 10 * 10
	bcd[2] = b
	return bcd
}

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

func unknowOpcode(opCode uint16, pc uint16) {
	panic(fmt.Sprintf("%04X at %04X is not a valid opcode", opCode, pc))
}

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
