package main

import (
	"flag"
	"strconv"
	"strings"
	"time"
)

func Substr(s string, start, end int) string {
	if end < 0 {
		end = len(s) + end
	}
	return s[start:end]
}

func getParams() (int, time.Month, []string) {
	// コマンドライン
	var ym string
	var daysStrings string
	flag.StringVar(&ym, "ym", "", "")
	flag.StringVar(&daysStrings, "days", "", "")
	flag.Parse()

	y, _ := strconv.Atoi(Substr(ym, 0, 4))
	m, _ := strconv.Atoi(Substr(ym, 4, 6))
	days := strings.Split(daysStrings, ",")

	return y, time.Month(m), days
}

func removeTarget(array []string, target string) []string {
	for i := 0; i < len(array); i++ {
		if array[i] == target {
			array = append(array[:i], array[i+1:]...)
			break
		}
	}

	return array
}
