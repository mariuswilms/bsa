package main

import (
	"fmt"
	"github.com/kr/beanstalk"
)

func listTubes() {
	tubes, _ := conn.ListTubes()
	for _, v := range tubes {
		fmt.Printf("%s\n", v)
	}
}

func inspectTube(name string) {
	tube := beanstalk.Tube{conn, name}
	stats, _ := tube.Stats()

	fmt.Print("General:\n")
	printStats(stats, []string{
		"pause",
		"pause-time-left",
	})
	fmt.Print("Jobs:\n")
	printStats(stats, []string{
		"total-jobs",
	})
	fmt.Print("Workers:\n")
	printStats(stats, []string{
		"current-waiting",
		"current-watching",
		"current-using",
	})
	fmt.Print("\n")

	printTubeJobSection("ready", stats)
	printTubeJobSection("delayed", stats)
	printTubeJobSection("buried", stats)
}

func kickTube(name string, bound int) {
	tube := beanstalk.Tube{conn, name}
	tube.Kick(bound)

	fmt.Printf("Kicked jobs in tube %s.\n", name)
}

func clearTube(name string, state string) {
	tube := beanstalk.Tube{conn, name}
	cnt := 0

	pf := func(state string) (id uint64, body []byte, err error) {
		switch state {
		case "ready":
			return tube.PeekReady()
		case "delayed":
			return tube.PeekDelayed()
		case "buried":
			return tube.PeekBuried()
		}
		return
	}

	for {
		if id, _, err := pf(state); err == nil {
			if err := conn.Delete(id); err != nil {
				panic(fmt.Sprintf("Failed deleting job %v\n", id))
			}
			cnt++
		} else {
			break
		}
	}
	fmt.Printf("Tube %s cleared, %d %s jobs %s deleted.\n", name, cnt, state)
}

func printTubeJobSection(t string, tubeStats map[string]string) {
	var id uint64
	var body []byte
	var err error

	switch t {
	case "ready":
		id, body, err = conn.Tube.PeekReady()
	case "delayed":
		id, body, err = conn.Tube.PeekDelayed()
	case "buried":
		id, body, err = conn.Tube.PeekBuried()
	}
	if err != nil {
		return
	}
	stats, _ := conn.StatsJob(id)

	fmt.Printf("\n-- Next %s job\n", t)
	printJob(id, body, stats)
}
