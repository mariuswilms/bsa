package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func castStatsValue(v string) int {
	r, _ := strconv.ParseUint(v, 0, 0)
	return int(r)
}

// FIXME Calculate mean values?
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
			castStatsValue(stats["current-jobs-waiting"]),
			castStatsValue(stats["current-jobs-watching"]),
			castStatsValue(stats["current-jobs-using"]),
		)
		fmt.Printf(lf, t.Name, pf, wf, jf)
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
		fmt.Printf("Tube %s cleared, %d %s jobs %s deleted.\n", t.Name, cnt, state)
	}
}
