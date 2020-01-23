package main

// takes a CF timestamp and compares it against Now()

import (
	"fmt"
	"time"
)

const (
	against = "2020-01-17 23:39:28.673486 +0000 UTC"

	// format generated from https://golang.org/src/time/format.go
	layoutCF = "2006-01-02 15:04:05.000000 -0700 MST"
)

func main() {
	fmt.Println("against:", against)
	//convertedagainst := time.Date(against, time.UTC)
	x := time.Now().UTC()
	fmt.Println("    now:", x)
	//fmt.Println("   diff:", x.Sub())

	t, _ := time.Parse(layoutCF, against)
	fmt.Println("       :", t)
}
