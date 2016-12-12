// Copyright 2014 David Persson. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

func inspectJob(id uint64) (err error) {
	body, err := conn.Peek(id)
	stats, _ := conn.StatsJob(id)

	if err != nil {
		return fmt.Errorf("Unknown job %v", id)
	}
	printJob(id, body, stats)

	return
}

func nextJobs(state string) {
	for _, t := range cTubes.Conns {
		if id, body, err := peekState(t, state); err == nil {
			fmt.Printf("Next %s job in %s:\n", state, t.Name)

			stats, _ := conn.StatsJob(id)
			printJob(id, body, stats)
			fmt.Println()
		}
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
		"timeouts",
		"buries",
	}
	printStats(stats, include)
}
