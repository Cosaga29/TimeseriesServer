package main

import "time"

var reqCount = 0
var start = time.Now().UnixMilli()

func GetRate() float64 {
	t := time.Now().UnixMilli() - start
	if t != 0 {
		return float64(reqCount) / float64(t) * 1000
	}

	return 0.0
}
