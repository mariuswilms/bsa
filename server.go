package main

func stats() {
	stats, _ := conn.Stats()
	printStats(stats, nil)
}
