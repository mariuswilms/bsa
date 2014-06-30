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
	// Used for autocompletion.
	commands = []string{
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
	hf     = "/tmp/.bsa_history"
	conn   *beanstalk.Conn // Our one and only beanstalkd connection.
	line   *liner.State
	cTubes Tubes
	sigc   chan os.Signal // Signal channel.
)

// Prints help and usage.
func help() {
	fmt.Printf(`
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

	if f, err := os.Create(hf); err == nil {
		line.WriteHistory(f)
		f.Close()
	}
	line.Close()
}

func main() {
	var err error
	if conn, err = beanstalk.Dial("tcp", "127.0.0.1:11300"); err != nil {
		fmt.Println("Fatal: failed to connect to beanstalkd server.")
		os.Exit(1)
	}
	cTubes.UseAll()

	// Register signal handler.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			fmt.Printf("Caught %v. Bye.\n", sig)
			cleanup()
			os.Exit(1)
		}
	}()

	//
	line = liner.NewLiner()

	// Autocomplete commands, tube names and states.
	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		if strings.HasPrefix(line, "use") {
			tns, _ := conn.ListTubes()
			for _, v := range tns {
				c = append(c, fmt.Sprintf("%s%s", line, v))
			}
		}
		if strings.HasPrefix(line, "clear") || strings.HasPrefix(line, "next") {
			for _, v := range []string{"ready", "delayed", "buried"} {
				c = append(c, fmt.Sprintf("%s%s", line, v))
			}
		}
		return c
	})

	// Load console history if possible.
	if f, err := os.Open(hf); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	fmt.Print("Enter 'help' for available commands and 'exit' to quit.\n\n")

	// Dispatch commands.
	for {
		// We may have a new set of selected tubes after an iteration, update prompt.
		// Show selected tubes in prompt, so that we know what commands operate on.

		var tStatus string
		if cTubes.All {
			tStatus = "*"
		} else {
			tStatus = strings.Join(cTubes.Names, ", ")
		}
		prompt := fmt.Sprintf("beanstalkd [%s] > ", tStatus)

		if input, err := line.Prompt(prompt); err == nil {
			// Always add input to history, even if it contains a syntax error. We
			// may want to skip back and correct ourselves.
			line.AppendHistory(input)

			args := strings.Split(input, " ")

			switch args[0] {
			case "exit", "quit":
				cleanup()
				os.Exit(0)
			case "help":
				help()
			case "stats":
				stats()
			case "use":
				if len(args) < 2 || args[1] == "*" {
					cTubes.UseAll()
					continue
				}
				cTubes.Use(args[1:])
			case "list":
				listTubes()
			case "pause":
				if len(args) < 2 {
					fmt.Printf("Error: no delay given.\n")
					continue
				}
				if r, err := strconv.ParseUint(args[1], 0, 0); err == nil {
					pauseTubes(time.Duration(r) * time.Second)
					continue
				}
				fmt.Printf("Error: given delay is not a valid number.\n")
			case "kick":
				if len(args) < 2 {
					fmt.Printf("Error: no bound given.\n")
					continue
				}
				if r, err := strconv.ParseUint(args[1], 0, 0); err == nil {
					kickTubes(int(r))
					continue
				}
				fmt.Printf("Error: given bound is not a valid number.\n")
			case "clear":
				if len(args) < 2 {
					fmt.Printf("Error: no state given.\n")
					continue
				}
				clearTubes(args[1])
			case "next":
				if len(args) < 2 {
					fmt.Printf("Error: no state given.\n")
					continue
				}
				nextJobs(args[1])
			case "inspect":
				if len(args) < 2 {
					fmt.Printf("Error: no job id given.\n")
					continue
				}
				if r, err := strconv.ParseUint(args[1], 0, 0); err == nil {
					inspectJob(uint64(r))
					continue
				}
				fmt.Printf("Error: not a valid job id.\n")
			case "":
				continue
			default:
				fmt.Println("Error: unknown command.")
				continue
			}
		}
	}
}
