package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Command int
type AppState struct {
	Session   string //For Break/Work
	Status    string //For Paused/Running
	Cycles    int
	TimeLeft  int
	TotalTime int
	Paused    bool
}

const (
	Work  = "WORK"
	Break = "BREAK"
)

const (
	Pause Command = iota
	Resume
	Stop
)

// Helper function
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func formatTime(secs int) (int, int) {
	//Return min first and then seconds
	return secs / 60, secs % 60
}

func notify(state AppState) {
	if state.Session == Break {
		exec.Command(
			"notify-send",
			"Time to get to Work",
		).Run()
	} else if state.Session == Work {
		exec.Command(
			"notify-send",
			"Break time!!!",
		).Run()

	}
}

func playBeep() {
	exec.Command(
		"paplay",
		"/usr/share/sounds/freedesktop/stereo/dialog-warning.oga",
	).Run()
}

func buildProgressBar(state *AppState) string {
	completed := state.TotalTime - state.TimeLeft
	perc := float64(completed) / float64(state.TotalTime)
	barLength := 35
	filled := int(perc * float64(barLength))
	remaining := barLength - filled
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < remaining; i++ {
		bar += "░"
	}

	return bar
}

func renderUI(state *AppState) {
	fmt.Print("====================================\r\n")
	fmt.Print("             GOMODORO              |\r\n")
	fmt.Print("====================================\r\n")
	if state.Session == Work {
		fmt.Printf("Session     :\033[31m%s\033[m                  |\r\n", state.Session)
	} else if state.Session == Break {
		fmt.Printf("Session     :\033[32m%s\033[m                 |\r\n", state.Session)
	}
	minute, second := formatTime(state.TimeLeft)
	fmt.Printf("Time        :%02d:%02d                 |\r\n", minute, second)
	if len(strconv.Itoa(state.Cycles)) == 1 {
		fmt.Printf("Cycles      :%d                     |\r\n", state.Cycles)
	} else if len(strconv.Itoa(state.Cycles)) == 2 {
		fmt.Printf("Cycles:     :%d                    |\r\n", state.Cycles)
	} else if len(strconv.Itoa(state.Cycles)) == 3 {
		fmt.Printf("Cycles:     :%d                   |\r\n", state.Cycles)
	}
	if state.Status == "PAUSED" {
		fmt.Printf("State       :\033[33m%s\033[m                |\r\n", state.Status)
	} else {
		fmt.Printf("State       :\033[36m%s\033[m               |\r\n", state.Status)
	}
	fmt.Printf("%s|\r\n", buildProgressBar(state))
	fmt.Print("====================================\r\n")
	// fmt.Println("Command (p=pause, r=resume, s=stop): ")
}

func runSession(label string, dur int, cmdChan chan Command, cyc int) bool {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	// paused := false
	// status := "RUNNING"
	state := AppState{
		Session:   label,
		Status:    "RUNNING",
		Cycles:    cyc,
		TimeLeft:  dur,
		TotalTime: dur,
		Paused:    false,
	}
	for {
		select {
		case pipe := <-cmdChan:
			switch pipe {
			case Pause:
				// paused = true
				// status = "PAUSED"
				state.Paused = true
				state.Status = "PAUSED"
			case Resume:
				state.Status = "RUNNING"
				state.Paused = false
			case Stop:
				// fmt.Println("\nExiting...")
				state.Status = "STOPPED"
				return true
			}

		case <-ticker.C:
			clearScreen()
			// renderUI(label, minute, second, status, cyc)
			renderUI(&state) //testing
			if state.Paused {
				continue
			} else {
				if state.TimeLeft >= 0 && state.TimeLeft < 5 {
					go playBeep()
				}
				state.TimeLeft--
				if state.TimeLeft < 0 {
					notify(state)
					return false
				}
			}

		}
	}
}

func runPomodoro(wrk int, brk int, cmdChan chan Command) {
	cyc := 0
	for {
		if runSession(Work, wrk, cmdChan, cyc) {
			return
		}
		if runSession(Break, brk, cmdChan, cyc) {
			return
		}
		cyc++
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <work_time> <break_time>")
		return
	}
	cmdChan := make(chan Command, 1)
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	cmd := os.Args[1]
	switch cmd {
	case "start":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run . <work_time> <break_time>")
			return
		}

		wrk_time, err1 := strconv.Atoi(os.Args[2])
		brk_time, err2 := strconv.Atoi(os.Args[3])
		wrk_time *= 60
		brk_time *= 60

		if err1 != nil || err2 != nil {
			fmt.Println("Invalid Input")
			return
		}
		go func() {
			buff := make([]byte, 1)
			for {
				os.Stdin.Read(buff)
				switch buff[0] {
				case 'p':
					cmdChan <- Pause
				case 'r':
					cmdChan <- Resume
				case 's':
					cmdChan <- Stop
				}
			}
		}()
		runPomodoro(wrk_time, brk_time, cmdChan)
	}
}
