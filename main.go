package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"

	"github.com/mattn/go-runewidth"

)

var defStyle tcell.Style

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

// const (
// 	RuneFilled = '█'
// 	RunEmptyCircle = '○'
// 	RunePipeV = '║'
// 	RunePipeUR = '╗'
// 	RunePipeLR = '╝'
// 	RunePipeUL = '╔'
// 	RunePipeLL = '╚'
// 	RunePipeH = '═'
// )



type Model struct{
	isPaused bool
	stretchIDX int
	currentTicks int
	ticksPerStretch int
	stretches []string
}

func tick(m Model, s tcell.Screen) Model {
	if (m.isPaused) {
		return m
	}
	if m.stretchIDX > len(m.stretches) {
		return m
	}
	m.currentTicks += 1
	if m.currentTicks >= m.ticksPerStretch {
		m = nextStretch(m, s)
		// m.stretchIDX += 1
		// m.currentTicks = 0
	}
	return m
}


func nextStretch(m Model, s tcell.Screen) Model {
	if len(m.stretches) > m.stretchIDX - 1 {
		s.Clear()
		m.stretchIDX += 1
		m.currentTicks = 0
		whatToSay := fmt.Sprintf("%s", m.stretches[m.stretchIDX])
		cmd := exec.Command("say", whatToSay)
		emitStr(s, 2, 1, tcell.StyleDefault, whatToSay)
		cmd.Run()
	} else {
		s.Clear()
		emitStr(s, 2, 1, tcell.StyleDefault, "All done!")
		cmd := exec.Command("say", "All done")
		cmd.Run()
	}
	return m
}

func lastStretch(m Model, s tcell.Screen) Model {
	if m.stretchIDX > 0 {
		m.stretchIDX -= 1
	}
	s.Clear()
	m.currentTicks = 0
	whatToSay := fmt.Sprintf("%s", m.stretches[m.stretchIDX])
	cmd := exec.Command("say", whatToSay)
	emitStr(s, 2, 1, tcell.StyleDefault, whatToSay)
	cmd.Run()
	return m
}

func displayInstructions(s tcell.Screen) {
	emitStr(s, 2, 2, tcell.StyleDefault, "Press f/b to go to next/previous stretches")
	emitStr(s, 2, 3, tcell.StyleDefault, "Press p to toggle pause")
	emitStr(s, 2, 4, tcell.StyleDefault, "Press ESC exit")
	return
}

func tick_once_per_second(m Model, s tcell.Screen) {
	lastTime :=  time.Now()

	// w, h := s.Size()
	// white := tcell.StyleDefault.
		// Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	// emitStr(s, 2, 5, white, "Press ESC exit")
	displayInstructions(s)
	emitStr(s, 2, 1, tcell.StyleDefault, "Press p to start")
	s.Show()
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			// s.SetContent(w-1, h-1, 'R', nil, st)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				// cmd := exec.Command("say", "Good bye")
				// cmd.Run()
				return
			} else {
				switch ch := ev.Rune(); ch {
				case 'f':
					m = nextStretch(m, s)
					// m.stretchIDX += 1
				case 'b':
					m = lastStretch(m, s)
					// m.stretchIDX -= 1
				case 'p':
					//                                    "Paused"
					emitStr(s, 2, 10, tcell.StyleDefault, "      ")

					m.isPaused = !m.isPaused
				}

				// ev.Rune() == 'n' {
				// }
				// if ev.Rune() == 'C' || ev.Rune() == 'c' {
				// 	s.Clear()
				// }
			}
		}
		now := time.Now()
		if now.Sub(lastTime).Milliseconds() > 1000 {
			m = tick(m, s)
			lastTime = lastTime.Add(time.Duration(1000*Time.Milliseconds))
		}
		if m.isPaused {
			emitStr(s, 2, 10, tcell.StyleDefault, "Paused")
		}
		emitStr(s, 2, 9, tcell.StyleDefault, fmt.Sprintf("%d s", m.ticksPerStretch - m.currentTicks))

		s.Show()
	}
}
func main() {

	encoding.Register()

	s, e := tcell.NewScreen()
	defer s.Fini()

	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	defStyle = tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)
	// s.EnableMouse()
	s.Clear()

	stretches := []string{
		"Left Arm Across",
		"Right Arm Across",
		"Left Arm Over and Behind Head",
		"Right Arm Over and Behind Head",
		"Left Arm Behind Back",
		"Right Arm Behind Back",
		"Neck Left",
		"Neck Up",
		"Neck Right",
		"Neck Down",
		"Left Calf Stretch",
		"Right Calf Stretch",
		"Left Calf Stretch",
		"Right Calf Stretch",
		"Touch Toes",
		"Wide Touch Toes",
		"Touch Toes",
		"Right Quadracep",
		"Left Quadracep",
		"Left Foot Forward Lunge",
		"Right Foot Forward Lunge",
		"Left Foot Crossed Over Right Knee",
		"Right Foot Crossed Over Left Knee",
		"Legs Diamond, Lean Forward",
		"Twist, Left Leg Over",
		"Twist, Right Leg Over",
	}
	m := Model{true, 0, 0, 33, stretches}
	tick_once_per_second(m, s)
    // Println outputs a line to stdout.
    // It comes from the package fmt.
    // fmt.Printf("%c\n", RuneFilled)
}
