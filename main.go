package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"github.com/peterh/liner"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

var (
	commands = []string{
		"help",
		"exit",
		"quit",
		"list-tubes",
		"inspect-tube",
		"kick-tube",
		"clear-tube",
		"kick-job",
		"inspect-job",
	}
	historyFile = "/tmp/.bsa_history"
	conn        *beanstalk.Conn
	line        *liner.State
	sigc        chan os.Signal
)

func help() {
	fmt.Printf(`
stats
	Shows server statistics. 

list-tubes
	List all tubes.

inspect-tube <tube>
	Inspects a tube.

kick-tube <tube> [<bound=10>]
	Kicks all jobs in given tube.

clear-tube <tube> [<state=buried>]
	Delete all jobs in given state and tube.

inspect-job <job>
	Inspects a job.

kick-job <job>
	Kicks a single job. 

help
	Show help.

exit, quit
	Exit the console.

`)
}

func cleanup() {
	conn.Close()

	if f, err := os.Create(historyFile); err == nil {
		line.WriteHistory(f)
		f.Close()
	}
	line.Close()
}

func main() {
	fmt.Print("Enter 'help' for available commands and 'quit' to exit.\n\n")
	conn, _ = beanstalk.Dial("tcp", "127.0.0.1:11300")
	line = liner.NewLiner()
	sigc = make(chan os.Signal, 1)

	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v.", sig)

			cleanup()
			os.Exit(1)
		}
	}()

	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		for _, cmd := range []string{"inspect-tube", "kick-tube", "clear-tube"} {
			if strings.HasPrefix(line, cmd) {
				tubes, _ := conn.ListTubes()

				for _, t := range tubes {
					c = append(c, fmt.Sprintf("%s %s", cmd, t))
				}
			}
		}
		return c
	})

	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	for {
		if input, err := line.Prompt("beanstalkd > "); err == nil {
			switch {
			case input == "exit", input == "quit":
				cleanup()
				os.Exit(0)
			case input == "help":
				help()
			case input == "stats":
				stats()
			case input == "list-tubes":
				listTubes()
			case strings.HasPrefix(input, "inspect-tube"):
				parts := strings.Split(input, " ")
				if len(parts) < 2 {
					fmt.Printf("Error: no tube name given.\n")
					continue
				}
				inspectTube(parts[1])
			case strings.HasPrefix(input, "kick-tube"):
				var tube string = "default"
				var bound int = 10

				parts := strings.Split(input, " ")
				if len(parts) < 3 {
					fmt.Printf("Error: no tube name or bound given.\n")
					continue
				}
				if len(parts) > 1 {
					tube = parts[1]
				}
				if len(parts) > 2 {
					b, _ := strconv.ParseInt(parts[2], 0, 0)
					bound = int(b)
				}
				kickTube(tube, bound)
			case strings.HasPrefix(input, "clear-tube"):
				var tube string = "default"
				var state string = "buried"

				parts := strings.Split(input, " ")
				if len(parts) < 2 { // state is optional
					fmt.Printf("Error: no tube name.\n")
					continue
				}
				if len(parts) > 1 {
					tube = parts[1]
				}
				if len(parts) > 2 {
					state = parts[2]
				}
				clearTube(tube, state)
			case strings.HasPrefix(input, "inspect-job"):
				var id uint64

				parts := strings.Split(input, " ")
				if len(parts) < 2 {
					fmt.Printf("Error: no job id given.\n")
					continue
				}
				if len(parts) > 1 {
					r, _ := strconv.ParseInt(parts[1], 0, 0)
					id = uint64(r)
				}
				inspectJob(id)
			}
			line.AppendHistory(input)
		}
	}
}
