package main

import (
	"testing"
)

func Test_draw(t *testing.T) {
	xStart := byte(0)
	yStart := byte(0)
	n := byte(1)
	s := []byte{0xFF}
	screen := [64][32]bool{}
	screen, over := draw(xStart, yStart, n, s, screen)

	m := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}

	if over == 1 {
		t.Error("over shouldn't be set")
	}

	for i := 0; i < 32; i = i + 2 {
		for j := 0; j < 64; j++ {
			v := screen[j][i]
			if _, ok := m[j]; ok {
				if i == 0 && !v {
					t.Errorf("j:%v i:%v should be true", j, i)
				} else if i != 0 && v {
					t.Errorf("j:%v i:%v should be false", j, i)
				}
			} else if v {
				t.Errorf("j:%v i:%v should be false", j, i)
			}
		}
	}

	screen, over = draw(xStart, yStart, n, s, screen)
	if over != 1 {
		t.Error("over should be set")
	}

	for i := 0; i < 32; i = i + 2 {
		for j := 0; j < 64; j++ {
			if screen[j][i] {
				t.Errorf("j:%v i:%v should be false", j, i)
			}
		}
	}
}
