package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"strings"
	"time"
)

// Selects tubes and validates each tube name before doing so.
func useTubes(tns []string) (err error) {
	ctubes = ctubes[:0]
	atns, _ := conn.ListTubes()

	for _, tn := range tns {
		if !contains(tn, atns) {
			return fmt.Errorf("Invalid tube %s", tn)
		}
		ctubes = append(ctubes, beanstalk.Tube{conn, tn})
	}
	return
}

// Prints most important statistics for each tube.
func listTubes() {
	lf := "%20s %10s %30s %30s\n"

	fmt.Printf(lf, "", "paused", "ready/delayed/buried", "waiting/watching/using")
	fmt.Println(strings.Repeat("-", 93))

	for _, t := range ctubes {
		var pf, wf, jf string
		stats, _ := t.Stats()

		if stats["pause"] == "0" {
			pf = "-"
		} else {
			pf = fmt.Sprintf("%ss", stats["pause-time-left"])
		}
		jf = fmt.Sprintf(
			"%d (%d) / %d / %d",
			castStatsValue(stats["current-jobs-ready"]),
			castStatsValue(stats["current-jobs-urgent"]),
			castStatsValue(stats["current-jobs-delayed"]),
			castStatsValue(stats["current-jobs-buried"]),
		)
		wf = fmt.Sprintf(
			"%d / %d / %d",
			castStatsValue(stats["current-waiting"]),
			castStatsValue(stats["current-watching"]),
			castStatsValue(stats["current-using"]),
		)
		fmt.Printf(lf, t.Name, pf, jf, wf)
	}
}

func kickTubes(bound int) {
	for _, t := range ctubes {
		t.Kick(bound)
		fmt.Printf("Kicked jobs in tube %s.\n", t.Name)
	}
}

func pauseTubes(delay time.Duration) {
	for _, t := range ctubes {
		t.Pause(delay)
		fmt.Printf("Paused tube %s for %v.\n", t.Name, delay)
	}
}

func clearTubes(state string) {
	cnt := 0

	for _, t := range ctubes {
		pf := func(state string) (id uint64, body []byte, err error) {
			switch state {
			case "ready":
				return t.PeekReady()
			case "delayed":
				return t.PeekDelayed()
			case "buried":
				return t.PeekBuried()
			}
			return
		}
		for {
			if id, _, err := pf(state); err == nil {
				if err := conn.Delete(id); err != nil {
					panic(fmt.Sprintf("Failed deleting job %v.\n", id))
				}
				cnt++
			} else {
				break
			}
		}
		fmt.Printf("Tube %s cleared, %d %s jobs deleted.\n", t.Name, cnt, state)
	}
}
