// Copyright 2014 David Persson. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func stats() {
	stats, _ := conn.Stats()
	printStats(stats, nil)
}
