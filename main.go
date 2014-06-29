package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"github.com/peterh/liner"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var (
	commands = []string{
		"bury",
		"clear",
		"help",
		"inspect",
		"exit",
		"quit",
		"kick",
		"list",
		"next",
		"pause",
		"stats",
		"use",
	}
	historyFile = "/tmp/.bsa_history"
	conn        *beanstalk.Conn
	line        *liner.State
	sigc        chan os.Signal
	ctubes      []beanstalk.Tube // The currently selected tubes.
)

func help() {
	fmt.Printf(`

bury <job>
	Buries a single job.

clear <state>
	Deletes all jobs in given state and selected tubes.
	<state> may be either 'ready', 'buried' or 'delayed'.

help
	Show this wonderful help.

exit, 
quit
	Exit the console.

inspect <job>
	Inspects a single job.

pause <delay>
	Pauses selected tubes for given number of seconds.

kick <bound>
	Kicks all jobs in selected tubes.

list
	Lists all selected tubes or if none is selected all exstings tubes 
	and shows status of each.

next <state> 
	Inspects next jobs in given state in selected tubes.
	<state> may be either 'ready', 'buried' or 'delayed'.

stats
	Shows server statistics. 

use [<tube0>] [<tube1> ...]
	Selects one or multiple tubes. Separate multiple tubes by spaces.
	If no tube name is given resets selection.

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
	fmt.Print("Enter 'help' for available commands and 'exit' to quit.\n\n")
	conn, _ = beanstalk.Dial("tcp", "127.0.0.1:11300")
	line = liner.NewLiner()
	sigc = make(chan os.Signal, 1)

	// Register signal handler.
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			fmt.Printf("Caught %v. Bye.\n", sig)
			cleanup()
			os.Exit(1)
		}
	}()

	// Autocomplete commands, tube names and states.
	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		if strings.HasPrefix(line, "use") {
			tubes, _ := conn.ListTubes()
			for _, t := range tubes {
				c = append(c, fmt.Sprintf("%s%s", line, t))
			}
		}
		if strings.HasPrefix(line, "clear") || strings.HasPrefix(line, "next") {
			states := []string{"ready", "delayed", "buried"}
			for _, s := range states {
				c = append(c, fmt.Sprintf("%s %s", "clear", s))
			}
		}
		return c
	})

	// Load console history.
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	// Dispatch commands.
	for {
		// We may have a new set of selected tubes after an iteration, update prompt.
		// Show selected tubes in prompt, so that we know what commands operate on.
		var names []string
		for _, t := range ctubes {
			names = append(names, t.Name)
		}
		prompt := fmt.Sprintf("beanstalkd [%s] > ", strings.Join(names, ", "))

		if input, err := line.Prompt(prompt); err == nil {
			parts := strings.Split(input, " ")

			switch parts[0] {
			case "exit", "quit":
				cleanup()
				os.Exit(0)
			case "help":
				help()
			case "stats":
				stats()
			case "use":
				ctubes = ctubes[:0]

				if len(parts) < 2 {
					continue // Just reset.
				}
				for _, n := range parts[1:] {
					ctubes = append(ctubes, beanstalk.Tube{conn, n})
				}
			case "list":
				if len(ctubes) == 1 {
					// Temporarily select all tubes.
					tubes, _ := conn.ListTubes()
					for _, n := range tubes {
						ctubes = append(ctubes, beanstalk.Tube{conn, n})
					}
				}
				listTubes()

				// Revert temporary selection back again.
				ctubes = ctubes[:0]
			case "pause":
				if len(parts) < 2 {
					fmt.Printf("Error: no delay given.\n")
					continue
				}
				r, _ := strconv.ParseUint(parts[1], 0, 0)
				pauseTubes(time.Duration(r) * time.Second)
			case "kick":
				if len(parts) < 2 {
					fmt.Printf("Error: no bound given.\n")
					continue
				}
				r, _ := strconv.ParseInt(parts[1], 0, 0)
				kickTubes(int(r))
			case "clear":
				if len(parts) < 2 {
					fmt.Printf("Error: no state given.\n")
					continue
				}
				clearTubes(parts[1])
			case "next":
				if len(parts) < 2 {
					fmt.Printf("Error: no state given.\n")
					continue
				}
				nextJobs(parts[1])
			case "inspect":
				if len(parts) < 2 {
					fmt.Printf("Error: no job id given.\n")
					continue
				}
				r, _ := strconv.ParseInt(parts[1], 0, 0)
				inspectJob(uint64(r))
			case "bury":
				if len(parts) < 2 {
					fmt.Printf("Error: no job id given.\n")
					continue
				}
				r, _ := strconv.ParseInt(parts[1], 0, 0)
				buryJob(uint64(r))
			default:
				fmt.Println("Error: unknown command.")
				continue
			}
			line.AppendHistory(input)
		}
	}
}
