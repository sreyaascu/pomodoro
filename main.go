package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func printHelp() {
	fmt.Printf("GOMODORO\r\n")
	// fmt.Println()
	fmt.Printf("\nUsage:\r\n")
	fmt.Printf("pomo start [work] [break]\r\n")
	// fmt.Printf()
	fmt.Printf("\nCommands:\r\n")
	fmt.Printf("  start       Start a pomodoro session\r\n")
	fmt.Printf("  --help      Show help\r\n")
	fmt.Printf("  --version   Show version\r\n")
}

type Command int
type AppState struct {
	Session   string //For Break/Work
	Status    string //For Paused/Running
	Cycles    int
	TimeLeft  int
	TotalTime int
	Paused    bool
}
type Config struct {
	WorkMinutes   int  `json:"work_minutes"`
	BreakMinutes  int  `json:"break_minutes"`
	Sound         bool `json:"sound"`
	Notifications bool `json:"notifications"`
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
	var config Config
	cmdChan := make(chan Command, 1)
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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomo start <work_minutes> <break_minutes>")
		return
	}
	cmd := os.Args[1]
	switch cmd {
	case "start":
		if len(os.Args) < 4 {
			home, home_err := os.UserHomeDir()
			if home_err != nil {
				fmt.Println(home_err)
			}
			config_path := filepath.Join(
				home,
				".config",
				"gomodoro",
				"config.json",
			)
			// return
			data, read_err := os.ReadFile(config_path)
			if read_err != nil {
				fmt.Println("Error reading config file")
				return
			}
			json.Unmarshal(data, &config)
			// fmt.Println("Usage: go run . <work_time> <break_time>")
			// return
			runPomodoro(config.WorkMinutes*60, config.BreakMinutes*60, cmdChan)
			return
		} else {

			wrk_time, err1 := strconv.Atoi(os.Args[2])
			brk_time, err2 := strconv.Atoi(os.Args[3])
			wrk_time *= 60
			brk_time *= 60

			if err1 != nil || err2 != nil {
				fmt.Println("Invalid Input")
				return
			}
			runPomodoro(wrk_time, brk_time, cmdChan)
		}
	case "--help":
		printHelp()
		return
	case "--version":
		fmt.Printf("Gomodoro v1.0.0\r\n")
		return
	default:
		fmt.Println("Unknown command")
	}
}
