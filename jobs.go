package main

import (
	"fmt"
)

func inspectJob(id uint64) {
	body, _ := conn.Peek(id)
	stats, _ := conn.StatsJob(id)

	fmt.Printf("== Job %v", id)
	printJob(id, body, stats)
}

func printJob(id uint64, body []byte, stats map[string]string) {
	fmt.Printf("%25s: %v\n", "id", id)
	fmt.Printf("%25s:\n---------------------\n%s\n---------------------\n", "body", body)

	var include = []string{
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
