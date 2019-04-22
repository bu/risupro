package main

import "github.com/gdamore/tcell"
import "log"
import "fmt"
import runewidth "github.com/mattn/go-runewidth"

import "strconv"

const (
	ModeNormal = iota
	ModeCommand
)

var CurrentMode int
var CurrentInputCommand string
var CurrentMessage string

func puts(s tcell.Screen, style tcell.Style, x, y int, str string) {
	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
}

func main() {
	s, err := tcell.NewScreen()

	quit := make(chan struct{})

	if err != nil {
		log.Fatal(err)
	}

	if err = s.Init(); err != nil {
		log.Fatal(err)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))

	s.Clear()

	CurrentMode = ModeNormal

	UpdateModeStatusAndCommandLine(s)
	s.Show()

	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {

				case tcell.KeyEscape:
					CurrentMode = ModeNormal
					CurrentInputCommand = ""
					UpdateModeStatusAndCommandLine(s)

				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if CurrentMode == ModeCommand {
						cmdRunes := []rune(CurrentInputCommand)
						CurrentInputCommand = string(cmdRunes[:len(cmdRunes)-1])
					}

					if CurrentInputCommand == "" {
						CurrentMode = ModeNormal
					}

					UpdateModeStatusAndCommandLine(s)

				case tcell.KeyEnter:
					if CurrentMode == ModeCommand {
						switch CurrentInputCommand {
						case ":quit", ":q":
							close(quit)
							return
						case ":redraw":
							CurrentInputCommand = ""
							CurrentMode = ModeNormal
							RedrawScreen(s)
							UpdateModeStatusAndCommandLine(s)

						default:
							CurrentMessage = "UNKNOWN"
							CurrentInputCommand = ""
							CurrentMode = ModeNormal
							UpdateModeStatusAndCommandLine(s)
						}
					}

				case tcell.KeyRune:
					if ev.Rune() == ':' {
						CurrentMode = ModeCommand
					}

					if CurrentMode == ModeCommand {
						CurrentInputCommand = CurrentInputCommand + string(ev.Rune())
						UpdateModeStatusAndCommandLine(s)
					}

				}
			case *tcell.EventResize:
				RedrawScreen(s)
				UpdateModeStatusAndCommandLine(s)
			}
		}
	}()

	<-quit

	s.Fini()
}

func UpdateModeStatusAndCommandLine(s tcell.Screen) {
	screen_x, screen_y := s.Size()

	switch CurrentMode {

	case ModeNormal:
		puts(s, tcell.StyleDefault, 1, screen_y-2, "NORMAL")
	case ModeCommand:
		puts(s, tcell.StyleDefault, 1, screen_y-2, "COMMAND")
	}

	puts(s, tcell.StyleDefault, 1, screen_y-1, fmt.Sprintf("%-"+strconv.Itoa(screen_x)+"v...", CurrentInputCommand))

	if CurrentInputCommand == "" && CurrentMessage != "" {
		puts(s, tcell.StyleDefault, 1, screen_y-1, fmt.Sprintf("%-"+strconv.Itoa(screen_x)+"v...", CurrentMessage))
		CurrentMessage = ""
	}

	s.Show()
}

func RedrawScreen(s tcell.Screen) {
	s.Clear()
	screen_x, screen_y := s.Size()

	puts(s, tcell.StyleDefault, 1, 5, strconv.Itoa(screen_x))
	puts(s, tcell.StyleDefault, 10, 5, strconv.Itoa(screen_y))

	s.Sync()
}
