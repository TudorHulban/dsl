package main

type Setting struct {
	Kind string // e.g., "baseline", "increment"
	Name string

	Value Expression // the value assigned
}
