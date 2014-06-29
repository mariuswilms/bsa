package main

import (
	"fmt"
	"github.com/kr/beanstalk"
	"sort"
	"strconv"
)

// Helper function to print statistics. Can use whitelist
// if provided. Otherwise will print all keys.
func printStats(data map[string]string, whitelist []string) {
	keys := make([]string, 0, len(data))
	for k := range data {
		if whitelist == nil || contains(k, whitelist) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for i := range keys {
		fmt.Printf("%25s: %s\n", keys[i], data[keys[i]])
	}
}

// Helper function to convert integers from
// strings as returned by stats commands.
func castStatsValue(v string) int {
	r, _ := strconv.ParseUint(v, 0, 0)
	return int(r)
}

// Helper function to check if a given string is contained in an slice of
// strings.
func contains(n string, h []string) bool {
	for _, v := range h {
		if v == n {
			return true
		}
	}
	return false
}

func peekState(t beanstalk.Tube, state string) (id uint64, body []byte, err error) {
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
