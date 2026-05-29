package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Command int
type AppState struct {
	Session  string
	Status   string
	Cycles   int
	TimeLeft int
	Paused   bool
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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// Helper function
func formatTime(secs int) (int, int) {
	//Return min first and then seconds
	return secs / 60, secs % 60
}

func playBeep() {
	exec.Command(
		"paplay",
		"/usr/share/sounds/freedesktop/stereo/dialog-warning.oga",
	).Run()
}

func renderUI(state *AppState) {
	fmt.Println("====================================")
	fmt.Println("             POMODORO              |")
	fmt.Println("====================================")
	if state.Session == Work {
		fmt.Printf("Session     :\033[31m%s\033[m                  |\n", state.Session)
	} else if state.Session == Break {
		fmt.Printf("Session     :\033[32m%s\033[m                 |\n", state.Session)
	}
	minute, second := formatTime(state.TimeLeft)
	fmt.Printf("Time        :%02d:%02d                 |\n", minute, second)
	if len(strconv.Itoa(state.Cycles)) == 1 {
		fmt.Printf("Cycles      :%d                     |\n", state.Cycles)
	} else if len(strconv.Itoa(state.Cycles)) == 2 {
		fmt.Printf("Cycles:     :%d                    |\n", state.Cycles)
	} else if len(strconv.Itoa(state.Cycles)) == 3 {
		fmt.Printf("Cycles:     :%d                   |\n", state.Cycles)
	}
	if state.Status == "PAUSED" {
		fmt.Printf("State       :\033[33m%s\033[m                |\n", state.Status)
	} else {
		fmt.Printf("State       :\033[36m%s\033[m               |\n", state.Status)
	}
	fmt.Println("====================================")
	fmt.Print("Command (p=pause, r=resume, s=stop): ")
}

func runSession(label string, dur int, cmdChan chan Command, cyc int) bool {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	// paused := false
	// status := "RUNNING"
	state := AppState{
		Session:  label,
		Status:   "RUNNING",
		Cycles:   cyc,
		TimeLeft: dur,
		Paused:   false,
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
			var inp string
			for {
				fmt.Scanln(&inp)
				switch inp {
				case "p":
					cmdChan <- Pause
				case "r":
					cmdChan <- Resume
				case "s":
					cmdChan <- Stop
				}
			}
		}()
		runPomodoro(wrk_time, brk_time, cmdChan)
	}
}
