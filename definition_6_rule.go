package main

type Rule struct {
	Level     int
	Condition Expression // the 'when' condition expression
}
