package main

import (
	"fmt"
	"sort"
	"strconv"
)

func printStats(data map[string]string, whitelist []string) {
	isIn := func(n string, h []string) bool {
		for _, v := range h {
			if v == n {
				return true
			}
		}
		return false
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		if whitelist == nil || isIn(k, whitelist) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for i := range keys {
		fmt.Printf("%25s: %s\n", keys[i], data[keys[i]])
	}
}

func castStatsValue(v string) int {
	r, _ := strconv.ParseUint(v, 0, 0)
	return int(r)
}
