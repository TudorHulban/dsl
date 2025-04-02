package main

type Criteria struct {
	Name string

	Settings []*Setting // Settings defined within this criteria
	Monitors []*Monitor
}
