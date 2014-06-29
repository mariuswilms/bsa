package main

import (
	"fmt"
)

func stats() {
	stats, _ := conn.Stats()

	fmt.Print("== Stats\n")
	printStats(stats, nil)
}
