package main

import (
	"golang.org/x/exp/slices"
)

type A struct {
	Name string
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
}
