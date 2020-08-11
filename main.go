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


type Msg int
const (
    Next Msg = iota
    Prev
    Paus
    Tick
    Loud
)

type CmdMsg int
const (
    SayStretch CmdMsg = iota
    SayTimeRemaining
    SayAllDone
)


type Model struct{
    isPaused bool
    sayExercises bool
    stretchIDX int
    currentTicks int
    ticksPerStretch int
    stretches []string
    sideEffects chan int
}

func tick(m Model) Model {
    if (m.isPaused) {
        return m
    }
    if m.stretchIDX > len(m.stretches) {
        return m
    }
    m.currentTicks += 1
    if m.currentTicks < m.ticksPerStretch {
        m.sideEffects <- int(SayTimeRemaining)
    }

    if m.currentTicks >= m.ticksPerStretch {
        m = nextStretch(m)
    }
    return m
}


func nextStretch(m Model) Model {
    if m.stretchIDX  < len(m.stretches) - 1 {
        m.stretchIDX += 1
        m.currentTicks = 0
        m.sideEffects <- int(SayStretch)
    }
    return m
}

func lastStretch(m Model) Model {
    if m.stretchIDX > 0 {
        m.stretchIDX -= 1
        m.sideEffects <- int(SayStretch)
    }
    m.currentTicks = 0

    return m
}

// func finished(m Model) bool
func displayInstructions(s tcell.Screen) {
    emitStr(s, 2, 2, tcell.StyleDefault, "Press f/b to go to next/previous stretches")
    emitStr(s, 2, 3, tcell.StyleDefault, "Press p to toggle pause")
    emitStr(s, 2, 4, tcell.StyleDefault, "Press ESC exit")
    return
}

func display(m Model, s tcell.Screen) {
    // TODO check if model is different
    s.Clear()
    displayInstructions(s)
    if  m.stretchIDX == len(m.stretches) - 1  && m.currentTicks >= m.ticksPerStretch {
        whatToSay := "All done!"
        emitStr(s, 2, 1, tcell.StyleDefault, whatToSay);
    } else {
        whatToSay := fmt.Sprintf("%s", m.stretches[m.stretchIDX])
        emitStr(s, 2, 1, tcell.StyleDefault, whatToSay);
    }

    if m.isPaused {
        emitStr(s, 2, 10, tcell.StyleDefault, "Paused")
    }

    if !m.sayExercises {
        emitStr(s, 2, 11, tcell.StyleDefault, "Quiet")
    }

    emitStr(s, 2, 9, tcell.StyleDefault, fmt.Sprintf("%d s", m.ticksPerStretch - m.currentTicks))
    s.Show()
}

func tick_once_per_second(m Model, s tcell.Screen, ui chan int) {
    lastTime :=  time.Now()
    go func() {
        for {
            now := time.Now()
            if now.Sub(lastTime) > 1000*time.Millisecond {
                lastTime = lastTime.Add(time.Duration(1000*time.Millisecond))
                ui <- int(Tick)
            }
            display(m, s)
            time.Sleep(100*time.Millisecond)
        }
    }()
    go func() {
        for msg := range m.sideEffects {
            switch unwrappedMsg := CmdMsg(msg); unwrappedMsg {
            case SayStretch:
                // TODO need to lock this with mutexes?
                if m.sayExercises && m.stretchIDX < len(m.stretches) {
                    cmd := exec.Command("say", m.stretches[m.stretchIDX])
                    cmd.Run()
                }
            case SayTimeRemaining:
                ticksLeft := m.ticksPerStretch - m.currentTicks
                if m.sayExercises &&  ticksLeft > 0 && ticksLeft <= 3 {
                    cmd := exec.Command("say", fmt.Sprintf("%d", m.ticksPerStretch - m.currentTicks))
                    cmd.Run()
                }
                if m.sayExercises &&  ticksLeft == 30 {
                    cmd := exec.Command("say", "start")
                    cmd.Run()
                }
                if m.sayExercises &&  ticksLeft == 15 {
                    cmd := exec.Command("say", "half way")
                    cmd.Run()
                }
            }
        }
    }()

    for msg := range ui {
        switch unwrappedMsg := Msg(msg); unwrappedMsg {
        case Next:
            m = nextStretch(m)
        case Prev:
            m = lastStretch(m)
        case Paus:
            m.isPaused = !m.isPaused
        case Loud:
            m.sayExercises = !m.sayExercises
        case Tick:
            m = tick(m)
        }
    }
    close(m.sideEffects)
}


func grabUserInput(s tcell.Screen, ui chan int) {
    for {
        ev := s.PollEvent()
        switch ev := ev.(type) {
        case *tcell.EventResize:
            s.Sync()
        case *tcell.EventKey:
            if ev.Key() == tcell.KeyEscape {
                close(ui)
            } else {
                switch ch := ev.Rune(); ch {
                case 'f':
                    ui <- int(Next)
                case 'b':
                    ui <- int(Prev)
                case 'p':
                    ui <- int(Paus)
                case 'v':
                    ui <- int(Loud)
                }
            }
        }
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
    s.Clear()

    ui := make(chan int)
    se := make(chan int)
    go grabUserInput(s, ui)

    // stretches := []string{
    //     "Left Arm Across",
    //     "Right Arm Across",
    //     "Left Arm Over and Behind Head",
    //     "Right Arm Over and Behind Head",
    //     "Left Arm Behind Back",
    //     "Right Arm Behind Back",
    //     "Neck Left",
    //     "Neck Up",
    //     "Neck Right",
    //     "Neck Down",
    //     "Left Calf Stretch",
    //     "Right Calf Stretch",
    //     "Left Calf Stretch",
    //     "Right Calf Stretch",
    //     "Touch Toes",
    //     "Wide Touch Toes",
    //     "Touch Toes",
    //     "Right Quadracep",
    //     "Left Quadracep",
    //     "Left Foot Forward Lunge",
    //     "Right Foot Forward Lunge",
    //     "Left Foot Crossed Over Right Knee",
    //     "Right Foot Crossed Over Left Knee",
    //     "Legs Diamond, Lean Forward",
    //     "Twist, Left Leg Over",
    //     "Twist, Right Leg Over",
    // }
    // m := Model{true, true, 0, 0, 33, stretches, se}

	stretches := []string{
		"Jumping Jacks",
		"Pushups",
		"Wall Sit",
		"Crunches",
		"Squats",
		"Dips",
		"Plank",
		"Lunges",
		"Pushups with Rotation",
		"Side Plank Left",
		"Side Plank Right",
	}

    m := Model{true, true, 0, 0, 38, stretches, se}
    tick_once_per_second(m, s, ui)
}
