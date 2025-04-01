package main

type setting struct {
	kind  string     // type of setting, e.g., "baseline", "increment"
	name  string     // name of the setting variable
	value expression // the value assigned (can be a literal or another expression)
}
