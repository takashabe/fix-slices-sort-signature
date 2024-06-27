package main

import (
	"slices"
	"time"
)

type A struct {
	Name string
	Ts   time.Time
}

func main() {
	as := []A{
		{"a"},
		{"b"},
		{"c"},
	}

	slices.SortFunc(as, func(a, b A) bool {
		return a.Name < b.Name
	})

	slices.SortStableFunc(as, func(a, b A) bool {
		return a.Name > b.Name
	})

	// slices.SortFunc(as, func(a, b A) bool {
	//   return a.Ts.Before(b.Ts)
	// })
	//
	// slices.SortStableFunc(as, func(a, b A) bool {
	//   return a.Ts.After(b.Ts)
	// })
}
