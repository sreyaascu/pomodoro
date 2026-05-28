package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Command int

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

func renderUI(label string, minute int, second int, status string, cyc int) {
	fmt.Println("====================================")
	fmt.Println("             POMODORO              |")
	fmt.Println("====================================")
	if label == Work {
		fmt.Printf("Session     :\033[31m%s\033[m                  |\n", label)
	} else if label == Break {
		fmt.Printf("Session     :\033[32m%s\033[m                 |\n", label)
	}
	fmt.Printf("Time        :%02d:%02d                 |\n", minute, second)
	if len(strconv.Itoa(cyc)) == 1 {
		fmt.Printf("Cycles      :%d                     |\n", cyc)
	} else if len(strconv.Itoa(cyc)) == 2 {
		fmt.Printf("Cycles:     :%d                    |\n", cyc)
	} else if len(strconv.Itoa(cyc)) == 3 {
		fmt.Printf("Cycles:     :%d                   |\n", cyc)
	}
	if status == "PAUSED" {
		fmt.Printf("State       :\033[33m%s\033[m                |\n", status)
	} else {
		fmt.Printf("State       :\033[36m%s\033[m               |\n", status)
	}
	fmt.Println("====================================")
	fmt.Print("Command (p=pause, r=resume, s=stop): ")
}

func runSession(label string, dur int, cmdChan chan Command, cyc int) bool {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	paused := false
	status := "RUNNING"
	for {
		select {
		case pipe := <-cmdChan:
			switch pipe {
			case Pause:
				paused = true
				status = "PAUSED"
				// fmt.Println("\nTimer Paused")
			case Resume:
				// fmt.Println("\nResuming...")
				status = "RUNNING"
				paused = false
			case Stop:
				// fmt.Println("\nExiting...")
				status = "STOPPED"
				return true
			}

		case <-ticker.C:
			clearScreen()
			minute, second := formatTime(dur)
			renderUI(label, minute, second, status, cyc)
			if paused {
				continue
			} else {
				if dur >= 0 && dur < 5 {
					go playBeep()
				}
				dur--
				if dur < 0 {
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
