// Copyright 2014 David Persson. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"strings"
	"time"
)

type Tubes struct {
	Names []string
	Conns []beanstalk.Tube
	All   bool // Flag indicating if all available tubes are represented.
}

func (ts *Tubes) Reset() {
	ts.Conns = ts.Conns[:0]
	ts.Names = ts.Names[:0]
}

// Selects tubes.
func (ts *Tubes) Use(tns []string) {
	ts.Reset()
	ts.All = false

	for _, tn := range tns {
		ts.Conns = append(ts.Conns, beanstalk.Tube{conn, tn})
		ts.Names = append(ts.Names, tn)
	}
	return
}

func (ts *Tubes) UseAll() {
	ts.Reset()
	ts.All = true

	tns, _ := conn.ListTubes()
	for _, tn := range tns {
		ts.Conns = append(ts.Conns, beanstalk.Tube{conn, tn})
		ts.Names = append(ts.Names, tn)
	}
	return
}

// Prints most important statistics for each tube.
func listTubes() {
	lf := "%20s %10s %30s %30s\n"

	fmt.Printf(lf, "", "paused", "ready/delayed/buried", "waiting/watching/using")
	fmt.Println(strings.Repeat("-", 93))

	for _, t := range cTubes.Conns {
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
	fmt.Println()
}

func kickTubes(bound int) {
	for _, t := range cTubes.Conns {
		t.Kick(bound)
		fmt.Printf("Kicked jobs in tube %s.\n", t.Name)
	}
}

func pauseTubes(delay time.Duration) {
	for _, t := range cTubes.Conns {
		t.Pause(delay)
		fmt.Printf("Paused tube %s for %v.\n", t.Name, delay)
	}
}

func clearTubes(state string) {
	cnt := 0

	for _, t := range cTubes.Conns {
		for {
			if id, _, err := peekState(t, state); err == nil {
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
