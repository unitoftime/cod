package main

import (
	"fmt"
	"time"
)

func printDuration(name string, t time.Time) {
	fmt.Printf("%s: %v\n", name, time.Since(t))
}

var enableDebug = false

func debugPrintf(format string, a ...any) {
	if enableDebug {
		fmt.Printf(format, a...)
	}
}

func debugPrintln(a ...any) {
	if enableDebug {
		fmt.Println(a...)
	}
}
