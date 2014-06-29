package main

import (
	"fmt"
)

func inspectJob(id uint64) {
	body, _ := conn.Peek(id)
	stats, _ := conn.StatsJob(id)
	printJob(id, body, stats)
}

func buryJob(id uint64) {
	conn.Bury(id, 0)
	fmt.Printf("Buried job %v.\n", id)
}

func nextJobs(state string) {
	for _, t := range ctubes {
		fmt.Printf("Next %s job in %s:\n", state, t.Name)

		pf := func() (id uint64, body []byte, err error) {
			switch state {
			case "ready":
				return t.PeekReady()
			case "delayed":
				return t.PeekDelayed()
			case "buried":
				return t.PeekBuried()
			}
			return 0, nil, fmt.Errorf("Unknown state %s", state)
		}
		id, body, _ := pf()
		stats, _ := conn.StatsJob(id)
		printJob(id, body, stats)
		fmt.Println()
	}
}

func printJob(id uint64, body []byte, stats map[string]string) {
	fmt.Printf("%25s: %v\n", "id", id)
	fmt.Printf("%25s:\n---------------------\n%s\n---------------------\n", "body", body)

	var include = []string{
		"tube",
		"age",
		"reserves",
		"kicks",
		"delay",
		"releases",
		"pri",
		"ttr",
		"time-left",
		"timeouts",
		"buries",
	}
	printStats(stats, include)
}
