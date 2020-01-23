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

	difference := x.Sub(t)

	fmt.Println("difference:", difference)
	fmt.Println("-seconds---------")
	x = x.Truncate(time.Second)
	t = t.Truncate(time.Second)
	fmt.Println("   trunc now:", x)
	fmt.Println("trunc parsed:", t)
	fmt.Println("        diff:", x.Sub(t))

	fmt.Println("-minutes---------")
	x = x.Truncate(time.Minute)
	t = t.Truncate(time.Minute)
	fmt.Println("   trunc now:", x)
	fmt.Println("trunc parsed:", t)
	fmt.Println("        diff:", x.Sub(t))

	check5start := time.Now().UTC()
	check5end := check5start
	check5end = check5end.Add(time.Minute * 5)

	fmt.Println("-5 min difference---------")
	fmt.Println("start:", check5start)
	fmt.Println("  end:", check5end)

	if check5start == check5end {
		fmt.Println("same time")
	} else {
		fmt.Println("different time")
		fmt.Println("difference=", check5end.Sub(check5start).Round(time.Second).Seconds())
	}
}
